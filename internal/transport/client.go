package transport

import (
	"net/http"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
)

// httpClient wraps a pub.HttpClient to provide rate limited
// access to the underlying client via a concurrency.WorkQueue.
type httpClient struct {
	client pub.HttpClient
	queue  concurrency.WorkQueue
}

// wrapClient returns an httpClient{} using given underlying client, and work queue of size max.
func wrapClient(client pub.HttpClient, max int) pub.HttpClient {
	return &httpClient{
		client: client,
		queue:  concurrency.NewWorkQueue(max),
	}
}

func (c *httpClient) Do(req *http.Request) (rsp *http.Response, err error) {
	c.queue.Run(func() {
		rsp, err = c.client.Do(req)
	})
	return
}
