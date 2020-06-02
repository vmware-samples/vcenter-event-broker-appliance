package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/eventbridge/eventbridgeiface"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/pkg/errors"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/color"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/connection"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
)

const (
	// ProviderAWS variable is the name used to identify this provider in the
	// VMware Event Router configuration file
	ProviderAWS           = "aws_event_bridge"
	authMethodAWS         = "access_key"    // only this method is supported by the processor
	defaultResyncInterval = time.Minute * 5 // resync rule patterns after interval
	defaultPageLimit      = 50              // max 50 results per page for list operations
	defaultBatchSize      = 10              // max 10 input events per batch sent to AWS
)

// awsEventBridgeProcessor implements the Processor interface
type awsEventBridgeProcessor struct {
	session session.Session
	eventbridgeiface.EventBridgeAPI

	// options
	verbose        bool
	resyncInterval time.Duration
	batchSize      int
	*log.Logger

	mu         sync.RWMutex
	patternMap map[string]string // rules pattern to event bus mapping
	stats      metrics.EventStats
}

type eventPattern struct {
	Detail struct {
		Subject []string `json:"subject,omitempty"`
	} `json:"detail,omitempty"`
}

// NewAWSEventBridgeProcessor returns an AWS EventBridge processor for the given
// stream source.
func NewAWSEventBridgeProcessor(ctx context.Context, cfg connection.Config, ms metrics.Receiver, opts ...AWSOption) (Processor, error) {
	logger := log.New(os.Stdout, color.Yellow("[AWS EventBridge] "), log.LstdFlags)
	eventBridge := awsEventBridgeProcessor{
		resyncInterval: defaultResyncInterval,
		batchSize:      defaultBatchSize,
		Logger:         logger,
		patternMap:     make(map[string]string),
	}

	// apply options
	for _, opt := range opts {
		opt(&eventBridge)
	}

	var accessKey, secretKey, region, eventbus, ruleARN string
	switch cfg.Auth.Method {
	case authMethodAWS:
		accessKey = cfg.Auth.Secret["aws_access_key_id"]
		secretKey = cfg.Auth.Secret["aws_secret_access_key"]
	default:
		return nil, errors.Errorf("unsupported authentication method for processor aws_event_bridge: %s", cfg.Auth.Method)
	}

	if cfg.Options["aws_region"] == "" {
		return nil, errors.Errorf("config option %q must be specified", "aws_region")
	}
	region = cfg.Options["aws_region"]

	if cfg.Options["aws_eventbridge_rule_arn"] == "" {
		return nil, errors.Errorf("config option %q for this processor must be specified", "aws_eventbridge_rule_arn")
	}
	ruleARN = cfg.Options["aws_eventbridge_rule_arn"]

	if cfg.Options["aws_eventbridge_event_bus"] == "" {
		eventBridge.Printf("config option %q not specified, assuming %q eventbus", "aws_eventbridge_event_bus", "default")
		cfg.Options["aws_eventbridge_event_bus"] = "default"
	}
	eventbus = cfg.Options["aws_eventbridge_event_bus"]

	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
		Credentials: credentials.NewStaticCredentials(
			accessKey,
			secretKey,
			"", // a token will be created when the session is used.
		),
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not create AWS session")
	}
	eventBridge.session = *awsSession
	ebSession := eventbridge.New(awsSession)
	if ebSession == nil {
		return nil, errors.Errorf("could not create AWS event bridge session")
	}
	eventBridge.EventBridgeAPI = ebSession

	var found bool
	var nextToken *string
	for !found {
		rules, err := eventBridge.ListRulesWithContext(ctx, &eventbridge.ListRulesInput{
			EventBusName: aws.String(eventbus),        // explicitely passing eventbus name because list assumes "default" otherwise
			Limit:        aws.Int64(defaultPageLimit), // up to n results per page for requests.
			NextToken:    nextToken,
		})
		if err != nil {
			return nil, errors.Wrap(err, "could not list event bridge rules")
		}

	arnLoop:
		for _, rule := range rules.Rules {
			switch {
			case *rule.Arn == ruleARN:
				if rule.EventPattern == nil {
					return nil, errors.Errorf("rule event pattern must not be empty")
				}

				var e eventPattern
				err := json.Unmarshal([]byte(*rule.EventPattern), &e)
				if err != nil {
					return nil, errors.Wrap(err, "could not parse rule event pattern")
				}

				if len(e.Detail.Subject) == 0 { // might be a valid scenario, emit warning
					eventBridge.Println("warning: rule event pattern does not contain any subjects")
				}
				for _, s := range e.Detail.Subject {
					eventBridge.Printf("adding rule event forwarding pattern %q to processor", s)
					eventBridge.patternMap[s] = *rule.EventBusName
				}
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
			return nil, errors.Errorf("rule %s not found for configured AWS event bridge account", ruleARN)
		}
	}

	// prepopulate the metrics stats
	eventBridge.stats = metrics.EventStats{
		Provider:     ProviderAWS,
		ProviderType: cfg.Type,
		Name:         ruleARN, // Using Rule ARN to uniquely identify and represent this processor
		Started:      time.Now().UTC(),
		Invocations:  make(map[string]int),
	}

	go eventBridge.PushMetrics(ctx, ms)
	go eventBridge.syncPatternMap(ctx, eventbus, ruleARN) // periodically sync rules
	return &eventBridge, nil
}

// Process implements the stream processor interface
func (awsEventBridge *awsEventBridgeProcessor) Process(ce cloudevents.Event) error {
	if awsEventBridge.verbose {
		awsEventBridge.Printf("processing event (ID %s): %v", ce.ID(), ce)
	}

	awsEventBridge.mu.RLock()
	defer awsEventBridge.mu.RUnlock()
	if _, ok := awsEventBridge.patternMap[ce.Subject()]; !ok {
		// no event bridge rule pattern (subscription) for event, skip
		if awsEventBridge.verbose {
			awsEventBridge.Printf("pattern rule does not match, skipping event (ID %s): %v", ce.ID(), ce)
		}
		return nil
	}

	jsonBytes, err := json.Marshal(ce)
	if err != nil {
		msg := fmt.Errorf("could not marshal event %v: %v", ce, err)
		awsEventBridge.Println(msg)
		return processorError(ProviderAWS, msg)
	}

	jsonString := string(jsonBytes)
	entry := eventbridge.PutEventsRequestEntry{
		Detail:       aws.String(jsonString),
		EventBusName: aws.String(awsEventBridge.patternMap[ce.Subject()]),
		Source:       aws.String(ce.Source()),
		DetailType:   aws.String(ce.Subject()),
	}

	// update metrics
	awsEventBridge.stats.Invocations[ce.Subject()]++

	input := eventbridge.PutEventsInput{
		Entries: []*eventbridge.PutEventsRequestEntry{&entry},
	}
	awsEventBridge.Printf("sending event %s", ce.ID())
	resp, err := awsEventBridge.PutEvents(&input)
	if err != nil {
		msg := fmt.Errorf("could not send event %v: %v", ce, err)
		awsEventBridge.Println(msg)
		return processorError(ProviderAWS, msg)
	}

	if awsEventBridge.verbose {
		awsEventBridge.Printf("successfully sent event %v: %v", ce, resp)
	} else {
		awsEventBridge.Printf("successfully sent event %s", ce.ID())
	}
	return nil
}

func (awsEventBridge *awsEventBridgeProcessor) syncPatternMap(ctx context.Context, eventbus string, ruleARN string) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(awsEventBridge.resyncInterval):
			awsEventBridge.Printf("syncing pattern map for rule ARN %s", ruleARN)
			err := awsEventBridge.syncRules(ctx, eventbus, ruleARN)
			if err != nil {
				awsEventBridge.Printf("could not sync pattern map for rule ARN %s: %v", ruleARN, err)
				awsEventBridge.Printf("retrying after %v", awsEventBridge.resyncInterval)
			}
			awsEventBridge.Printf("successfully synced pattern map for rule ARN %s", ruleARN)
		}
	}
}

func (awsEventBridge *awsEventBridgeProcessor) syncRules(ctx context.Context, eventbus, ruleARN string) error {
	awsEventBridge.mu.Lock()
	// clear pattern map
	awsEventBridge.patternMap = make(map[string]string)
	awsEventBridge.mu.Unlock()

	var found bool
	var nextToken *string
	for !found {
		rules, err := awsEventBridge.ListRulesWithContext(ctx, &eventbridge.ListRulesInput{
			EventBusName: aws.String(eventbus), // explicitely passing eventbus name because list assumes "default" otherwise
			Limit:        aws.Int64(defaultPageLimit),
			NextToken:    nextToken,
		})
		if err != nil {
			return errors.Wrap(err, "could not list event bridge rules")
		}

	arnLoop:
		for _, rule := range rules.Rules {
			switch {
			case *rule.Arn == ruleARN:
				if rule.EventPattern == nil {
					return errors.Errorf("rule event pattern must not be empty")
				}

				var e eventPattern
				err := json.Unmarshal([]byte(*rule.EventPattern), &e)
				if err != nil {
					return errors.Wrap(err, "could not parse rule event pattern")
				}

				if len(e.Detail.Subject) == 0 { // might be a valid scenario, emit warning
					awsEventBridge.Println("warning: rule event pattern does not contain any subjects")
				}

				awsEventBridge.mu.Lock()
				for _, s := range e.Detail.Subject {
					awsEventBridge.Printf("adding rule event forwarding pattern %q to processor", s)
					awsEventBridge.patternMap[s] = *rule.EventBusName
				}
				awsEventBridge.mu.Unlock()

				found = true
				break arnLoop

			default:
				continue
			}
		}

		switch {
		case found: // return early
			return nil
		case rules.NextToken != nil: // try next batch of rules, if any
			nextToken = rules.NextToken
			continue
		default: // nothing found
			return errors.Errorf("rule %s not found for configured AWS event bridge account", ruleARN)
		}
	}
	return nil
}

func (awsEventBridge *awsEventBridgeProcessor) PushMetrics(ctx context.Context, ms metrics.Receiver) {
	ticker := time.NewTicker(metrics.PushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			awsEventBridge.mu.RLock()
			ms.Receive(awsEventBridge.stats)
			awsEventBridge.mu.RUnlock()
		}
	}
}
