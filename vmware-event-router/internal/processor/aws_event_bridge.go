package processor

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/eventbridge/eventbridgeiface"
	"github.com/pkg/errors"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/color"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/connection"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/events"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	// ProviderAWS is the name used to identify this provider in the
	// VMware Event Router configuration file
	ProviderAWS    = "aws_event_bridge"
	authMethodAWS  = "access_key"    // only this method is supported by the processor
	resyncInterval = time.Minute * 5 // resync rule patterns after interval
	pageLimit      = 50              // max 50 results per page for list operations
)

// awsEventBridgeProcessor implements the Processor interface
type awsEventBridgeProcessor struct {
	session session.Session
	eventbridgeiface.EventBridgeAPI
	source  string
	verbose bool
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
func NewAWSEventBridgeProcessor(ctx context.Context, cfg connection.Config, source string, verbose bool, ms *metrics.Server) (Processor, error) {
	logger := log.New(os.Stdout, color.Yellow("[AWS EventBridge] "), log.LstdFlags)
	eventBridge := awsEventBridgeProcessor{
		source:     source,
		verbose:    verbose,
		Logger:     logger,
		patternMap: make(map[string]string),
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
			EventBusName: aws.String(eventbus), // explicitely passing eventbus name because list assumes "default" otherwise
			Limit:        aws.Int64(pageLimit), // up to n results per page for requests.
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
		case found: // return early
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

// Process implements the stream processor interface TODO: handle
// throttling/batching
// https://docs.aws.amazon.com/eventbridge/latest/userguide/cloudwatch-limits-eventbridge.html#putevents-limits
func (awsEventBridge *awsEventBridgeProcessor) Process(moref types.ManagedObjectReference, baseEvent []types.BaseEvent) error {
	input, err := awsEventBridge.createPutEventsInput(baseEvent)
	if err != nil {
		awsEventBridge.Printf("could not create PutEventsInput for event(s): %v", err)
		return nil
	}

	// nothing to send
	if len(input.Entries) == 0 {
		return nil
	}

	// TODO: investigate limits on number/size of entries in a single put
	resp, err := awsEventBridge.PutEvents(&input)
	if err != nil {
		awsEventBridge.Printf("could not send event(s): %v", err)
		return nil
	}
	if awsEventBridge.verbose {
		awsEventBridge.Printf("successfully sent event(s) from source %s: %+v", awsEventBridge.source, resp)
	}
	return nil
}

func (awsEventBridge *awsEventBridgeProcessor) createPutEventsInput(baseEvent []types.BaseEvent) (eventbridge.PutEventsInput, error) {
	// TODO: Array Members: Minimum number of 1 item. Maximum number of 10 items. for []*eventbridge.PutEventsRequestEntry{}
	// https://github.com/pacedotdev/batch
	awsEventBridge.mu.Lock()
	defer awsEventBridge.mu.Unlock()

	input := eventbridge.PutEventsInput{
		Entries: []*eventbridge.PutEventsRequestEntry{},
	}

	for idx := range baseEvent {
		// process slice in reverse order to maintain Event.Key ordering
		event := baseEvent[len(baseEvent)-1-idx]

		if awsEventBridge.verbose {
			awsEventBridge.Printf("processing event [%d] of type %T from source %s: %+v", idx, event, awsEventBridge.source, event)
		}
		eventInfo := events.GetDetails(event)
		if _, ok := awsEventBridge.patternMap[eventInfo.Name]; !ok {
			// no event bridge rule pattern (subscription) for event, skip
			continue
		}
		cloudEvent := events.NewCloudEvent(event, eventInfo, awsEventBridge.source)
		jsonBytes, err := json.Marshal(cloudEvent)
		if err != nil {
			return eventbridge.PutEventsInput{}, errors.Wrapf(err, "could not marshal cloud event for vSphere event %d from source %s", event.GetEvent().Key, awsEventBridge.source)
		}

		jsonString := string(jsonBytes)
		entry := eventbridge.PutEventsRequestEntry{
			Detail:       aws.String(jsonString),
			EventBusName: aws.String(awsEventBridge.patternMap[eventInfo.Name]),
			Source:       aws.String(cloudEvent.Source),
			DetailType:   aws.String(cloudEvent.Subject),
		}
		input.Entries = append(input.Entries, &entry)

		// update metrics
		awsEventBridge.stats.Invocations[eventInfo.Name]++
	}

	return input, nil
}

func (awsEventBridge *awsEventBridgeProcessor) syncPatternMap(ctx context.Context, eventbus string, ruleARN string) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(resyncInterval):
			awsEventBridge.Printf("syncing pattern map for rule ARN %s", ruleARN)
			err := awsEventBridge.syncRules(ctx, eventbus, ruleARN)
			if err != nil {
				awsEventBridge.Printf("could not sync pattern map for rule ARN %s: %v", ruleARN, err)
				awsEventBridge.Printf("retrying after %v", resyncInterval)
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
			Limit:        aws.Int64(pageLimit),
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
			break
		case rules.NextToken != nil: // try next batch of rules, if any
			nextToken = rules.NextToken
			continue
		default: // nothing found
			return errors.Errorf("rule %s not found for configured AWS event bridge account", ruleARN)
		}
	}
	return nil
}

func (awsEventBridge *awsEventBridgeProcessor) PushMetrics(ctx context.Context, ms *metrics.Server) {
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
