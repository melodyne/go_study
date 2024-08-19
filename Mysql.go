package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

func main() {
	// 打开数据库连接（实际上创建了一个连接池）
	db, err := sql.Open("mysql", "root:root@tcp(localhost:8889)/dotest")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 设置连接池的最大连接数
	db.SetMaxOpenConns(100)

	// 设置连接池中的最大空闲连接数
	db.SetMaxIdleConns(10)

	// 确保连接可用
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
}
