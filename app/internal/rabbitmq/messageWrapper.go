package rabbitmq

type MessageWrapper struct {
	Message []byte
	UserId  string
}
