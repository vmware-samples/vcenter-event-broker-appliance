package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

// echoSuffix indicates the event is an echo.
const echoSuffix = ".echo"

func main() {
	ctx := context.Background()

	client, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("creating HTTP client, %v", err)
	}

	log.Println("listening on :8080")
	if err := client.StartReceiver(ctx, receive); err != nil {
		log.Fatalf("start receiver: %s", err.Error())
	}
}

func receive(ctx context.Context, event cloudevents.Event) *cloudevents.Event {
	fmt.Printf("***cloud event***\n%s", event)

	// Don't return the event if it is an echo. Prevent an infinite event echo loop.
	if isEcho(event.Type()) {
		return nil
	}

	event.SetType(event.Type() + echoSuffix)
	return &event
}

func isEcho(s string) bool {
	trimmed := strings.TrimSuffix(s, echoSuffix)
	return trimmed != s
}
