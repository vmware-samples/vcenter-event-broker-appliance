package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/eventbridge/eventbridgeiface"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/pkg/errors"
	"github.com/timbray/quamina"
	"go.uber.org/zap"

	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/logger"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor"
)

const (
	defaultPageLimit = 50 // max 50 results per page for list operations
)

// TODO(@mgasch): allow for multiple event rules for configured bus
type matcher struct {
	*quamina.Quamina
	bus     string // uses cfg.eventbus as pattern name
	pattern string
}

// EventBridgeProcessor implements the Processor interface
type EventBridgeProcessor struct {
	session session.Session
	eventbridgeiface.EventBridgeAPI
	matcher matcher

	// options
	logger.Logger

	mu    sync.RWMutex
	stats metrics.EventStats
}

// assert we implement Processor interface
var _ processor.Processor = (*EventBridgeProcessor)(nil)

// NewEventBridgeProcessor returns an AWS EventBridge processor for the given
// configuration
func NewEventBridgeProcessor(ctx context.Context, cfg *config.ProcessorConfigEventBridge, ms metrics.Receiver, log logger.Logger, opts ...Option) (*EventBridgeProcessor, error) {
	// Initialize awsSession for the AWS SDK client
	var awsSession *session.Session

	awsLog := log
	if zapSugared, ok := log.(*zap.SugaredLogger); ok {
		proc := strings.ToUpper(string(config.ProcessorEventBridge))
		awsLog = zapSugared.Named(fmt.Sprintf("[%s]", proc))
	}

	proc := EventBridgeProcessor{
		Logger: awsLog,
	}

	// apply options
	for _, opt := range opts {
		opt(&proc)
	}

	if cfg == nil {
		return nil, errors.New("no AWS EventBridge configuration found")
	}

	if cfg.Region == "" {
		return nil, errors.New("region must be specified")
	}

	if cfg.RuleARN == "" {
		return nil, errors.New("rule ARN must be specified")
	}

	if cfg.EventBus == "" {
		return nil, errors.New("event bus must be specified")
	}

	if proc.EventBridgeAPI == nil {
		// Check the Auth Method to determine how the Session should be established
		if cfg.Auth.Type == config.AWSAccessKeyAuth {
			if cfg.Auth == nil || cfg.Auth.AWSAccessKeyAuth == nil {
				return nil, fmt.Errorf("invalid %s credentials: accessKey and secretKey must be set", config.AWSAccessKeyAuth)
			}
			accessKey := cfg.Auth.AWSAccessKeyAuth.AccessKey
			secretKey := cfg.Auth.AWSAccessKeyAuth.SecretKey

			awsSessionAccessKey, err := session.NewSession(&aws.Config{
				Region: aws.String(cfg.Region),
				Credentials: credentials.NewStaticCredentials(
					accessKey,
					secretKey,
					"", // a token will be created when the session is used.
				),
			})
			if err != nil {
				return nil, errors.Wrap(err, "create AWS session")
			}
			// Set the AWS Session to the IAM Role authenticated session
			awsSession = awsSessionAccessKey
		}
		if cfg.Auth.Type == config.AWSIAMRoleAuth {
			// Create Session without additional options will load credentials region,
			// and profile loaded from the environment and shared config automatically
			awsSessionIam, err := session.NewSession(&aws.Config{
				Region: aws.String(cfg.Region),
			})
			if err != nil {
				return nil, errors.Wrap(err, "create AWS session")
			}
			// Set the AWS Session to the IAM Role authenticated session
			awsSession = awsSessionIam
		}

		proc.session = *awsSession
		ebSession := eventbridge.New(awsSession)

		if ebSession == nil {
			return nil, errors.New("create AWS event bridge session")
		}

		proc.EventBridgeAPI = ebSession
	}

	if err := configureRuleMatcher(ctx, &proc, cfg.EventBus, cfg.RuleARN); err != nil {
		return nil, errors.Wrap(err, "configure rule matcher")
	}

	// pre-populate the metrics stats
	proc.stats = metrics.EventStats{
		Provider:    string(config.ProcessorEventBridge),
		Type:        config.EventProcessor,
		Address:     cfg.RuleARN, // Using Rule ARN to uniquely identify and represent this processor
		Started:     time.Now().UTC(),
		Invocations: make(map[string]*metrics.InvocationDetails),
	}

	go proc.PushMetrics(ctx, ms)

	return &proc, nil
}

func configureRuleMatcher(ctx context.Context, proc *EventBridgeProcessor, bus string, ruleARN string) error {
	q, err := quamina.New()
	if err != nil {
		return errors.Wrap(err, "create quamina pattern match instance")
	}

	proc.matcher = matcher{
		Quamina: q,
		bus:     bus,
	}

	var (
		found     bool
		nextToken *string
	)

	for !found {
		rules, err := proc.ListRulesWithContext(ctx, &eventbridge.ListRulesInput{
			EventBusName: aws.String(bus),             // explicitly passing eventbus name because list assumes "default" otherwise
			Limit:        aws.Int64(defaultPageLimit), // up to n results per page for requests.
			NextToken:    nextToken,
		})
		if err != nil {
			return errors.Wrap(err, "list event bridge rules")
		}

	arnLoop:
		for _, rule := range rules.Rules {
			switch {
			case *rule.Arn == ruleARN:
				if rule.EventPattern == nil {
					return errors.New("rule event pattern must not be nil")
				}

				pattern := *rule.EventPattern
				if err := proc.matcher.AddPattern(proc.matcher.bus, pattern); err != nil {
					return errors.Wrap(err, "add rule event pattern to matcher")
				}

				proc.Infow(
					"adding rule event forwarding pattern to processor",
					"bus",
					proc.matcher.bus,
					"pattern",
					pattern,
				)
				proc.matcher.pattern = pattern

				found = true
				break arnLoop

			default:
				continue
			}
		}

		switch {
		case found:
			break
		case rules.NextToken != nil: // try next batch of rules, if any
			nextToken = rules.NextToken
			continue
		default: // nothing found
			return errors.Errorf("rule %q not found for configured AWS event bridge account", ruleARN)
		}
	}
	return nil
}

// Process implements the stream processor interface
func (eb *EventBridgeProcessor) Process(ctx context.Context, ce cloudevents.Event) error {
	eb.Debugw("processing event", "eventID", ce.ID(), "event", ce.String())

	subject := ce.Subject()
	eb.mu.Lock()
	// initialize invocation stats
	if _, ok := eb.stats.Invocations[subject]; !ok {
		eb.stats.Invocations[subject] = &metrics.InvocationDetails{}
	}
	eb.mu.Unlock()

	// this is the format eventbridge uses (also for matching) so we need to wrap
	// the cloudevent into the Detail field in order to correctly match
	awsEvent := struct {
		Detail cloudevents.Event `json:"detail,omitempty"`
	}{
		Detail: ce,
	}

	e, err := json.Marshal(awsEvent)
	if err != nil {
		return processor.NewError(config.ProcessorEventBridge, errors.Wrapf(err, "convert cloudevent to aws eventbridge event: %s", ce.ID()))
	}

	matches, err := eb.matcher.MatchesForEvent(e)
	if err != nil {
		return processor.NewError(config.ProcessorEventBridge, errors.Wrapf(err, "match event: %s", ce.ID()))
	}

	for _, m := range matches {
		if m == eb.matcher.bus {
			jsonBytes, err := json.Marshal(ce)
			if err != nil {
				return processor.NewError(config.ProcessorEventBridge, errors.Wrapf(err, "marshal event: %s", ce.ID()))
			}

			jsonString := string(jsonBytes)
			entry := eventbridge.PutEventsRequestEntry{
				Detail:       aws.String(jsonString),
				EventBusName: aws.String(eb.matcher.bus),
				Source:       aws.String(ce.Source()),
				DetailType:   aws.String(subject),
			}

			// TODO: add batching (metrics stats currently assume single item)
			input := eventbridge.PutEventsInput{
				Entries: []*eventbridge.PutEventsRequestEntry{&entry},
			}

			eb.Infow("sending event", "eventID", ce.ID(), "type", ce.Type(), "subject", subject)
			resp, err := eb.PutEventsWithContext(ctx, &input)
			eb.Debugw("got response", "eventID", ce.ID(), "response", resp)

			updateStats := func(err error) error {
				eb.mu.Lock()
				defer eb.mu.Unlock()
				if err != nil {
					eb.stats.Invocations[subject].Failure()
					return processor.NewError(config.ProcessorEventBridge, errors.Wrapf(err, "send event: %s", ce.ID()))
				}

				eb.Infow("successfully sent event", "eventID", ce.ID())
				eb.stats.Invocations[subject].Success()
				return nil
			}
			return updateStats(err)
		}
	}

	eb.Debugw(
		"skipping event: pattern rule does not match",
		"eventID", ce.ID(),
		"event", ce.String(),
		"pattern", eb.matcher.pattern,
	)
	return nil
}

// PushMetrics pushes metrics to the specified metrics receiver
func (eb *EventBridgeProcessor) PushMetrics(ctx context.Context, ms metrics.Receiver) {
	ticker := time.NewTicker(metrics.PushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			eb.mu.RLock()
			ms.Receive(&eb.stats)
			eb.mu.RUnlock()
		}
	}
}

// Shutdown attempts a clean shutdown of the AWS EventBridge processor
// TODO: check if we need to perform anything here
func (eb *EventBridgeProcessor) Shutdown(_ context.Context) error {
	eb.Logger.Infof("attempting graceful shutdown") // noop for now
	return nil
}
