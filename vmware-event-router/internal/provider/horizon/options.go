package horizon

import (
	"time"

	"github.com/benbjohnson/clock"
)

// Option allows for customization of the horizon event provider
// TODO: change signature to return errors
type Option func(*EventStream)

// WithClock injects a custom clock to the horizon event provider
func WithClock(c clock.Clock) Option {
	return func(stream *EventStream) {
		stream.clock = c
	}
}

// WithPollInterval sets a custom Horizon API polling interval
func WithPollInterval(interval time.Duration) Option {
	return func(stream *EventStream) {
		stream.pollInterval = interval
	}
}
