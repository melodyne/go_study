package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"sync"
)

// 定义一个简单的连接池
type SimplePool struct {
	pool *redis.Pool
	once sync.Once
}

// 初始化连接池
func NewSimplePool(host string, size int) *SimplePool {
	return &SimplePool{
		pool: &redis.Pool{
			MaxIdle: size,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", host)
			},
		},
	}
}

// 获取连接
func (p *SimplePool) Get() redis.Conn {
	return p.pool.Get()
}

func main() {
	pool := NewSimplePool("localhost:6379", 10)
	conn := pool.Get()
	_, err := conn.Do("PING")
	if err != nil {
		fmt.Println("Redis connection error:", err)
		return
	}
	fmt.Println("Redis connected successfully!")
	// 使用完毕后关闭连接
	defer conn.Close()
}
