package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var db *sql.DB

var store = sessions.NewCookieStore([]byte("secret"))

type LoginForm struct {
	User     string `form:"user" binding:"required"`
	Password string `form:"password" binding:"required"`
}

type Person struct {
	Name     string    `form:"name"`
	Address  string    `form:"address"`
	Birthday time.Time `form:"birthday" time_format:"2006-01-02" time_utc:"1"`
}

func sessionMiddleware(c *gin.Context) {
	session, err := store.Get(c.Request, "session-name")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Set("session", session)
}

func main() {
	// 打开数据库连接（实际上创建了一个连接池）
	db, err := sql.Open("mysql", "root:root@tcp(localhost:8889)/gotest")
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

	r := gin.Default()
	// 使用Gorilla sessions中间件
	r.Use(sessionMiddleware)

	r.LoadHTMLGlob("templates/*")

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/set", setSession)
	r.GET("/get", getSession)
	r.GET("/rm", rmSession)
	r.GET("/index", index)
	r.GET("/logout", logout)
	r.POST("/login", login)
	r.Run() // 监听并在 0.0.0.0:8080 上启动服务
}

func index(c *gin.Context) {

	session := c.MustGet("session").(*sessions.Session)
	user := session.Values["username"]
	fmt.Printf("登录：%v\n", user)
	//c.JSON(http.StatusOK, gin.H{"user": user})
	//var user = sessions.NewCookieStore([]byte(os.Getenv("user")))
	// 如果是 `GET` 请求，只使用 `Form` 绑定引擎（`query`）。
	// 如果是 `POST` 请求，首先检查 `content-type` 是否为 `JSON` 或 `XML`，然后再使用 `Form`（`form-data`）。
	// 查看更多：https://github.com/gin-gonic/gin/blob/master/binding/binding.go#L88

	// 获取session
	//session, _ := store.Get(c.Request, "user")

	// 读取session值
	//foo := session.Values["foo"].(string)
	// 使用session值
	//_, _ = c.Writer.Write([]byte("Session value of 'foo': " + foo))

	if user == nil {
		content, err := ioutil.ReadFile("templates/login.html") // 读取文件内容
		if err != nil {
			fmt.Printf("读取文件失败：%v\n", err)
			return
		}
		c.Writer.Header().Set("Content-Type", "text/html")
		_, err2 := c.Writer.Write(content)
		if err2 != nil {
			log.Printf("Error writing HTML: %v", err2)
		}
		//c.String(200, string(content))
	} else {
		//c.String(200, "已登录"+fmt.Sprintf("%t", user))
		// 连接数据库
		db, err := sql.Open("mysql", "root:root@tcp(localhost:8889)/gotest")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// 检查数据库连接是否成功
		err = db.Ping()
		if err != nil {
			log.Fatal(err)
		}
		// 执行查询
		rows, err := db.Query("SELECT id,order_name,price FROM `order`")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		type Order struct {
			ID     int    `json:"id"`
			Name   string `json:"name"`
			Amount string `json:"price"`
		}

		var orders []Order
		for rows.Next() {
			var id int
			var name string
			var amount string
			if err := rows.Scan(&id, &name, &amount); err != nil {
				log.Fatal(err)
			}
			orders = append(orders, Order{ID: id, Name: name, Amount: amount})
		}

		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}
		c.Writer.Header().Set("Content-Type", "text/html")
		// 将订单数据传递给模板
		c.HTML(http.StatusOK, "order_list.html", gin.H{
			"Username":  user,
			"Orders":    orders,
			"Timestamp": time.Now().Unix(),
		})
	}

}

func login(c *gin.Context) {

	uname := c.PostForm("username")
	// 如果参数不存在，返回错误
	if uname == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is missing"})
		return
	}
	pwd := c.PostForm("password")
	// 如果参数不存在，返回错误
	if pwd == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password is missing"})
		return
	}

	// 连接数据库
	db, err := sql.Open("mysql", "root:root@tcp(localhost:8889)/gotest")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 检查数据库连接是否成功
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// 准备SQL查询，使用?作为占位符
	query := "SELECT id,name,password FROM user WHERE username = ?"

	// 执行查询
	var id int
	var name string
	var password string
	err = db.QueryRow(query, uname).Scan(&id, &name, &password)
	if err != nil {
		fmt.Println(err)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{"error": "该用户不存在～"})
			return
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
	}

	if pwd != password {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码错误"})
		return
	}

	session := c.MustGet("session").(*sessions.Session)
	session.Values["username"] = uname
	session.Save(c.Request, c.Writer)

	// 输出获取到的参数
	c.JSON(http.StatusOK, gin.H{"username": uname})
}

func logout(c *gin.Context) {
	session := c.MustGet("session").(*sessions.Session)
	session.Values["username"] = nil
	session.Save(c.Request, c.Writer)
	c.Redirect(http.StatusMovedPermanently, "/index")
}

func startPage(c *gin.Context) {
	var person Person
	// 如果是 `GET` 请求，只使用 `Form` 绑定引擎（`query`）。
	// 如果是 `POST` 请求，首先检查 `content-type` 是否为 `JSON` 或 `XML`，然后再使用 `Form`（`form-data`）。
	// 查看更多：https://github.com/gin-gonic/gin/blob/master/binding/binding.go#L88
	if c.ShouldBind(&person) == nil {
		log.Println(person.Name)
		log.Println(person.Address)
		log.Println(person.Birthday)
	}

	c.String(200, "Success")
}

func setSession(c *gin.Context) {
	session := c.MustGet("session").(*sessions.Session)
	session.Values["foo"] = "bar"
	session.Save(c.Request, c.Writer)
	c.JSON(http.StatusOK, gin.H{"message": "session set"})
}

func rmSession(c *gin.Context) {
	session := c.MustGet("session").(*sessions.Session)
	session.Values["foo"] = nil
	session.Save(c.Request, c.Writer)
	c.JSON(http.StatusOK, gin.H{"message": "session set"})
}

func getSession(c *gin.Context) {
	session := c.MustGet("session").(*sessions.Session)
	foo := session.Values["foo"]
	c.JSON(http.StatusOK, gin.H{"foo": foo})
}
