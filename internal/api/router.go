package api

import "github.com/gin-gonic/gin"

// Router provides the http routes used by the API
type Router interface {
	Route()
}

// NewRouter returns a new router
func NewRouter() Router {
	return &router{}
}

// router implements the router interface
type router struct {

}

func (r *router) Route() {
	ginRouter := gin.Default()
	ginRouter.LoadHTMLGlob("web/template/*")

	apiGroup := ginRouter.Group("/api")
	{
		v1 := apiGroup.Group("/v1")
		{
			statusesGroup := v1.Group("/statuses")
			{
				statusesGroup.GET(":id", statusGet)
			}

		}
	}
	ginRouter.Run()
}
