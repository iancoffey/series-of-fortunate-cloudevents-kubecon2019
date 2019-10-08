Thank you for coming to this session, Im ian Coffey and we are going to be spending a little time hacking around with CloudEvents. The idea being that, by both showing some code, building an example system and being able to see it function, maybe we can be utilize this time we have together.

Quick aside: I have noticed that this method of learning, combining functional examples along with terminology and info, really works well for me than, and I am also intensly interested in if you feel similarly.

Ok, wicked cool.

So what we are aiming to do in this session is to automate a simple conversation. Really its meant to be a ridiculous fake problem, but technically... I mean, arent conversations toil anyway? Who has read the SRE book? I think it might be a fit. Lets do what we do best and automate our way out of this. Maybe build a little conversation hub to handle all the tiresome effort we put into our otherwise delightful daily converations.

We will build all this using CloudEvents, Knative eventing, two (2) intentionally silly custom resources and a few simple Services, which will act as our personalities.

We can filter events using triggerfilters, which give type and source. If the source is unique, then that IP can be used. Can we setup local DNS then to resolve them into short names "Dan" "Gena"

Updated Idea:

- Two New Resources
  - Conversation
    - []Messages: ordered list of Messages to create
    - reconciler creates messages
  - Message
    - Payload: What we want to say.
    - TargetRef: Who do we want to say it to.
    - reconciler sends message via creating container source
      - source image is a custom image that just sends our message as CloudEvent
- Our "people" will be represented by Knative Services.
- Converation flow
  - Instead of directly interacting, our "people" will recieve their convo messages from our Conversation Hub.
  - When we create a new convo, our system needs to make sure each message is received in the correct order.
    - Otherwise, this convo makes no sense!
    - But how? Sequences?
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
