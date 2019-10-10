# kubecon-cloudevent-demo-app

Automated conversations imagines an alternate method of allowing semi-prescripted conversations to play out in a sandbox environment.

We will put CloudEvents and Knative Eventing to work making this machinery possible!

## Create an actor

The `bin/add_actor $NAME $GREETING` script will populate our conversation with a new actor.

No other steps are needed for the actor to join the conversation. The actor is assigned a randomized script, which is mounted in a configmap.

Actors only address other actors they have words for, but they can reply to anyone with their default greeting.
