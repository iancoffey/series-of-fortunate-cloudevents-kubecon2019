package main

import (
	"log"

	"github.com/iancoffey/kubecon-cloudevent-demo-app/pkg/types"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/joeshaw/envdecode"
	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
)

// containersource image!

// We will get the broker sink address from the containersource reconciler.
// Those reconcilers are super useful to our CloudEvent Conversation!
type envConfig struct {
	// Sink URL where to send our messages as cloudevents
	Sink string `env:"SINK"`

	// Event type
	SenderName string `env:"EVENT_TYPE"`

	// Name of the sender
	SenderName string `env:"SENDER_NAME"`

	// Name of the recipient which maps to a broker->trogger setup. All folks are subscribed to the to the "all" events.
	RecipientName string `env:"RECIPIENT_NAME"`

	// Message we are sending
	Message string `env:"MESSAGE"`
}

func main() {
	var cfg Config
	err := envdecode.StrictDecode(&cfg)

	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget(cfg.Sink),
		cloudevents.WithEncoding(cloudevents.HTTPBinaryV02),
	)
	if err != nil {
		panic("failed to create transport, " + err.Error())
	}

	c, err := cloudevents.NewClient(t)
	if err != nil {
		panic("unable to create cloudevent client: " + err.Error())
	}

	event := cloudevents.Event{
		Context: cloudevents.EventContextV02{
			Type:   eventType,
			Source: *types.ParseURLRef(eventSource),
			Extensions: map[string]interface{}{
				"the":   42,
				"heart": "yes",
				"beats": true,
			},
		}.AsV02(),
		Data: hb,
	}

	if err := c.Send(ctx, event); err != nil {
		panic("failed to send cloudevent: " + err.Error())
	}
}
