package horizon

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/logger"
)

const (
	// HTTP client
	defaultTimeout = time.Second * 5
	defaultRetries = 3

	// Horizon API
	loginPath   = "/rest/login"
	logoutPath  = "/rest/logout"
	refreshPath = "/rest/refresh"
	eventsPath  = "/rest/external/v1/audit-events"
)

var errTokenExpired = errors.New("refresh token expired")

// Client gets events from the configured Horizon API REST server. Remote()
// returns the address of the Horizon API REST server.
type Client interface {
	GetEvents(ctx context.Context, since Timestamp) ([]AuditEventSummary, error)
	Remote() string
}

type horizonClient struct {
	client      *resty.Client
	credentials AuthLoginRequest
	tokens      AuthTokens
	logger      logger.Logger
}

var _ Client = (*horizonClient)(nil)

func newHorizonClient(ctx context.Context, server string, credentials AuthLoginRequest, insecure bool, log logger.Logger) (*horizonClient, error) {
	rc := newRESTClient(server, insecure, log)
	c := horizonClient{
		client:      rc,
		logger:      log,
		credentials: credentials,
	}

	if insecure {
		c.logger.Warnw("using potentially insecure connection to Horizon API server", "address", server, "insecure", insecure)
	}

	c.logger.Debug("authenticating against Horizon API")
	if err := c.login(ctx); err != nil {
		return nil, errors.Wrap(err, "horizon API login")
	}

	return &c, nil
}

func newRESTClient(server string, insecure bool, log logger.Logger) *resty.Client {
	// REST global client defaults
	r := resty.New().SetLogger(log)
	r.SetBaseURL(server)
	r.SetHeader("content-type", "application/json")
	r.SetAuthScheme("Bearer")
	r.SetRetryCount(defaultRetries).SetRetryMaxWaitTime(defaultTimeout)

	if insecure {
		r.SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: true,
		})
	}

	return r
}

// login performs an authentication request to the Horizon API server, sets and
// stores the returned auth and refresh tokens
func (h *horizonClient) login(ctx context.Context) error {
	/* Access tokens would be valid for 30 minutes while the refresh token would be
	valid for 8 hours. Once the access token has expired, the user will get a 401
	response from the APIs and would need to get a new access token from the refresh
	endpoint using the refresh token. If the Refresh token is also expired (after 8
	hours and when user gets a 400), it indicates that user needs to fully
	re-authenticate using login endpoint due to invalid refresh token.
	*/

	// check if we can use an existing refresh token
	if h.tokens.RefreshToken != "" {
		err := h.refresh(ctx)

		// success
		if err == nil {
			return nil
		}

		if !errors.Is(err, errTokenExpired) {
			return errors.Wrap(err, "refresh token")
		}
	}

	// perform full login
	res, err := h.client.R().SetContext(ctx).SetBody(h.credentials).Post(loginPath)
	if err != nil {
		return err
	}

	if !res.IsSuccess() {
		return fmt.Errorf("horizon API login returned non-success status code: %d", res.StatusCode())
	}

	var tokens AuthTokens
	err = json.Unmarshal(res.Body(), &tokens)
	if err != nil {
		return errors.Wrap(err, "unmarshal JSON authentication token response")
	}

	h.tokens = tokens
	h.client.SetAuthToken(h.tokens.AccessToken)
	h.logger.Debug("Horizon API login successful")

	return nil
}

// refresh attempts to refresh an expired auth token. If the refresh token has
// expired, errTokenExpired will be returned.
func (h *horizonClient) refresh(ctx context.Context) error {
	request := RefreshTokenRequest{h.tokens.RefreshToken}
	res, err := h.client.R().SetContext(ctx).SetBody(request).Post(refreshPath)
	if err != nil {
		return err
	}

	if !res.IsSuccess() {
		switch res.StatusCode() {
		case http.StatusBadRequest:
			return errTokenExpired

		default:
			return fmt.Errorf("unexpected HTTP response: %d %s", res.StatusCode(), string(res.Body()))
		}
	}

	var accessToken AccessToken
	err = json.Unmarshal(res.Body(), &accessToken)
	if err != nil {
		return errors.Wrap(err, "unmarshal JSON access token response")
	}

	token := accessToken.AccessToken
	h.tokens.AccessToken = token
	h.client.SetAuthToken(token)
	h.logger.Debug("auth token refresh successful")
	return nil
}

// GetEvents returns a list of AuditEventSummary from the Horizon API
func (h *horizonClient) GetEvents(ctx context.Context, since Timestamp) ([]AuditEventSummary, error) {
	var (
		res     *resty.Response
		retries int
		err     error

		timeRange string
		params    map[string]string
	)

	// handle auth expired cases
	for retries < 2 {
		if since == 0 {
			// return last (up to) 10 initial events if no timestamp is specified
			params = map[string]string{
				"size": "10",
				"page": "1",
			}
		} else {
			timeRange, err = timeRangeFilter(since, 0)
			h.logger.Debugw("using time range filter", "filter", timeRange)
			if err != nil {
				return nil, errors.Wrap(err, "create time range query filter")
			}

			params = map[string]string{
				"filter": timeRange,
			}
		}

		res, err = h.client.R().SetContext(ctx).SetQueryParams(params).Get(eventsPath)
		if err != nil {
			return nil, err
		}

		h.logger.Debugf("request: %+v", *res.Request)
		h.logger.Debugf("response headers: %+v", res.Header())
		h.logger.Debugf("response body: %s", string(res.Body()))

		if !res.IsSuccess() {
			switch res.StatusCode() {
			// perform re-auth
			case http.StatusUnauthorized:
				if err = h.login(ctx); err != nil {
					h.logger.Error(string(res.Body()))
					return nil, errors.Wrapf(err, "not authenticated: %s", string(res.Body()))
				}
				h.logger.Debug("retrying get events after re-authentication")
				retries++
				continue

			// 	conflict (note: should never happen on GET and incorrectly used in spec for
			// 	DB missing error)
			case http.StatusConflict:
				h.logger.Error(string(res.Body()))
				return nil, errors.New("HTTP conflict error: 401 (DB not initialized?)")

			// 	not defined in spec
			default:
				h.logger.Error(string(res.Body()))
				return nil, fmt.Errorf("unexpected status code: %d %s", res.StatusCode(), string(res.Body()))
			}
		}

		var events []AuditEventSummary
		err = json.Unmarshal(res.Body(), &events)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshal JSON audit events response")
		}

		return events, nil
	}

	return nil, fmt.Errorf("get events status code: %d %s", res.StatusCode(), string(res.Body()))
}

// timeRangeFilter returns the JSON-encoded query string for the given timestamp
// range. Both values are interpreted as inclusive range values. If to is 0 an
// arbitrary time (UTC) in the future is used as the upper range bound.
func timeRangeFilter(from, to Timestamp) (string, error) {
	// avoid small clock sync issues between client and server and use 1d as future
	// timestamp buffer
	timeBuffer := time.Hour * 24

	if to == 0 {
		to = Timestamp(time.Now().Add(timeBuffer).Unix() * 1000) // milliseconds
	}

	f := BetweenFilter{
		Type:      "Between",
		Name:      "time",
		FromValue: from,
		ToValue:   to,
	}

	filter, err := json.Marshal(f)
	if err != nil {
		return "", errors.Wrap(err, "marshal filter")
	}

	return string(filter), nil
}

// logout performs a logout against the Horizon API. No error is returned if the
// underlying token has already expired or the server returned a non-successful
// HTTP response.
func (h *horizonClient) logout(ctx context.Context) error {
	request := RefreshTokenRequest{h.tokens.RefreshToken}
	res, err := h.client.R().SetContext(ctx).SetBody(request).Post(logoutPath)
	if err != nil {
		return err
	}

	if !res.IsSuccess() {
		switch res.StatusCode() {
		case http.StatusBadRequest:
			h.logger.Warn("auth token already expired")

		default:
			h.logger.Errorf("unexpected status code: %d %s", res.StatusCode(), string(res.Body()))
		}
	}

	return nil
}

// Remote returns the remote server address the client is connected to
func (h *horizonClient) Remote() string {
	return h.client.BaseURL
}
