package gtsmodel

// // ToClientAPI wraps a message that travels from the processor into the client API
// type ToClientAPI struct {
// 	APObjectType   ActivityStreamsObject
// 	APActivityType ActivityStreamsActivity
// 	Activity       interface{}
// }

// FromClientAPI wraps a message that travels from client API into the processor
type FromClientAPI struct {
	APObjectType   string
	APActivityType string
	GTSModel       interface{}
	OriginAccount  *Account
	TargetAccount  *Account
}

// // ToFederator wraps a message that travels from the processor into the federator
// type ToFederator struct {
// 	APObjectType   ActivityStreamsObject
// 	APActivityType ActivityStreamsActivity
// 	GTSModel       interface{}
// }

// FromFederator wraps a message that travels from the federator into the processor
type FromFederator struct {
	APObjectType     string
	APActivityType   string
	GTSModel         interface{}
	ReceivingAccount *Account
}
