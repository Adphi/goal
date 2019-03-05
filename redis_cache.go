// cacher object to redis automatically by registering
// callback to gorm

package goal

import (
	"encoding/json"
	"fmt"

	"github.com/garyburd/redigo/redis"
)

// RedisCache implements Cacher interface
type RedisCache struct {
	pool *redis.Pool
}

func NewRedisCache(address string, maxConnections int) (*RedisCache, error) {
	pool := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", address)

		if err != nil {
			return nil, err
		}

		return c, err
	}, maxConnections)

	redisCache := &RedisCache{pool}
	if err := redisCache.initRedisPool(); err != nil {

		return nil, err
	}
	return redisCache, nil
}

func (r *RedisCache) Close() error {
	r.clearAll()
	r.pool.Close()
	return nil
}

// Get returns data for a key
func (r *RedisCache) Get(key string, val interface{}) error {
	conn, err := r.pool.Dial()
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer conn.Close()

	if key == "" {
		return nil
	}

	var reply []byte
	reply, err = redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return err
	}

	// Populate resource
	json.Unmarshal(reply, val)

	return nil
}

// Set a val for a key into Redis
func (r *RedisCache) Set(key string, val interface{}) error {
	conn, err := r.pool.Dial()
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer conn.Close()

	var data []byte
	data, err = json.Marshal(val)
	if err != nil {
		return err
	}

	_, err = conn.Do("SET", key, data)
	return err
}

// Delete a key from Redis
func (r *RedisCache) Delete(key string) error {
	conn, err := r.pool.Dial()
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer conn.Close()

	_, err = conn.Do("DEL", key)
	return err
}

// Exists checks if a key exists inside Redis
func (r *RedisCache) Exists(key string) (bool, error) {
	conn, err := r.pool.Dial()
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	defer conn.Close()

	var reply bool
	reply, err = redis.Bool(conn.Do("EXISTS", key))

	return reply, err
}

// initRedisPool initializes Redis and connection pool
func (r *RedisCache) initRedisPool() error {
	conn, err := r.pool.Dial()
	if err != nil {
		r.pool = nil
		return err
	}
	defer conn.Close()
	return nil
}

// RedisClearAll clear all data from connection's CURRENT database
func (r *RedisCache) clearAll() error {
	conn, err := r.pool.Dial()
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer conn.Close()

	_, err = conn.Do("FLUSHDB")

	if err != nil {
		fmt.Println("Error clear redis ", err)
	}

	return err
}
