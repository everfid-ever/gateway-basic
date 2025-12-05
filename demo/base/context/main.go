package main

import (
	"fmt"
	"net/http"
)

type HandlerFunc func(ctx *Context)

// Context 是每个请求对应的上下文实例
type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request

	// Path parameters like /users/:id
	Params map[string]string

	// index 用于中间件链的执行
	index int

	// handlers 是这个请求对应的所有 Handler（中间件 + 最终 Handler）
	handlers []HandlerFunc
}

// Next 执行下一个中间件或最终处理函数
func (c *Context) Next() {
	c.index++
	if c.index < len(c.handlers) {
		c.handlers[c.index](c)
	}
}

// JSON 简易返回 JSON（不依赖别的库）
func (c *Context) JSON(code int, data string) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(code)
	c.Writer.Write([]byte(data))
}

// String 简易返回字符串
func (c *Context) String(code int, msg string) {
	c.Writer.WriteHeader(code)
	c.Writer.Write([]byte(msg))
}

// Engine 是我们的小框架
type Engine struct {
	middlewares []HandlerFunc
	routes      map[string]HandlerFunc
}

func NewEngine() *Engine {
	return &Engine{
		middlewares: []HandlerFunc{},
		routes:      map[string]HandlerFunc{},
	}
}

// Use 注册中间件
func (e *Engine) Use(mw HandlerFunc) {
	e.middlewares = append(e.middlewares, mw)
}

// GET 注册 GET 路由和处理函数
func (e *Engine) GET(path string, handler HandlerFunc) {
	key := "GET-" + path
	e.routes[key] = handler
}

// ServeHTTP 是 Engine 成为 http.Handler 的关键
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 构建 Context
	c := &Context{
		Writer:  w,
		Request: r,
		index:   -1,
	}

	// 查找路由处理函数
	key := r.Method + "-" + r.URL.Path
	routeHandler, ok := e.routes[key]
	if !ok {
		w.WriteHeader(404)
		w.Write([]byte("404 Not Found"))
		return
	}

	// 将所有中间件 + 最终 Handler 保存到 Context
	c.handlers = append(e.middlewares, routeHandler)

	c.Next()
}

func main() {
	r := NewEngine()

	// 中间件 1：日志
	r.Use(func(c *Context) {
		fmt.Printf("[LOG] %s %s\n", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})

	// 中间件 2：性能计时
	r.Use(func(c *Context) {
		fmt.Println("[TIMER] before")
		c.Next()
		fmt.Println("[TIMER] after")
	})

	// 业务 Handler
	r.GET("/hello", func(c *Context) {
		c.String(200, "Hello from my mini Gin!")
	})

	http.ListenAndServe(":8080", r)
}
