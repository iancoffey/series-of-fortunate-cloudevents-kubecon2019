package main

import (
	"context"
	"log"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/uuid"
	convoTypes "github.com/iancoffey/kubecon-cloudevent-demo-app/pkg/types"
	"github.com/joeshaw/envdecode"
)

const moodExtension = "mood-ext"

// We will get the broker sink address from the containersource reconciler.
// Those reconcilers are super useful to our CloudEvent Conversation!
type Config struct {
	// Sink URL where to send our messages as cloudevents
	Sink string `env:"SINK"`
	// Event type
	EventType string `env:"EVENT_TYPE"`
	// Name of the sender
	SenderName string `env:"SENDER_NAME"`
	// Name of the recipient which maps to a broker->trogger setup. All folks are subscribed to the to the "all" events.
	RecipientName string `env:"RECIPIENT_NAME"`
	// Message we are sending
	Message string `env:"MESSAGE"`
	// Mood to set the Extension to
	Mood string `env:"MOOD"`
}

func main() {
	var cfg Config
	err := envdecode.Decode(&cfg)

	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget(cfg.Sink),
		cloudevents.WithEncoding(cloudevents.HTTPBinaryV03),
	)
	if err != nil {
		log.Fatalf("failed to create transport %q", err)
	}

	c, err := cloudevents.NewClient(t)
	if err != nil {
		log.Fatalf("unable to create cloudevent client %q", err)
	}
	payload := convoTypes.EventPayload{
		Message: cfg.Message,
	}

	event := cloudevents.Event{
		Context: cloudevents.EventContextV03{
			ID:      uuid.New().String(),
			Type:    cfg.EventType,
			Subject: &cfg.RecipientName,
			Source:  *types.ParseURLRef(cfg.SenderName),
		}.AsV03(),
		Data: payload,
	}
	event.SetExtension(moodExtension, cfg.Mood)

	_, _, err = c.Send(context.Background(), event)
	if err != nil {
		log.Fatalf("failed to send cloudevent: %q", err)
	}
	log.Printf("Cloudevent sent to %s", cfg.RecipientName)
}
