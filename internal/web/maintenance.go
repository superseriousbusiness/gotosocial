package web

import (
	"net/http"
	"time"

	"codeberg.org/gruf/go-cache/v3"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api/health"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

type MaintenanceModule struct {
	eTagCache cache.Cache[string, eTagCacheEntry]
}

// NewMaintenance returns a module that routes only
// static assets, and returns a code 503 maintenance
// message template to all other requests.
func NewMaintenance() *MaintenanceModule {
	return &MaintenanceModule{
		eTagCache: newETagCache(),
	}
}

// ETagCache implements withETagCache.
func (m *MaintenanceModule) ETagCache() cache.Cache[string, eTagCacheEntry] {
	return m.eTagCache
}

func (m *MaintenanceModule) Route(r *router.Router, mi ...gin.HandlerFunc) {
	// Route static assets.
	routeAssets(m, r, mi...)

	// Serve OK in response to live
	// requests, but not ready requests.
	liveHandler := func(c *gin.Context) {
		c.Status(http.StatusOK)
	}
	r.AttachHandler(http.MethodGet, health.LivePath, liveHandler)
	r.AttachHandler(http.MethodHead, health.LivePath, liveHandler)

	// For everything else, serve maintenance template.
	obj := map[string]string{"host": config.GetHost()}
	r.AttachNoRouteHandler(func(c *gin.Context) {
		retryAfter := time.Now().Add(120 * time.Second).UTC()
		c.Writer.Header().Add("Retry-After", "120")
		c.Writer.Header().Add("Retry-After", retryAfter.Format(http.TimeFormat))
		c.Header("Cache-Control", "no-store")
		c.HTML(http.StatusServiceUnavailable, "maintenance.tmpl", obj)
	})
}
