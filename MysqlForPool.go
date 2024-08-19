package main

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type ConnectionPool struct {
	pool        chan *sql.DB
	maxOpen     int
	idleTimeout time.Duration
	mu          sync.Mutex
}

func NewConnectionPool(maxOpen int, idleTimeout time.Duration) (*ConnectionPool, error) {
	if maxOpen <= 0 {
		return nil, fmt.Errorf("maxOpen must be greater than 0")
	}

	dsn := "root:root@tcp(localhost:8889)/gotest" // 替换为你的DSN
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if db.Ping() != nil {
		return nil, fmt.Errorf("failed to ping database")
	}

	pool := make(chan *sql.DB, maxOpen)
	for i := 0; i < maxOpen; i++ {
		pool <- db
	}

	return &ConnectionPool{
		pool:        pool,
		maxOpen:     maxOpen,
		idleTimeout: idleTimeout,
	}, nil
}

func (p *ConnectionPool) Get() (*sql.DB, error) {
	select {
	case db := <-p.pool:
		return db, nil
	default:
		// 如果池中没有空闲连接，根据实际需求处理，这里简单返回错误
		return nil, fmt.Errorf("no available connection in the pool")
	}
}

func (p *ConnectionPool) Put(db *sql.DB) {
	select {
	case p.pool <- db:
	default:
		// 如果池已满，根据实际需求处理，这里简单关闭连接
		db.Close()
	}
}

func (p *ConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	close(p.pool)
	for db := range p.pool {
		db.Close()
	}
}

func main() {
	pool, err := NewConnectionPool(5, 30*time.Second)
	if err != nil {
		fmt.Println("Failed to create connection pool:", err)
		return
	}

	// 模拟使用连接池
	db, err := pool.Get()
	if err != nil {
		fmt.Println("Failed to get connection from pool:", err)
		return
	}
	// 使用db执行操作...
	db.Query("select * from user limit 100")

	// 将连接放回池中
	pool.Put(db)

	// 关闭连接池
	pool.Close()
}
