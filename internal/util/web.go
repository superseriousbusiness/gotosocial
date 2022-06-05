package util

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// NotFoundHandler serves a 404 html page through the provided gin context.
// It calls the provided InstanceGet function to fetch the apimodel
// representation of the instance, for serving in the 404 header and footer.
// If an error is returned by InstanceGet, the function will panic.
func NotFoundHandler(c *gin.Context, InstanceGet func(ctx context.Context, domain string) (*apimodel.Instance, gtserror.WithCode)) {
	host := config.GetHost()
	instance, err := InstanceGet(c.Request.Context(), host)
	if err != nil {
		panic(err)
	}

	c.HTML(http.StatusNotFound, "404.tmpl", gin.H{
		"instance": instance,
	})
}

func ErrorHandler(c *gin.Context, errWithCode gtserror.WithCode, InstanceGet func(ctx context.Context, domain string) (*apimodel.Instance, gtserror.WithCode)) {
	path := c.Request.URL.Path
	if raw := c.Request.URL.RawQuery; raw != "" {
		path = path + "?" + raw
	}

	l := logrus.WithFields(logrus.Fields{
		"path":  path,
		"error": errWithCode.Error(),
	})

	l.Trace("handling error")
	switch errWithCode.Code() {
	case http.StatusNotFound:
		// if we panic for any reason during the 404
		// we should still try to return a basic code
		defer func() {
			if r := recover(); r != nil {
				c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
			}
		}()
		NotFoundHandler(c, InstanceGet)
	case http.StatusInternalServerError:
		l.Error(errWithCode.Safe())
	default:
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
	}
}
