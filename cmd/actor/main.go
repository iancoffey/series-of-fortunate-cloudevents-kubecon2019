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
	// Every so often its important to speak up in a conversation to ensure it stays interesting
	interjectInConvoRate = 20 * time.Second
	// bind address
	localBindAddress = "http://0.0.0.0:8080/"
	// we need time for dns to start resolving
	wakeUpDelay = 10 * time.Second
)

func main() {
	ctx := context.Background()
	rand.Seed(time.Now().UTC().UnixNano())

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

	// setup CloudEvents client
	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget(localBindAddress),
		cloudevents.WithEncoding(cloudevents.HTTPBinaryV1),
	)
	if err != nil {
		log.Fatalf("failed to create transport: %q", err)
	}

	c, err := cloudevents.NewClient(t)
	if err != nil {
		log.Fatalf("unable to create cloudevent client: %q", err)
	}
	// We need to pause a sec and let everything settle
	time.Sleep(wakeUpDelay)

	// Quick gc on the way upto ensure the lane is open
	actor.GarbageCollect(true)

	// First our actor starts Listening
	go c.StartReceiver(ctx, actor.GotMessage)

	// now we can introduce ourselves
	if err := actor.Introduction(); err != nil {
		log.Fatalf("%s had a problem introducing themself. err=%q", actor.Name, err)
	}

	// then we can start our conversation Ticker
	actor.TickMessages()
	if actor.Debug {
		log.Println("Done!")
	}
}
