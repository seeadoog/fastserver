## go fasthttp router

#### 基于fasthttp 封装的路由功能，支持原生标准库的http 协议转换

usage：

````go
	s := NewServer()
	ng := s.Group("/v1")

	ng.Method("POST", "/iast", func(ctx *Context) {
		ctx.AbortWithStatusJson(200, H{"yes": "iat"})
	})

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
		req, _ := c.StdHttpRequest()
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
	panic(s.Run(":8080"))
````
