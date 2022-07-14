package tags

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"time"

	"github.com/embano1/vsphere/logger"
	"github.com/kelseyhightower/envconfig"
	nsxt "github.com/vmware/go-vmware-nsxt"
	"go.uber.org/zap"
)

const (
	nsxApiBase        = "/api/v1"
	nsxAPITimeout     = time.Second * 5 // http client transport timeout per request
	nsxSessionRefresh = time.Minute
)

var defaultRetryOnStatusCodes = []int{429, 500, 503} // nsx client

func newNSXClient(ctx context.Context) (*nsxt.APIClient, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("process environment variables: %w", err)
	}

	log := logger.Get(ctx)

	nsxMgr, err := url.Parse(cfg.NSXAddress)
	if err != nil {
		return nil, fmt.Errorf("parse NSX_URL server value: %w", err)
	}

	user, err := readKey("username")
	if err != nil {
		return nil, fmt.Errorf("read nsx username secret value: %w", err)
	}
	if user == "" {
		return nil, fmt.Errorf("nsx username secret value must not be empty")
	}

	pass, err := readKey("password")
	if err != nil {
		return nil, fmt.Errorf("read nsx password secret value: %w", err)
	}
	if pass == "" {
		return nil, fmt.Errorf("nsx password secret value must not be empty")
	}

	nsxCfg := nsxt.Configuration{
		// TODO (mgasch): CA/certs
		BasePath:        nsxApiBase,
		Host:            nsxMgr.Host,
		Scheme:          nsxMgr.Scheme,
		UserName:        user,
		Password:        pass,
		SkipSessionAuth: true, // https://github.com/vmware/go-vmware-nsxt/issues/51
		UserAgent:       "veba-nsx-tag-sync/1.0",
		Insecure:        cfg.NSXInsecure,
		DefaultHeader:   map[string]string{},
		RetriesConfiguration: nsxt.ClientRetriesConfiguration{
			MaxRetries:      3,
			RetryMinDelay:   500,
			RetryMaxDelay:   30000,
			RetryOnStatuses: defaultRetryOnStatusCodes,
		},
	}

	// we need to do this trick to use init logic for http client but customize
	// request timeouts to not block on requests forever
	if err = nsxt.InitHttpClient(&nsxCfg); err != nil {
		return nil, fmt.Errorf("initialize nsx http client: %w", err)
	}
	nsxCfg.HTTPClient.Timeout = nsxAPITimeout

	// creates unauthenticated client
	nsxClient, err := nsxt.NewAPIClient(&nsxCfg)
	if err != nil {
		return nil, fmt.Errorf("create nsx client: %w", err)
	}

	log.Info("connecting to nsx manager", zap.String("host", cfg.NSXAddress))
	if cfg.NSXInsecure {
		log.Warn("using potentially insecure connection to nsx manager", zap.Bool("insecure", cfg.NSXInsecure))
	}

	// create session
	err = nsxt.GetDefaultHeaders(nsxClient)
	if err != nil {
		return nil, fmt.Errorf("connect to nsx manager: %w", err)
	}

	return nsxClient, nil
}

// readKey reads the file from the secret path
func readKey(key string) (string, error) {
	var env Config
	if err := envconfig.Process("", &env); err != nil {
		return "", err
	}

	data, err := ioutil.ReadFile(filepath.Join(env.NSXSecretPath, key))
	if err != nil {
		return "", err
	}
	return string(data), nil
}
