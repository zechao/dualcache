// pubsub/redis_pubsub.go
package pubsub

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type RedisPubSub struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisPubSub(addr string) (*RedisPubSub, error) {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	return &RedisPubSub{client: rdb, ctx: ctx}, nil
}

func (r *RedisPubSub) Publish(subject string, msg []byte) error {
	return r.client.Publish(r.ctx, subject, msg).Err()
}

func (r *RedisPubSub) Subscribe(subject string, handler func(msg []byte)) error {
	pubsub := r.client.Subscribe(r.ctx, subject)
	ch := pubsub.Channel()

	go func() {
		for m := range ch {
			handler([]byte(m.Payload))
		}
	}()

	return nil
}

func (r *RedisPubSub) Close() error {
	return r.client.Close()
}
