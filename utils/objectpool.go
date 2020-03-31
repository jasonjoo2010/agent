package utils

// ObjectPool is a shared pool for non-efficient but filo get/put guaranteed
type ObjectPool struct {
	buf            chan interface{}
	createHandler  func() interface{}
	recycleHandler func(interface{})
}

// NewObjectPool create a pool with specifiec capacity and creation handler
func NewObjectPool(capacity int, createHandler func() interface{}, recycleHandler func(interface{})) *ObjectPool {
	return &ObjectPool{
		buf:            make(chan interface{}, capacity),
		createHandler:  createHandler,
		recycleHandler: recycleHandler,
	}
}

func (p *ObjectPool) Get() interface{} {
	select {
	case obj := <-p.buf:
		return obj
	default:
		return p.createHandler()
	}
}

func (p *ObjectPool) Put(obj interface{}) {
	select {
	case p.buf <- obj:
		return
	default:
		p.recycleHandler(obj)
	}
}
