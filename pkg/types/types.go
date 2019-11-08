package types

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	sourcesv1 "knative.dev/eventing/pkg/apis/sources/v1alpha1"
	eventingClientset "knative.dev/eventing/pkg/client/clientset/versioned"
	apisv1alpha1 "knative.dev/pkg/apis/v1alpha1"
)

const (
	MessageEventType = "com.iancoffey.conversation.message"

	// unfortunately, this event-type is so interesting that it distracts everyone
	DistractedSpellEventType = "com.iancoffey.conversation.distracted"
)

var (
	// Basically, how often to speak up
	interjectInConvoRate = 10 * time.Second
)

type Actor struct {
	Name            string `env:"NAME,default=ian"`
	MessagesData    string `env:"MESSAGES_DATA,default={}"`
	ConvoListenPort uint16 `env:"PORT,default=8080"`
	StatsListenPort uint16 `env:"STATS_PORT,default=8082"`
	ConvoBroker     string `env:"CONVO_BROKER,default=conversation-broker"`
	Debug           bool   `env:"DEBUG,default=false"`
	Namespace       string `env:"NAMESPACE,default=work-conversation"`
	MessageImage    string `env:"MESSAGE_IMAGE,default=iancoffey/conversation-message:latest"`

	CloudEventClient cloudevents.Client
	EventingClient   eventingClientset.Clientset

	Conversation Conversation
	actors       []string // list of this actors friends names
	messageIDs   []string // UUID of messages we have already gotten
	entranced    bool
}

// The exchange described both what it would send in this context and what it will respond with when necessary!
// Since there can be N exchanges per conversation topc
type Exchange struct {
	Output string `json:"output,omitempty"`
	Input  string `json:"input,omitempty"`
}

type EventPayload struct {
	Message string `json:"message"`
	Shiny   bool   `json:"shiny,default=false"`
}

// Standard topic types, which map directly to CloudEvent Type
// eg, "com.iancoffey.conversation.compliment", source bob subject frank/all
// "Unix" mode = "hello = EHLO, Angry = OOMKILLER Message"
type Conversation struct {
	Hello        []Exchange `json:"hello,omitempty"`        // either sending or being sent hello
	Conversation []Exchange `json:"conversation,omitempty"` // Once running, a ticker will just schedule Convos every N seconds
	Shiny        []Exchange `json:"shiny,omitempty"`        // shiny objects!
}

type ConversationManifests struct {
	Conversations []Conversation `json:"conversations,omitempty"`
}

func (a *Actor) IntroMessage() Exchange {
	switch {
	case a.entranced:
		return a.ShinyMessage()
	}

	return a.HelloMessage()
}

// Just being awake doesnt mean the actor will behave how we want them to :(
func (a *Actor) Introduction() error {
	if err := a.SpeakToAll(MessageEventType, a.IntroMessage()); err != nil {
		return err
	}

	return nil
}
func (a *Actor) SpeakToAll(eventType string, e Exchange) error {
	if a.Debug {
		log.Printf("at=speak-to-all eventType=%s", eventType)
	}
	cs := a.ContainerSource(eventType, "all", e.Output)
	_, err := a.EventingClient.SourcesV1alpha1().ContainerSources(a.Namespace).Create(cs)
	return err
}

// Speak to random actor you have heard from
func (a *Actor) SpeakToActor(eventType, target string, e Exchange) error {
	if len(a.actors) == 0 {
		return errors.New("This actor has no friends! And you shouldnt be here.")
	}

	cs := a.ContainerSource(eventType, target, e.Output)
	_, err := a.EventingClient.SourcesV1alpha1().ContainerSources(a.Namespace).Create(cs)
	return err
}

func (a *Actor) ReplyToActor(eventType, target string, e Exchange) error {
	if len(a.actors) == 0 {
		return errors.New("This actor has no friends! And you shouldnt be here.")
	}

	cs := a.ContainerSource(eventType, target, e.Input)
	_, err := a.EventingClient.SourcesV1alpha1().ContainerSources(a.Namespace).Create(cs)
	return err
}

func (a *Actor) AddToFriends(name string) {
	found := false
	for _, ac := range a.actors {
		if ac == name {
			found = true
			break
		}
	}
	if !found {
		a.actors = append(a.actors, name)
	}
}

func (a *Actor) IsDuplicate(event cloudevents.Event) bool {
	for _, id := range a.messageIDs {
		if id == event.ID() {
			return true
		}
	}
	return false
}

func (a *Actor) GotMessage(ctx context.Context, event cloudevents.Event) error {
	if a.Debug {
		// TODO: convert to k/v logging
		log.Printf("Message Type %s ID %s Source %s Subject %s", event.Type(), event.ID(), event.Source(), event.Subject())
	}
	// We want to avoid talking to ourselves
	if event.Source() == a.Name {
		return nil
	}

	// Warning: Our actors may become entranced by shiny objects!
	if event.Type() == DistractedSpellEventType {
		log.Printf("conversation-> (%s) has become distracted by a shiny object produced by %s \n", a.Name, event.Source())
		a.entranced = true
	}
	if a.entranced {
		if err := a.SpeakToActor(DistractedSpellEventType, event.Source(), a.ShinyMessage()); err != nil {
			log.Printf("SpeakToActor Error: %s", err)
			return nil
		}
	}

	// Lets add this actor to our friends list!
	a.AddToFriends(event.Source())

	if a.Debug {
		log.Printf("Friends List -> %s", a.actors)
	}

	// We wont handle duplicate events!
	if a.IsDuplicate(event) {
		if a.Debug {
			log.Printf("Duplicate event %s", event.ID())
		}
		return nil
	}
	// We want to remember what IDs we have seen!
	a.messageIDs = append(a.messageIDs, event.ID())

	// Create our event Payload
	payload := &EventPayload{}
	if err := event.DataAs(&payload); err != nil {
		log.Printf("Got Data Error: %s", err)
		return nil
	}

	// Log the raw cloudevent output so we can zoom in on them
	log.Printf("cloudevent-> \n%s\n", event)

	// Record the message to our convo dialog stream
	log.Printf("conversation->(%s) %s replied %s\n", a.Name, payload.Message, event.Source())

	// Lets reply as well
	// Our actor will respond with a similar types message - unless they are angry or asleep of course
	if err := a.ReplyToActor(MessageEventType, event.Source(), a.ConversationMessage()); err != nil {
		log.Printf("SpeakToActor Error: %s", err)
		return nil
	}
	return nil
}

// Every `interjectInConvoRate` seconds emit a little message via our "messages" containersource image
// The actor will only live for maxLifeTime. After that it says goodbye and exits.
func (a *Actor) TickMessages() {
	ticker := time.NewTicker(interjectInConvoRate)
	done := make(chan bool)

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			a.GarbageCollect(true)
			if err := a.SpeakToAll(MessageEventType, a.ConversationMessage()); err != nil {
				log.Printf("at=TickMessages error=%q", err)
				continue
			}
		}
	}

	if a.Debug {
		log.Println("Exiting TickMessages")
	}
}

func (a *Actor) GarbageCollect(force bool) {
	if a.Debug {
		log.Println("garbage collecting")
	}

	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"actor": a.Name}}
	listOptions := metav1.ListOptions{
		LabelSelector: k8slabels.Set(labelSelector.MatchLabels).String(),
		Limit:         100,
	}

	list, err := a.EventingClient.SourcesV1alpha1().ContainerSources(a.Namespace).List(listOptions)
	if err != nil {
		log.Printf("Error during GC: %q", err)
		return
	}

	if a.Debug {
		log.Printf("garbage collection item list count: %d", len(list.Items))
	}

	for _, cs := range list.Items {
		if force {
			err := a.EventingClient.SourcesV1alpha1().ContainerSources(a.Namespace).Delete(cs.Name, &metav1.DeleteOptions{})
			if err != nil {
				log.Printf("error deleting containersource %s: %q", cs.Name, err)
				return
			}
		}

		for _, cd := range cs.Status.Conditions {
			if cd.Type == "Deployed" && cd.Status == "True" {
				if a.Debug {
					log.Printf("garbage collection deleting %s", cs.Name)
				}

				err := a.EventingClient.SourcesV1alpha1().ContainerSources(a.Namespace).Delete(cs.Name, &metav1.DeleteOptions{})
				if err != nil {
					log.Printf("error deleting containersource %s: %q", cs.Name, err)
					return
				}
			}
		}
	}
}

func (a *Actor) ContainerSource(eventType, recipientName, message string) *sourcesv1.ContainerSource {
	if a.Debug {
		log.Printf("ContainerSource type: %s receipient: %s message: %q entranced: %t", eventType, recipientName, message, a.entranced)
	}

	labels := make(map[string]string)
	labels["actor"] = a.Name

	return &sourcesv1.ContainerSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-to-%s-source", a.Name, recipientName),
			Namespace: a.Namespace,
			Labels:    labels,
		},
		Spec: sourcesv1.ContainerSourceSpec{
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: a.Name,
					Namespace:    a.Namespace,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "convo-message",
							Image: a.MessageImage,
							Env: []corev1.EnvVar{
								{
									Name:  "EVENT_TYPE",
									Value: eventType,
								},
								{
									Name:  "SENDER_NAME",
									Value: a.Name,
								},
								{
									Name:  "RECIPIENT_NAME",
									Value: recipientName,
								},
								{
									Name:  "MESSAGE",
									Value: message,
								},
							},
						},
					},
				},
			},
			Sink: &apisv1alpha1.Destination{
				Ref: &corev1.ObjectReference{
					Name:       a.ConvoBroker,
					Namespace:  a.Namespace,
					Kind:       "Broker",
					APIVersion: "eventing.knative.dev/v1alpha1",
				},
			},
		},
	}
}

func (a *Actor) HelloMessage() Exchange {
	return a.Conversation.Hello[rand.Intn(len(a.Conversation.Hello))]
}
func (a *Actor) ShinyMessage() Exchange {
	return a.Conversation.Shiny[rand.Intn(len(a.Conversation.Shiny))]
}

func (a *Actor) ConversationMessage() Exchange {
	switch {
	case a.entranced:
		return a.ShinyMessage()
	}

	return a.Conversation.Conversation[rand.Intn(len(a.Conversation.Conversation))]
}
