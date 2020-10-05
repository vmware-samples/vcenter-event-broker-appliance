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
	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
)

const (
	defaultResyncInterval = time.Minute * 5 // resync rule patterns after interval
	defaultPageLimit      = 50              // max 50 results per page for list operations
	defaultBatchSize      = 10              // max 10 input events per batch sent to AWS
)

// EventBridgeProcessor implements the Processor interface
type EventBridgeProcessor struct {
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

// NewEventBridgeProcessor returns an AWS EventBridge processor for the given
// stream source
func NewEventBridgeProcessor(ctx context.Context, cfg *config.ProcessorConfigEventBridge, ms metrics.Receiver, opts ...AWSOption) (*EventBridgeProcessor, error) {
	logger := log.New(os.Stdout, color.Yellow("[AWS EventBridge] "), log.LstdFlags)
	eventBridge := EventBridgeProcessor{
		resyncInterval: defaultResyncInterval,
		batchSize:      defaultBatchSize,
		Logger:         logger,
		patternMap:     make(map[string]string),
	}

	// apply options
	for _, opt := range opts {
		opt(&eventBridge)
	}

	if cfg == nil {
		return nil, errors.New("no AWS EventBridge configuration found")
	}

	if cfg.Auth == nil || cfg.Auth.AWSAccessKeyAuth == nil {
		return nil, fmt.Errorf("invalid %s credentials: accessKey and secretKey must be set", config.AWSAccessKeyAuth)
	}

	accessKey := cfg.Auth.AWSAccessKeyAuth.AccessKey
	secretKey := cfg.Auth.AWSAccessKeyAuth.SecretKey

	if cfg.Region == "" {
		return nil, errors.New("region must be specified")
	}

	if cfg.RuleARN == "" {
		return nil, errors.New("rule ARN must be specified")
	}

	if cfg.EventBus == "" {
		return nil, errors.New("event bus must be specified")
	}

	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region),
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

	var (
		found     bool
		nextToken *string
	)

	for !found {
		rules, err := eventBridge.ListRulesWithContext(ctx, &eventbridge.ListRulesInput{
			EventBusName: aws.String(cfg.EventBus),    // explicitly passing eventbus name because list assumes "default" otherwise
			Limit:        aws.Int64(defaultPageLimit), // up to n results per page for requests.
			NextToken:    nextToken,
		})
		if err != nil {
			return nil, errors.Wrap(err, "could not list event bridge rules")
		}

	arnLoop:
		for _, rule := range rules.Rules {
			switch {
			case *rule.Arn == cfg.RuleARN:
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
			return nil, errors.Errorf("rule %s not found for configured AWS event bridge account", cfg.RuleARN)
		}
	}

	// prepopulate the metrics stats
	eventBridge.stats = metrics.EventStats{
		Provider:    string(config.ProcessorEventBridge),
		Type:        config.EventProcessor,
		Address:     cfg.RuleARN, // Using Rule ARN to uniquely identify and represent this processor
		Started:     time.Now().UTC(),
		Invocations: make(map[string]int),
	}

	go eventBridge.PushMetrics(ctx, ms)
	go eventBridge.syncPatternMap(ctx, cfg.EventBus, cfg.RuleARN) // periodically sync rules

	return &eventBridge, nil
}

// Process implements the stream processor interface
func (eb *EventBridgeProcessor) Process(ce cloudevents.Event) error {
	if eb.verbose {
		eb.Printf("processing event (ID %s): %v", ce.ID(), ce)
	}

	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if _, ok := eb.patternMap[ce.Subject()]; !ok {
		// no event bridge rule pattern (subscription) for event, skip
		if eb.verbose {
			eb.Printf("pattern rule does not match, skipping event (ID %s): %v", ce.ID(), ce)
		}

		return nil
	}

	jsonBytes, err := json.Marshal(ce)
	if err != nil {
		msg := fmt.Errorf("could not marshal event %v: %v", ce, err)
		eb.Println(msg)
		return processorError(config.ProcessorEventBridge, msg)
	}

	jsonString := string(jsonBytes)
	entry := eventbridge.PutEventsRequestEntry{
		Detail:       aws.String(jsonString),
		EventBusName: aws.String(eb.patternMap[ce.Subject()]),
		Source:       aws.String(ce.Source()),
		DetailType:   aws.String(ce.Subject()),
	}

	// update metrics
	eb.stats.Invocations[ce.Subject()]++

	input := eventbridge.PutEventsInput{
		Entries: []*eventbridge.PutEventsRequestEntry{&entry},
	}

	eb.Printf("sending event %s", ce.ID())
	resp, err := eb.PutEvents(&input)

	if err != nil {
		msg := fmt.Errorf("could not send event %v: %v", ce, err)
		eb.Println(msg)
		return processorError(config.ProcessorEventBridge, msg)
	}

	if eb.verbose {
		eb.Printf("successfully sent event %v: %v", ce, resp)
	} else {
		eb.Printf("successfully sent event %s", ce.ID())
	}
	return nil
}

func (eb *EventBridgeProcessor) syncPatternMap(ctx context.Context, eventbus, ruleARN string) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(eb.resyncInterval):
			eb.Printf("syncing pattern map for rule ARN %s", ruleARN)

			err := eb.syncRules(ctx, eventbus, ruleARN)
			if err != nil {
				eb.Printf("could not sync pattern map for rule ARN %s: %v", ruleARN, err)
				eb.Printf("retrying after %v", eb.resyncInterval)
			}

			eb.Printf("successfully synced pattern map for rule ARN %s", ruleARN)
		}
	}
}

func (eb *EventBridgeProcessor) syncRules(ctx context.Context, eventbus, ruleARN string) error {
	eb.mu.Lock()
	// clear pattern map
	eb.patternMap = make(map[string]string)
	eb.mu.Unlock()

	var (
		found     bool
		nextToken *string
	)

	for !found {
		rules, err := eb.ListRulesWithContext(ctx, &eventbridge.ListRulesInput{
			EventBusName: aws.String(eventbus), // explicitly passing eventbus name because list assumes "default" otherwise
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
					eb.Println("warning: rule event pattern does not contain any subjects")
				}

				eb.mu.Lock()
				for _, s := range e.Detail.Subject {
					eb.Printf("adding rule event forwarding pattern %q to processor", s)
					eb.patternMap[s] = *rule.EventBusName
				}
				eb.mu.Unlock()

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
