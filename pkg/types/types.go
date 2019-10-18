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
)

const (
	MessageEventType = "com.iancoffey.conversation.message"

	// ohhh fun. lets cast this last and put everyone to sleep
	SleepSpellEventType = "com.iancoffey.conversation.sleepspell"

	// Here we have the moods our actors will find themselves in
	HelloType        = "message.hello"
	GoodnightType    = "message.goodbye"
	ConversationType = "message.conversation"
	AsleepType       = "message.asleep"
	AngryType        = "message.angry"

	// Rate at which we want to break the silence
)

var (
	// Basically, how often to speak up
	interjectInConvoRate = 30 * time.Second
	// Our own little chaos monkey
	maxLifeTime = time.Duration(rand.Intn(240)) * time.Second
)

type Actor struct {
	Name            string `env:"NAME,default=ian"`
	MessagesData    string `env:"MESSAGES_DATA,default={}"`
	ConvoListenPort uint16 `env:"PORT,default=8080"`
	StatsListenPort uint16 `env:"STATS_PORT,default=8082"`
	ConvoBroker     string `env:"CONVO_BROKER,default=conversation-broker"`
	Greeting        string `env:"GREETING,default=hello"`
	Asleep          bool   `env:"ASLEEP,default=false"`
	Angry           bool   `env:"ANGRY,default=false"`
	Debug           bool   `env:"DEBUG,default=false"`
	Namespace       string `env:"NAMESPACE,default=work-conversation"`
	MessageImage    string `env:"MESSAGE_IMAGE,default=iancoffey/conversation-message:latest"`

	CloudEventClient cloudevents.Client
	EventingClient   eventingClientset.Clientset

	Conversation Conversation
	actors       []string // list of this actors friends names
}

// The exchange described both what it would send in this context and what it will respond with when necessary!
// Since there can be N exchanges per conversation topc
type Exchange struct {
	Output string `json:"output,omitempty"`
	Input  string `json:"input,omitempty"`
}

type EventPayload struct {
	Message string `json:"Message"`
	Off     bool   `json:"OFF,default=false"`
}

// Standard topic types, which map directly to CloudEvent Type
// eg, "com.iancoffey.conversation.compliment", source bob subject frank/all
// "Unix" mode = "hello = EHLO, Angry = OOMKILLER Message"
type Conversation struct {
	Hello        []Exchange `json:"hello,omitempty"`        // either sending or being sent hello
	Goodbye      []Exchange `json:"goodbye,omitempty"`      // if we need to send or get sent Goodbye messages
	Conversation []Exchange `json:"conversation,omitempty"` // Once running, a ticker will just schedule Convos every N seconds
	Asleep       []Exchange `json:"asleep,omitempty"`       // zzzz
	Angry        []Exchange `json:"angry,omitempty"`        // enable angry mode, which replies and responds only with angry stuff
}

type ConversationManifests struct {
	Conversations []Conversation `json:"conversations,omitempty"`
}

func (a *Actor) AsleepMessage() Exchange {
	return a.Conversation.Asleep[rand.Intn(len(a.Conversation.Asleep))]
}
func (a *Actor) HelloMessage() Exchange {
	return a.Conversation.Hello[rand.Intn(len(a.Conversation.Hello))]
}
func (a *Actor) AngryMessage() Exchange {
	return a.Conversation.Angry[rand.Intn(len(a.Conversation.Angry))]
}
func (a *Actor) GoodbyeMessage() Exchange {
	return a.Conversation.Goodbye[rand.Intn(len(a.Conversation.Goodbye))]
}
func (a *Actor) ConversationMessage() Exchange {
	return a.Conversation.Conversation[rand.Intn(len(a.Conversation.Conversation))]
}

func (a *Actor) Introduction() error {
	switch {
	case a.Asleep:
		if err := a.SpeakToAll(MessageEventType, AsleepType, a.AsleepMessage()); err != nil {
			return err
		}
	case a.Angry:
		if err := a.SpeakToAll(MessageEventType, AngryType, a.AngryMessage()); err != nil {
			return err
		}
	default:
		if err := a.SpeakToAll(MessageEventType, HelloType, a.HelloMessage()); err != nil {
			return err
		}
	}
	return nil
}
func (a *Actor) SpeakToAll(eventType, mood string, e Exchange) error {
	cs := a.ContainerSource(eventType, "all", e.Output, mood)
	_, err := a.EventingClient.SourcesV1alpha1().ContainerSources(a.Namespace).Create(cs)
	return err
}

func (a *Actor) StatsEndpoint() {
}

// Speak to random actor you have heard from
func (a *Actor) SpeakToActor(eventType, mood string, e Exchange) error {
	if len(a.actors) == 0 {
		return errors.New("This actor has no friends! And you shouldnt be here.")
	}
	target := a.actors[rand.Intn(len(a.actors))]

	cs := a.ContainerSource(eventType, target, e.Output, mood)
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

func (a *Actor) GotMessage(ctx context.Context, event cloudevents.Event) error {
	if a.Debug {
		log.Printf("Got message ID %s Source %s Subject %s\n", event.ID(), event.Source(), event.Subject())
	}
	// We dont need to be talking to ourselves
	if event.Source() == a.Name {
		return nil
	}

	// Lets add this actor to our list
	a.AddToFriends(event.Source())

	if a.Debug {
		log.Printf("Friends List -> %s", a.actors)
	}

	payload := &EventPayload{}
	if err := event.DataAs(&payload); err != nil {
		log.Printf("Got Data Error: %s\n", err.Error())
		return err
	}

	log.Printf("conversation-> (%s) %s said %s", a.Name, payload.Message, event.Source())

	return nil
}

// Every `interjectInConvoRate` seconds emit a little message via our "messages" containersource image
// The actor will only live for maxLifeTime. After that it says goodbye and exits.
func (a *Actor) TickMessages() {
	ticker := time.NewTicker(interjectInConvoRate)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				a.GarbageCollect()
				if len(a.actors) == 0 {
					continue
				}
				if err := a.SpeakToActor(MessageEventType, ConversationType, a.ConversationMessage()); err != nil {
					return
				}

			}
		}
	}()

	time.Sleep(maxLifeTime)
	ticker.Stop()
	done <- true

	if a.Debug {
		log.Println("Exiting TickMessages")
	}
}

func (a *Actor) GarbageCollect() {
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
		log.Println("garbage collection item list count: %d", len(list.Items))
	}

	for _, cs := range list.Items {
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

func (a *Actor) ContainerSource(eventType, recipientName, message, mood string) *sourcesv1.ContainerSource {
	if a.Debug {
		log.Printf("ContainerSource type:%s receipient: message:%s mood:%s", eventType, recipientName, message, mood)
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
								{
									Name:  "MOOD",
									Value: mood,
								},
							},
						},
					},
				},
			},
			Sink: &corev1.ObjectReference{
				Name:       a.ConvoBroker,
				Namespace:  a.Namespace,
				Kind:       "Broker",
				APIVersion: "eventing.knative.dev/v1alpha1",
			},
		},
	}
}

// on Term or Int, send everyone Goodbye!
//func (a *Actor) HandleTerm(done chan<- bool) {
//	sigs := make(chan os.Signal, 1)

//	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
//	go func() {
//		<-sigs
//		a.SpeakToAll(MessageEventType, GoodnightType, a.GoodbyeMessage())
//		done <- true
//	}()
//}
