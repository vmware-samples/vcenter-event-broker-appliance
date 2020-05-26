// +build unit

package processor

import (
	"fmt"
	"testing"
)

func Test_processor_error(t *testing.T) {
	tests := []struct {
		title       string
		provider    string
		errMessage  error
		expectedErr string
	}{
		{
			title:       "Event Bridge unable to process VmPoweredOnEvent",
			provider:    ProviderAWS,
			errMessage:  fmt.Errorf("could not create PutEventsInput for event(s): VmPoweredOnEvent"),
			expectedErr: "aws_event_bridge: could not create PutEventsInput for event(s): VmPoweredOnEvent",
		},
		{
			title:       "OpenFaaS processor unable to handle VmPoweredOnEvent",
			provider:    ProviderOpenFaaS,
			errMessage:  fmt.Errorf("error handling event: VmPoweredOnEvent"),
			expectedErr: "openfaas: error handling event: VmPoweredOnEvent",
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			actualErr := processorError(test.provider, test.errMessage)
			if actualErr.Error() != test.expectedErr {
				t.Errorf("Expected error: %s got: %s", test.expectedErr, actualErr.Error())
			}
		})
	}
}
