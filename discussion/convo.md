# Commands

- **Fields**
  - `./bin/up`: Kind, Gloo, Knative, our example Broker and RBAC
  - `cat manifest/setup/triggers.yaml`: Our broker
  - `kail -n work-conversation`: Lets check out our silent channel
  - `./bin/add_actor (fred|jane|tracy|peter)`
- **Automated Convo Diagram**
  - `./bin/listen`: See the conversation
  - `./bin/events`: This is the raw look at the events on the wire
- **Magic event**
  - `cat triggers-entrance.yaml`: lets look at the new triggers
  - `kubectl apply -f triggers-entrance.yaml`: apply that yaml
  - `./bin/listen`: Whoa, things have changed
