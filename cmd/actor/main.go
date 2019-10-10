package main

import (
	"log"

	"github.com/joeshaw/envdecode"
)

const (
	// In a conventional conversation, just walking up to the group is probably enough to join
	// Similarly, in our automated world, just existing in this namespace is enough for the actors to begin participating
	heartBeatRate = 10 * time.Second
	// Every so often its important to speak up in a conversation to ensure it stays interesting
	interjectInConvoRate = 30 * time.Second
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
	Name string `env:"SERVER_HOSTNAME,default=ian"`
	// location where the messages yaml is located
	MessagesMap string `env:"MESSAGES_MAP,default=localhost"`
	ListenPort  uint16 `env:"PORT,default=8080"`
	ConvoBroker string `env:"HEARTBEAT_CHAN,default=conversation-broker"`
}

func main() {
	var actor Actor
	err := envdecode.Decode(&actor)

	log.Printf("% wakes up and looks around", actor.Name)
}

// Every 10 seconds
// we need a tiny heartbeat image
func (a *Actor) Heartbeat() {
}

// Start a CloudEvent listener!
// We should only be getting messages we want due to our filter on the Trigger.
func (a *Actor) EventServer() {
}

func (a *Actor) TickMessages() {
}
