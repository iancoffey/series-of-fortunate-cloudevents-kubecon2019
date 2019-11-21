# kubecon-cloudevent-demo-app

This session will work to leverage CloudEvents and Knative Eventing to build a solution to automated conversations, with the conversation flow provided by the base projects.

New actors will boot and make themselves known to the group, and after this, they will begin conversing!

**The actors only print to stdout what they hear, not what they said**. So everything shown in the ./bin/listen script are events that have successfully reached the destination.

## Pre-Reqs

- [Kind](https://github.com/kubernetes-sigs/kind)
- [Kail](https://github.com/boz/kail)

## Boot the demo system

The `./bin/up` script will bring up the entire demo system for you, by creating a k8s cluster with Kind, installing knative via Gloo and create the necessary conversation details.

## Create an actor

The `bin/add_actor $NAME` script will create and populate the namespace with a new actor, as well as the Knative Triggers necessary to allow them to join the conversation broker.

No other steps are needed for the actor to join the conversation. The actor is assigned a randomized script, which is mounted in a configmap.

They will address the whole group, or individuals who they know about.

## Conversation

All of our actors are provided a conversation script, which lets them chat.

## Log / Output Commands for demo

As you can see there is a lot going on:

kail -n work-conversations

kubectl logs -l "actor=fred" -n work-conversation --all-containers

kubectl describe containersource.sources.eventing.knative.dev -n work-conversation

# KubeConfig

export KUBECONFIG="$(kind get kubeconfig-path --name="conversations")"
