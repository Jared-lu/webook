package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"webook/webook/internal/domain"
)

var ErrKeyNotExist = redis.Nil // 缓存里没数据

func NewRedisUserCache(client redis.Cmdable, expiration time.Duration) *RedisUserCache {
	return &RedisUserCache{client: client, expiration: expiration}
}

// Get 拿到User缓存
// 只要 err == nil，就认为缓存里有数据
// 如果没有数据，返回一个特定的err
func (cache *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.key(id)
	val, err := cache.client.Get(ctx, key).Bytes()
	// 缓存中有没有数据，只有操作缓存的设计者才知道
	// 因为要通过返回值区分没有数据还是数据库出错
	if err == redis.Nil {
		// 这里是为了以后要改 ErrKeyNotExist 的定义时，不影响使用者
		return domain.User{}, ErrKeyNotExist
	}
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(val, &u)
	return u, err
}

// Set 缓存User
func (cache *RedisUserCache) Set(ctx context.Context, u domain.User) error {
	// 将对象转换为json，因为redis不知道怎么处理User结构体
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := cache.key(u.Id)
	return cache.client.Set(ctx, key, val, cache.expiration).Err()
}

// key 根据user id生成key值
func (cache *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
