package fastserver

import (
	"fmt"
	"github.com/gorilla/websocket"
	"runtime/debug"
	"testing"
)

// 50，60，50
func TestServer_Run(t *testing.T) {
	s := NewServer()
	ng := s.Group("/v1")
	ng.Use(func(ctx *Context) {

	})
	ng.Method("POST", "/iast", func(ctx *Context) {
		ctx.AbortWithStatusJson(200, H{"yes": "iat"})
	})

	//ng.Method("POST", "/iat", func(ctx *Context) {
	//	ctx.AbortWithStatusJson(200,H{"yes":"iat"})
	//})

	ng.Method("GET", "/tsts", func(ctx *Context) {
		panic("ser")
		fmt.Println("/tts")
		ctx.AbortWithStatusJson(200, H{"yes": "tts"})
	})

	ng.GET("/servers/:dx/name", func(ctx *Context) {
		dx, ok := ctx.Params.Get("dx")
		ctx.AbortWithStatusJson(200, Message{
			Message: fmt.Sprintf("dx:%s,ok:%v:apth:%s", dx, ok, ctx.Path),
		})
	})

	ng.GET("/users/:name", func(ctx *Context) {
		dx, ok := ctx.Params.Get("dx")
		ctx.AbortWithStatusJson(200, Message{
			Message: fmt.Sprintf("dx:%s,ok:%v:apth:%s", dx, ok, ctx.Path),
		})
	})

	s.NotFound(func(ctx *Context) {
		//ctx.AbortWithStatusJson(302,Message{
		//	"redirect",
		//})
		ctx.FastCtx.Redirect("http://10.1.87.70:8000", 302)
	})

	g2 := s.Group("/v3")

	g2.Use(func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				ctx.AbortWithStatusJson(500, H{"message": "internal error"})
				stack := debug.Stack()
				ctx.FastCtx.Write(stack)
				return
			}
		}()
		ctx.Next()
	})

	g2.Use(func(ctx *Context) {
		if string(ctx.FastCtx.Path()) == "/v2/isat" {
			panic("invalid path")
		}
	})

	g2.Method("GET", "/isat", func(ctx *Context) {
		ctx.AbortWithStatusJson(200, H{"ok": "yes"})
	})

	g2.GET("/websocket", func(c *Context) {
		//fc:=c.FastCtx
		wg := websocket.Upgrader{}
		req := c.StdHttpRequest()
		conn, err := wg.Upgrade(c.StdResponseWriter(), req, nil)
		if err != nil {
			c.AbortWithStatusJson(400, Message{
				err.Error(),
			})
			return
		}

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("err", err)
				return
			}
			conn.WriteMessage(websocket.TextMessage, msg)
		}
		c.Abort()

	})

	type Value struct {
		ClientIp string
	}


	g2.GET("/testVal", func(ctx *Context) {

	})
	panic(s.Run(":8080"))
}



