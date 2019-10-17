package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/iancoffey/kubecon-cloudevent-demo-app/pkg/types"
	"github.com/joeshaw/envdecode"
	"k8s.io/client-go/rest"
	eventingClientset "knative.dev/eventing/pkg/client/clientset/versioned"
)

const (
	// In a conventional conversation, just walking up to the group is probably enough to join
	// Similarly, in our automated world, just existing in this namespace is enough for the actors to begin participating
	heartBeatRate = 10 * time.Second
	// Every so often its important to speak up in a conversation to ensure it stays interesting
	interjectInConvoRate = 30 * time.Second
	// bind address
	localBindAddress = "http://0.0.0.0:8080/"
)

func main() {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())

	var actor types.Actor
	err := envdecode.StrictDecode(&actor)

	log.Printf("%s wakes up and looks around", actor.Name)

	// we need to load our
	manifest := &types.ConversationManifests{}
	if err := json.Unmarshal([]byte(actor.MessagesData), manifest); err != nil {
		log.Fatalf("%s had an error ", err)
	}
	convos := manifest.Conversations
	// we want to pick a random conversation profile for our new actor
	actor.Conversation = (convos)[rand.Intn(len(convos))]

	if actor.Debug {
		log.Println("Creating clientset")
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("unable to create rest config: %q", err)
	}
	// creates the clientset
	clientset, err := eventingClientset.NewForConfig(config)
	if err != nil {
		log.Fatalf("unable to create eventingv1 client: %q", err)
	}
	actor.EventingClient = *clientset

	if actor.Debug {
		log.Println("Creating cloudevent client")
	}

	// setup CloudEvents client
	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget(localBindAddress),
		cloudevents.WithEncoding(cloudevents.HTTPBinaryV02),
	)
	if err != nil {
		log.Fatalf("failed to create transport: %q", err)
	}

	c, err := cloudevents.NewClient(t)
	if err != nil {
		log.Fatalf("unable to create cloudevent client: %q", err)
	}

	if actor.Debug {
		log.Println("Starting CloudEvent Receiver")
	}

	// first our actor starts Listening
	go c.StartReceiver(ctx, actor.GotMessage)

	if actor.Debug {
		log.Println("Introduction time")
	}

	// now we can introduce ourselves, and everyone can start to figure out our mood
	if err := actor.Introduction(); err != nil {
		log.Fatalf("%s had a problem introducing themself! err=%q", err)
	}

	if actor.Debug {
		log.Println("TickMessages time")
	}

	// then we can start our conversation Ticker
	go actor.TickMessages()

	if actor.Debug {
		log.Println("Stats Endpoint")
	}

	// then we can enable our stats endpoint
	// we can maybe use prometheus to see our conversation metrics go nuts!
	go actor.StatsEndpoint()

	done := make(chan bool, 1)
	go actor.HandleTerm(done)
	<-done
	if actor.Debug {
		log.Println("After Done!")
	}
}
