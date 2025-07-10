// pubsub/nats_pubsub.go
package pubsub

import (
	"github.com/nats-io/nats.go"
)

type NatsPubSub struct {
	conn *nats.Conn
}

func NewNatsPubSub(url string) (*NatsPubSub, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return &NatsPubSub{conn: nc}, nil
}

func (n *NatsPubSub) Publish(subject string, msg []byte) error {
	return n.conn.Publish(subject, msg)
}

func (n *NatsPubSub) Subscribe(subject string, handler func(msg []byte)) error {
	_, err := n.conn.Subscribe(subject, func(m *nats.Msg) {
		handler(m.Data)
	})
	return err
}

func (n *NatsPubSub) Close() error {
	n.conn.Close()
	return nil
}
