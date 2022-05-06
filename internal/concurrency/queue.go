package concurrency

// WorkQueue ...
type WorkQueue (chan struct{})

// NewWorkQueue ...
func NewWorkQueue(size int) WorkQueue {
	return make(chan struct{}, size)
}

// Run ...
func (wq WorkQueue) Run(fn func()) {
	wq <- struct{}{}
	defer func() { <-wq }()
	fn()
}

// RunNoBlock ...
func (wq WorkQueue) RunNoBlock(fn func()) (ok bool) {
	select {
	case wq <- struct{}{}:
		defer func() { <-wq }()
		ok = true
		fn()
	default:
	}
	return
}
