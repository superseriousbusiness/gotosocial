package media

import "context"

func newWorkerPool(workers int) *workerPool {
	// make a pool with the given worker capacity
	pool := &workerPool{
		workerQueue: make(chan *worker, workers),
	}

	// fill the pool with workers
	for i := 0; i < workers; i++ {
		pool.workerQueue <- &worker{
			// give each worker a reference to the pool so it
			// can put itself back in when it's finished
			workerQueue: pool.workerQueue,
			data:        []byte{},
			contentType: "",
			accountID:   "",
		}
	}

	return pool
}

type workerPool struct {
	workerQueue chan *worker
}

func (p *workerPool) run(fn func(ctx context.Context, data []byte, contentType string, accountID string)) (*Media, error) {

	m := &Media{}

	go func() {
		// take a worker from the worker pool
		worker := <-p.workerQueue
		// tell it to work
		worker.work(fn)
	}()

	return m, nil
}

type worker struct {
	workerQueue chan *worker
	data        []byte
	contentType string
	accountID   string
}

func (w *worker) work(fn func(ctx context.Context, data []byte, contentType string, accountID string)) {
	// return self to pool when finished
	defer w.finish()
	// do the work
	fn(context.Background(), w.data, w.contentType, w.accountID)
}

func (w *worker) finish() {
	// clear self
	w.data = []byte{}
	w.contentType = ""
	w.accountID = ""
	// put self back in the worker pool
	w.workerQueue <- w
}
