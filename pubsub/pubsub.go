package pubsub

// PubSub is an interface for a publish-subscribe messaging system.
type PubSub interface {
	Publish(subject string, msg []byte) error
	Subscribe(subject string, handler func(msg []byte)) error
	Close() error
}
