package conversation

const (
	helloEventType        = "com.iancoffey.conversation.message.hello"
	goodnightEventType    = "com.iancoffey.conversation.message.goodnight"
	conversationEventType = "com.iancoffey.conversation.message.conversation"
	asleepEventType       = "com.iancoffey.conversation.message.asleep"
	angryEventType        = "com.iancoffey.conversation.message.angry"
)

// The exchange described both what it would send in this context and what it will respond with when necessary!
// Since there can be N exchanges per conversation topc
type Exchange struct {
	Output string `json:"output,omitempty"`
	Input  string `json:"input,omitempty"`
}

// Standard topic types, which map directly to CloudEvent Type
// eg, "com.iancoffey.conversation.compliment", source bob subject frank/all
// "Unix" mode = "hello = EHLO, Angry = OOMKILLER Message"
type Conversation struct {
	Hello        []Exchange `json:"hello,omitempty"`        // either sending or being sent hello
	Goodbye      []Exchange `json:"goodbye,omitempty"`      // if we need to send or get sent Goodbye messages
	Conversation []Exchange `json:"conversation,omitempty"` // Once running, a ticker will just schedule Convos every N seconds
	Asleep       []Exchange `json:"asleep,omitempty"`       // zzzz
	Angry        []Exchange `json:"angry,omitempty"`        // enable angry mode, which replies and responds only with angry stuff
}
