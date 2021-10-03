/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"

	"github.com/dop251/goja"
)

// Module implements the api.ClientModule interface for web pages.
type Module struct {
	config    *config.Config
	processor processing.Processor
	log       *logrus.Logger
	jsVM      *goja.Runtime
}

// New returns a new api.ClientModule for web pages.
func New(config *config.Config, processor processing.Processor, log *logrus.Logger) api.ClientModule {
	return &Module{
		config:    config,
		log:       log,
		processor: processor,
	}
}

func (m *Module) baseHandler(c *gin.Context) {
	l := m.log.WithField("func", "BaseGETHandler")
	l.Trace("serving index html")

	instance, err := m.processor.InstanceGet(c.Request.Context(), m.config.Host)
	if err != nil {
		l.Debugf("error getting instance from processor: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"instance": instance,
	})
}

func (m *Module) reactTest(c *gin.Context) {
	l := m.log.WithField("func", "ReactTestHandler")
	l.Trace("rendering")

	_ = m.jsVM.Set("timestamp", time.Now().String())

	val, _ := m.jsVM.RunString("ReactDOMServer.renderToString(reactGo({context: timestamp}))")

	fmt.Printf("output: %s\n", val.Export().(string))

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(val.Export().(string)))
}

// NotFoundHandler serves a 404 html page instead of a blank 404 error.
func (m *Module) NotFoundHandler(c *gin.Context) {
	l := m.log.WithField("func", "404")
	l.Trace("serving 404 html")

	instance, err := m.processor.InstanceGet(c.Request.Context(), m.config.Host)
	if err != nil {
		l.Debugf("error getting instance from processor: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.HTML(404, "404.tmpl", gin.H{
		"instance": instance,
	})
}

// Route satisfies the RESTAPIModule interface
func (m *Module) Route(s router.Router) error {
	l := m.log.WithField("func", "Web-router")

	// serve static files from /assets
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %s", err)
	}
	assetPath := filepath.Join(cwd, m.config.TemplateConfig.AssetBaseDir)
	s.AttachStaticFS("/assets", fileSystem{http.Dir(assetPath)})

	// Admin panel route, if it exists
	adminPath := filepath.Join(cwd, m.config.TemplateConfig.AssetBaseDir, "/admin")
	s.AttachStaticFS("/admin", fileSystem{http.Dir(adminPath)})

	// serve front-page
	s.AttachHandler(http.MethodGet, "/", m.baseHandler)

	m.jsVM = goja.New()

	if _, err = m.jsVM.RunString("var self = {};"); err != nil {
		l.Errorf("Error with Goja jsVM: %s", err)
	}

	for _, script := range []string{"react.production.min.js", "react-dom-server.min.js"} {
		text, err := ioutil.ReadFile(filepath.Join(assetPath, script))
		if err != nil {
			l.Errorf("Unable to load web asset %s: %s", script, err)
		}
		_, err = m.jsVM.RunScript(script, string(text))
		if err != nil {
			l.Errorf("Unable to load web asset %s: %s", script, err)
		}
	}

	if _, err = m.jsVM.RunString("var React = self.React; var ReactDOMServer = self.ReactDOMServer;"); err != nil {
		l.Errorf("Error with Goja jsVM: %s", err)
	}

	template, err := ioutil.ReadFile(filepath.Join(assetPath, "bundle.js"))
	if err != nil {
		l.Errorf("Unable to load web asset bundle.js: %s", err)
	}
	if _, err = m.jsVM.RunScript("bundle.js", string(template)); err != nil {
		l.Errorf("Error with Goja jsVM: %s", err)
	}

	if _, err = m.jsVM.RunString("var reactGo = self.reactGo"); err != nil {
		l.Errorf("Error with Goja jsVM: %s", err)
	}

	val, err := m.jsVM.RunString(" 'React: ' + React.version + ', ReactDOMServer: ' + ReactDOMServer.version ")
	if err != nil {
		l.Errorf("Error with Goja jsVM: %v", err.(*goja.Exception).String())
	} else {
		fmt.Printf("loaded libraries: %s\n", val.Export().(string))
	}

	s.AttachHandler(http.MethodGet, "/react", m.reactTest)

	// serve statuses
	s.AttachHandler(http.MethodGet, "/:user/statuses/:id", m.threadTemplateHandler)

	// 404 handler
	s.AttachNoRouteHandler(m.NotFoundHandler)

	return nil
}
