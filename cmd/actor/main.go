package main

import (
	"log"
	"time"

	"github.com/iancoffey/kubecon-cloudevent-demo-app/pkg/types"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/joeshaw/envdecode"
	eventingv1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	sourcesv1 "github.com/knative/eventing/pkg/apis/sources/v1alpha1"
)

const (
	// In a conventional conversation, just walking up to the group is probably enough to join
	// Similarly, in our automated world, just existing in this namespace is enough for the actors to begin participating
	heartBeatRate = 10 * time.Second
	// Every so often its important to speak up in a conversation to ensure it stays interesting
	interjectInConvoRate = 30 * time.Second
	// update me
	messageImage = "iancoffey/conversation-message:latest"
)

// we need to
// - emit heartbeat event on boot up
// - listen for heartbeat events
// - the convo is:
//   - on 10 second ticker, send a random message to a random actor.
//   - when we get a message, if we have messages for the speaker, reply with a random one!

// the actors are each subscribed to the correct messages for the based on the Type and subject of the event. The Source will be the speaker.

//  sink:
//    apiVersion: eventing.knative.dev/v1alpha1
//    kind: Broker
//    name: broker-test

type Actor struct {
	Sink            string `env:"SINK,default=""`
	Name            string `env:"SERVER_HOSTNAME,default=ian"`
	MessagesMap     string `env:"MESSAGES_MAP,default=localhost"`
	ConvoListenPort uint16 `env:"PORT,default=8080"`
	StatsListenPort uint16 `env:"StatsPORT,default=8080"`
	ConvoBroker     string `env:"CONVO_BROKER,default=conversation-broker"`
	Greeting        string `env:"GREETING,default=hello"`
	Asleep          bool   `env:"ASLEEP,default=false"`
	Angry           bool   `env:"ANGRY,default=false"`
	Namespace       bool   `env:"NAMESPACE,default=work-conversation"`

	cloudEventClient cloudevents.Client
	eventingClient   eventingv1.Client
	messagesSent     map[time.Datetime]string
}

func main() {
	ctx := context.Background()

	var actor types.Actor
	err := envdecode.Decode(&actor)

	log.Printf("% wakes up and looks around", actor.Name)

	// first we introduce ourselves, and everyone will figure out our mood
	if err := a.Introduction(); err != nil {
		log.Fatalf("%s had a problem introducing themself! err=%q", err)
	}

	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget("http://localhost:8080/"),
		cloudevents.WithEncoding(cloudevents.HTTPBinaryV02),
	)
	if err != nil {
		log.Fatalf("failed to create transport: %q", err)
	}

	c, err := cloudevents.NewClient(t)
	if err != nil {
		log.Fatalf("unable to create cloudevent client: %q", err)
	}
	actor.cloudEventClient = c

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("unable to create rest config: %q", err)
	}
	// creates the clientset
	clientset, err := eventingv1.NewForConfig(config)
	if err != nil {
		log.Fatalf("unable to create eventingv1 client: %q", err)
	}
	actor.eventingClient = clientset

	// first our actor starts Listening
	go c.StartReceiver(ctx, gotMessage)
	// then it starts participating
	go actor.TickMessages()

	go actor.statsEndpoint()

	<-done
	SpeakToAll(Goodbye)
}

func (a *types.Actor) Introduction() error {
	switch {
	case actor.Asleep:
		if err := a.SpeakToAll(a.Goodnight); err != nil {
			return err
		}
	case actor.Angry:
		if err := a.SpeakToAll(a.Angry); err != nil {
			return err
		}
	default:
		if err := a.SpeakToAll(a.Goodnight); err != nil {
			return err
		}
	}
}

func (a *types.Actor) handleTerm() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println(sig)
		l
		done <- true
	}()
}

//apiVersion: sources.eventing.knative.dev/v1alpha1
//kind: ContainerSource
//metadata:
//  name: test-heartbeats
//spec:
//  template:
//    spec:
//      containers:
//        - image: <heartbeats_image_uri>
//          name: heartbeats
//          args:
//            - --period=1
//          env:
//            - name: POD_NAME
//              value: "mypod"
//            - name: POD_NAMESPACE
//              value: "event-test"
//  sink:
//    apiVersion: serving.knative.dev/v1
//    kind: Service
//    name: event-display

func (a *types.Actor) SpeakToAll(eventType string) {
	cs := &sourcev1.ContainerSource{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: actor.Name,
			Namespace:    actor.Namespace,
		},
		Spec: &sourcev1.ContainerSourceSpec{
			Template: &corev1.PodTemplateSpec{
				&metav1.ObjectMeta{
					GenerateName: actor.Name,
					Namespace:    actor.Namespace,
				},
				Spec: corev1.PodSpec{
					Containers: corev1.Container{
						Name:  "convo-message",
						Image: messageImage,
						Env: []EnvVar{
							{
								Name:  "SINK",
								Value: "idk",
							},
							{
								Name:  "EVENT_TYPE",
								Value: eventType,
							},
							{
								Name:  "SENDER_NAME",
								Value: actor.Name,
							},
							{
								Name:  "RECIPIENT_NAME",
								Value: "all",
							},
						},
					},
				},
			},
			Sink: &corev1.ObjectReference{
				Name:      actor.ConvoBroker,
				Namespace: actor.Namespace,
				Type:      eventingv1.Broker,
			},
		},
	}

}

func SpeakToPerson() {
}

func (a *Actor) gotMessage(ctx context.Context, event cloudevents.Event) error {
	data := &types.Conversation{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("Got Data Error: %s\n", err.Error())
	}
	fmt.Printf("Got Data: %+v\n", data)

	fmt.Printf(" %+v\n", cloudevents.HTTPTransportContextFrom(ctx))

	fmt.Printf(" TIME TO REPLY - CREATE CONTAINER SOURCE\n")
	return nil
}

// Start a CloudEvent listener!
// We should only be getting messages we want due to our filter on the Trigger.
func (a *Actor) EventServer(actor *Actor) {
	log.Fatalf("failed to start receiver: %s", c.StartReceiver(ctx, gotEvent))

}

// Every `interjectInConvoRate` seconds we run a tiny heartbeat image in a containersource
func (a *Actor) TickMessages() {
}

func (a *types.Actor) ContainerSource(eventType, recipientName string) {
	return &sourcev1.ContainerSource{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: actor.Name,
			Namespace:    actor.Namespace,
		},
		Spec: &sourcev1.ContainerSourceSpec{
			Template: &corev1.PodTemplateSpec{
				&metav1.ObjectMeta{
					GenerateName: actor.Name,
					Namespace:    actor.Namespace,
				},
				Spec: corev1.PodSpec{
					Containers: corev1.Container{
						Name:  "convo-message",
						Image: messageImage,
						Env: []EnvVar{
							{
								Name:  "EVENT_TYPE",
								Value: eventType,
							},
							{
								Name:  "SENDER_NAME",
								Value: actor.Name,
							},
							{
								Name:  "RECIPIENT_NAME",
								Value: recipientName,
							},
						},
					},
				},
			},
			Sink: &corev1.ObjectReference{
				Name:      actor.ConvoBroker,
				Namespace: actor.Namespace,
				Type:      eventingv1.Broker,
			},
		},
	}
}
