package fastserver

import (
	"fmt"
	"git.iflytek.com/AIaaS/nameServer/tools/str"
	"github.com/valyala/fasthttp"
	"net"
	"os"
	"os/signal"
	"path"
	"runtime/debug"
	"sync"
	"syscall"
	"time"
)

type Message struct {
	Message string `json:"message"`
}

var tostring = str.ToString

func (c HandlersChain) Last() Handler {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

// http server
// 基于fastHttp 封装路由功能
// 路由部分参考gin 的路由树，支持path 参数
type Server struct {
	RouterGroup     //
	closed          bool
	stopWg          sync.WaitGroup // 用于优雅启动停止时，等待所有的请求处理完毕时再退出
	notFoundHandler HandlersChain  // 找不到路由时默认执行的路由
}

func (s *Server) stopWgCounter(ctx *Context) {
	s.stopWg.Add(1)
	ctx.Next()
	s.stopWg.Done()
}

func NewServer() *Server {
	s := &Server{
		RouterGroup: RouterGroup{
			path:       "",
			handlers:   nil,
			routerTree: &routerTree{},
		},
	}
	s.routerTree.server = s
	//每次接受到一个请求，wg +1 ，处理完一个请求，wg -1

	//stopWg := sync.WaitGroup{}
	s.Use(s.stopWgCounter)
	//s.Use(DefaultRecover)

	return s
}

func (r *Server) NotFound(handler Handler) {
	r.notFoundHandler = combineHandlers(r.handlers, handler)
}

func (s *Server) Run(addr string) error {
	// listen
	ls, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}
	go func() {
		// 启动服务
		if err := fasthttp.Serve(ls, func(ctx *fasthttp.RequestCtx) {
			c := getContext()
			c.reset()
			c.FastCtx = ctx
			c.Path = tostring(ctx.Path())
			c.Method = tostring(ctx.Method())
			c.handlers = nil
			c.RequestURI = tostring(ctx.RequestURI())
			s.routerTree.handleHTTPRequest(c)
			putContext(c)
		}); err != nil {
			if s.closed { // 正常关闭直接return
				return
			} else {
				panic(err) // 否则panic
			}
		}
	}()
	//监听退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGSTOP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	sig := <-sigChan
	fmt.Printf("server receive signal:%v ,close listener and start to stop,%v\n", sig, time.Now().String())
	s.closed = true
	// 获取到退出信号时，关闭listener，不再接受新的请求，并且等待所有的请求处理完毕后退出
	ls.Close()
	s.stopWg.Wait()
	fmt.Printf("server successful stoped  %v \n", time.Now().String())
	return nil
}

type Handler func(ctx *Context)

type HandlersChain []Handler

type RouterGroup struct {
	path       string
	handlers   []Handler
	routerTree *routerTree
}

// 添加拦截器（handlers）
func (r *RouterGroup) Use(handlers ...Handler) {
	r.handlers = append(r.handlers, handlers...)
}

func (r *RouterGroup) Method(method string, pth string, handler Handler) {
	pth = path.Join(r.path, pth)
	r.routerTree.addRoute(method, pth, combineHandlers(r.handlers, handler))
}

//
func (r *RouterGroup) GET(path string, handler Handler) {
	r.Method("GET", path, handler)
}

func (r *RouterGroup) POST(path string, handler Handler) {
	r.Method("POST", path, handler)
}

func (r *RouterGroup) PUT(path string, handler Handler) {
	r.Method("PUT", path, handler)
}

func (r *RouterGroup) DELETE(path string, handler Handler) {
	r.Method("DELETE", path, handler)
}

func (r *RouterGroup) PATCH(path string, handler Handler) {
	r.Method("PATCH", path, handler)
}

func (r *RouterGroup) HEAD(path string, handler Handler) {
	r.Method("HEAD", path, handler)
}

func (r *RouterGroup) OPTION(path string, handler Handler) {
	r.Method("OPTION", path, handler)
}

//创建一个group，可以构建新的拦截器链路，新的group 会继承父拦截器（handler）
func (r *RouterGroup) Group(path string) *RouterGroup {
	g := &RouterGroup{
		path:       path,
		handlers:   combineHandlers(r.handlers), // 这个地方需要复制一份，否则会出现不同的group handler 相互覆盖的情况
		routerTree: r.routerTree,
	}
	return g
}

//recover handler
// 兜底，防止业务逻辑出现panic导致服务直接不可用
func DefaultRecover(c *Context) {
	defer func() {
		if err := recover(); err != nil {
			stack := debug.Stack()
			fmt.Fprintf(os.Stdout, "panic: err:%v stack:%s", err, tostring(stack))
			server500(c)
			return
		}
	}()
	c.Next()
}

func combineHandlers(hs HandlersChain, handler ...Handler) HandlersChain {
	targets := make(HandlersChain, len(hs), len(hs)+len(handler))
	copy(targets, hs)
	return append(targets, handler...)
}
