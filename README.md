# kubecon-cloudevent-demo-app

This session will work to leverage CloudEvents and Knative Eventing to build a solution to automated conversations, with the conversation flow provided by the base projects.

New actors will boot and make themselves known to the group, and after this, they will begin conversing!

## Create an actor

The `bin/add_actor $NAME $GREETING` script will populate our conversation with a new actor.

No other steps are needed for the actor to join the conversation. The actor is assigned a randomized script, which is mounted in a configmap.

Actors only address other actors they have words for, but they can reply to anyone with their default greeting.

## Conversation

All of our actors get a conversation type, and each type defines a simple disposition.

## Log / Output Commands for demo

As you can see there is a lot going on:

kail -n work-conversations

kubectl logs -l "serving.knative.dev/service=fred" -n work-conversation --all-containers

kubectl describe containersource.sources.eventing.knative.dev -n work-conversation

# Message containersource output

kubectl logs -l actor=fred -n work-conversation

# KubeConfig

export KUBECONFIG="$(kind get kubeconfig-path --name="conversations")"
