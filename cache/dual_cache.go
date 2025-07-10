// cache/dual_cache.go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/bluele/gcache"
	"github.com/go-redis/redis/v8"

	"github.com/zechao158/dualcache/pubsub"
)

type DualCache[K comparable, V any] struct {
	localCache gcache.Cache
	redis      *redis.Client
	ps         pubsub.PubSub
	ctx        context.Context
	topic      string
}

func NewDualCache[K comparable, V any](ctx context.Context, redisAddr string, ps pubsub.PubSub, topic string) *DualCache[K, V] {
	local := gcache.New(5000).LRU().Expiration(5 * time.Minute).Build()
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	dc := &DualCache[K, V]{
		localCache: local,
		redis:      rdb,
		ps:         ps,
		ctx:        ctx,
		topic:      topic,
	}

	// 启动订阅监听
	ps.Subscribe(topic, func(msg []byte) {
		key := string(msg)
		log.Printf("[Cache Sync] Invalidating key %s", key)
		dc.localCache.Remove(key)
	})

	return dc
}

func (c *DualCache[K, V]) Get(key K, loader func(K) (V, error)) (V, error) {
	if val, err := c.localCache.GetIFPresent(key); err == nil {
		return val.(V), nil
	}

	redisKey := fmt.Sprintf("cache:%v", key)
	valStr, err := c.redis.Get(c.ctx, redisKey).Result()
	if err == nil {
		var val V
		if err := json.Unmarshal([]byte(valStr), &val); err == nil {
			c.localCache.Set(key, val)
			return val, nil
		}
	}

	// 模拟DB查询
	val, err := loader(key)
	if err != nil {
		var zero V
		return zero, err
	}
	c.Set(key, val, 5*time.Minute)
	return val, nil
}

func (c *DualCache[K, V]) Set(key K, value V, ttl time.Duration) {
	c.localCache.SetWithExpire(key, value, ttl)
	valBytes, _ := json.Marshal(value)
	redisKey := fmt.Sprintf("cache:%v", key)
	c.redis.Set(c.ctx, redisKey, valBytes, ttl)

	// 通知分布式同步
	c.ps.Publish(c.topic, []byte(fmt.Sprintf("%v", key)))
}
