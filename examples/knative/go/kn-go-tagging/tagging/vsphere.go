package tagging

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/session/keepalive"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/soap"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/logging"
)

const (
	// defaultMountPath is where the vSphere secret will be saved.
	defaultMountPath = "/var/bindings/vsphere"
	// vCenter APIs keep-alive time interval.
	keepAliveInterval = 5 * time.Minute
)

// ReadKey reads the key from the secret.
func ReadKey(key, path string) (string, error) {
	mountPath := defaultMountPath
	if path != "" {
		mountPath = path
	}

	data, err := ioutil.ReadFile(filepath.Join(mountPath, key))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// newSOAPClient returns a vCenter SOAP API client with active keep-alive. Use
// Logout() to release resources and perform a clean logout from vCenter.
func newSOAPClient(ctx context.Context, env EnvConfig) (*govmomi.Client, error) {
	parsedURL, err := soap.ParseURL(env.VCAddress)
	if err != nil {
		return nil, err
	}

	// Read the username and password from the filesystem.
	username, err := ReadKey(corev1.BasicAuthUsernameKey, env.SecretPath)
	if err != nil {
		return nil, err
	}
	password, err := ReadKey(corev1.BasicAuthPasswordKey, env.SecretPath)
	if err != nil {
		return nil, err
	}
	parsedURL.User = url.UserPassword(username, password)

	soapClient := soap.NewClient(parsedURL, env.Insecure)
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, fmt.Errorf("new VIM client: %w", err)
	}

	vimClient.RoundTripper = keepalive.NewHandlerSOAP(
		vimClient.RoundTripper,
		keepAliveInterval,
		soapKeepAliveHandler(ctx, vimClient),
	)

	// Explicitly create session to activate keep-alive handler via login.
	mgr := session.NewManager(vimClient)
	err = mgr.Login(ctx, parsedURL.User)
	if err != nil {
		return nil, fmt.Errorf("SOAP session manager login: %w", err)
	}

	vSphereClient := govmomi.Client{
		Client:         vimClient,
		SessionManager: mgr,
	}

	return &vSphereClient, nil
}

func soapKeepAliveHandler(ctx context.Context, c *vim25.Client) func() error {
	logger := logging.FromContext(ctx)

	return func() error {
		logger.Debug("executing SOAP keep-alive handler")
		t, err := methods.GetCurrentTime(ctx, c)
		if err != nil {
			return err
		}

		logger.Debug("vCenter current time", "time", t.String())
		return nil
	}
}

// newRESTClient returns a vCenter REST API client with active keep-alive. Use
// Logout() to release resources and perform a clean logout from vCenter.
func newRESTClient(ctx context.Context, soapClient *vim25.Client, env EnvConfig) (*rest.Client, error) {
	parsedURL, err := soap.ParseURL(env.VCAddress)
	if err != nil {
		return nil, err
	}

	// Read the username and password from the filesystem.
	username, err := ReadKey(corev1.BasicAuthUsernameKey, env.SecretPath)
	if err != nil {
		return nil, err
	}
	password, err := ReadKey(corev1.BasicAuthPasswordKey, env.SecretPath)
	if err != nil {
		return nil, err
	}
	parsedURL.User = url.UserPassword(username, password)

	restclient := rest.NewClient(soapClient)
	restclient.Transport = keepalive.NewHandlerREST(restclient, keepAliveInterval, restKeepAliveHandler(ctx, restclient))

	// Login activates the keep-alive handler.
	if err = restclient.Login(ctx, parsedURL.User); err != nil {
		return nil, err
	}
	return restclient, nil
}

func restKeepAliveHandler(ctx context.Context, restclient *rest.Client) func() error {
	logger := logging.FromContext(ctx)

	return func() error {
		logger.Debug("executing REST keep-alive handler")
		s, err := restclient.Session(ctx)
		if err != nil {
			return err
		}
		if s != nil {
			return nil
		}
		return errors.New(http.StatusText(http.StatusUnauthorized))
	}
}
