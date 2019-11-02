package main

import (
	"context"
	"log"

	cloudevents "github.com/cloudevents/sdk-go"
	cetypes "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/uuid"
	convoTypes "github.com/iancoffey/kubecon-cloudevent-demo-app/pkg/types"
	"github.com/joeshaw/envdecode"
)

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
	// Convey the power of the shiny object?
	Shiny bool `env:"SHINY"`
}

func main() {
	var cfg Config
	err := envdecode.Decode(&cfg)

	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget(cfg.Sink),
		cloudevents.WithEncoding(cloudevents.HTTPBinaryV1),
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
		Shiny:   cfg.Shiny,
	}

	event := cloudevents.Event{
		Context: &cloudevents.EventContextV1{
			ID:      uuid.New().String(),
			Type:    cfg.EventType,
			Subject: &cfg.RecipientName,
			Source:  *types.ParseURIRef(cfg.SenderName),
		},
		Data: payload,
	}
	event.SetSpecVersion(cetypes.CloudEventsVersionV1)

	_, _, err = c.Send(context.Background(), event)
	if err != nil {
		log.Fatalf("failed to send cloudevent: %q", err)
	}
	log.Printf("Cloudevent sent to %s", cfg.RecipientName)
	log.Printf("Cloudevent Version %s", event.Context.GetSpecVersion())
}
