Thank you for coming to this session, Im ian Coffey and we are going to be spending a little time hacking around with CloudEvents. The idea being that, by both showing some code, building an example system and being able to see it function, maybe we can be utilize this time we have together.

Quick aside: I have noticed that this method of learning, combining functional examples along with terminology and info, really works well for me than, and I am also intensly interested in if you feel similarly.

Ok, wicked cool.

So what we are aiming to do in this session is to automate a simple conversation. Really its meant to be a ridiculous fake problem, but technically... I mean, arent conversations toil anyway? Who has read the SRE book? I think it might be a fit. Lets do what we do best and automate our way out of this. Maybe build a little conversation hub to handle all the tiresome effort we put into our otherwise delightful daily converations. So now we can check our future conversations into SCM, PR them, comment and then have a consesus future conversation powered by kubernetes.

We will build this using CloudEvents, Knative eventing, intentionally silly custom resources and a few simple Services, which will act as our personalities.


--

Actors boot up and run container source image which just sends registration
https://github.com/knative/eventing-contrib/blob/master/cmd/heartbeats/main.go


We will use the ContainerSource to set cloudevent Type and Source, and subscribe events using TriggerFilters

`bin/create_actor $NAME` will create an Actor and also the knative eventing TriggerFilter, which subscribes him to "com.iancoffey.conversation.message" events for his actor name. Also every actor gets subscribed to the "com.iancoffey.conversation.heartbeat" channel/broker/whatever.

TriggerFilters can filter any field it seems like? but we can filter on source if we need to and cant reach Data or something

when the actor boots it starts sending heartbeats to "com.iancoffey.conversation.heartbeat" to the `HEARTBEAT_BROKER` or whhatever, where they are all subscribed.

Every actor then needs a script. By default it gets a random one mounted by `bin/create_actor`, made from a configmap mounted from a local random script.

Actors just keep sending messages every 10 seconds. They send to whomever is in the script, if they have gotten a heartbeat, skipping those that dont.

## Components we are covering

- CloudEvents
- Knative Eventing








Updated Idea:

- Two New Resources
  - Conversation
    - []Messages: ordered list of Messages to create
    - []Actors: List of actors to create.
    - reconciler creates actors and then creates all messages
  - Message
    - Payload: What we want to say.
    - Actor: Who will be talking.
    - Recipient: Who they want to say it to.
  - Actor:
    - knative service
    - named Bob, Susan, Whatever
    - checks for Messages that it owns, starts saying them to the Recepients actors.
    - actors listem for events and post scroll to stdout

- Since all of our people live in the "village" namespace, we can use `kail` to easily snoop on their convos!
  -  we truly are villians

```
Data,
  {
    "apiVersion": "v1",
    "count": 1,
    "eventTime": null,
    "firstTimestamp": "2019-05-10T23:27:06Z",
    "involvedObject": {
      "apiVersion": "v1",
      "fieldPath": "spec.containers{busybox}",
      "kind": "Pod",
      "name": "busybox",
      "namespace": "default",
      "resourceVersion": "28987493",
      "uid": "1efb342a-737b-11e9-a6c5-42010a8a00ed"
    },
```

Anyone (service) can join the conversation (channel) by using Trigger filtering!

https://knative.dev/docs/eventing/broker-trigger/#trigger-filtering

https://godoc.org/knative.dev/eventing/pkg/apis/eventing/v1alpha1#TriggerFilter
