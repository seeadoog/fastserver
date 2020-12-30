package fastserver

import "sync"

var (
	contextPool = sync.Pool{}
)

func init() {
	contextPool.New = func() interface{} {
		return &Context{}
	}
}

func getContext() *Context {
	return contextPool.Get().(*Context)
}

func putContext(c *Context) {
	contextPool.Put(c)
}
