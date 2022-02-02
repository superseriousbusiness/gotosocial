/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package testrig

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

// CreateTestContextWithTemplatesAndSessions calls gin.CreateTestContext and then configures the sessions and templates similarly to how the router does
func CreateTestContextWithTemplatesAndSessions(w http.ResponseWriter) (engine *gin.Engine) {

	engine = gin.New()

	auth, _ := base64.StdEncoding.DecodeString("b6vrpSW3V5HkDBZaNU08DaCG15WeCWNvWf21l6AczDo=")
	crypt, _ := base64.StdEncoding.DecodeString("72k9LuzqnPgDjHultGJS+xkcRCw+Ekl6ZugIVh99jGs=")

	store := memstore.NewStore(auth, crypt)
	store.Options(router.SessionOptions())

	//sessionName, err := router.SessionName()
	sessionName := "gotosocial-localhost"

	engine.Use(sessions.Sessions(sessionName, store))

	router.LoadTemplateFunctions(engine)

	// does not work because CWD is messed up while tests are running
	// // load templates onto the engine
	// if err := router.LoadTemplates(engine); err != nil {
	// 	panic(err)
	// }

	// https://stackoverflow.com/questions/31873396/is-it-possible-to-get-the-current-root-of-package-structure-as-a-string-in-golan
	_, runtimeCallerLocation, _, _ := runtime.Caller(0)
	projectRoot, err := filepath.Abs(filepath.Join(filepath.Dir(runtimeCallerLocation), "../"))
	if err != nil {
		panic(err)
	}

	templateBaseDir := viper.GetString(config.Keys.WebTemplateBaseDir)

	_, err = os.Stat(filepath.Join(projectRoot, templateBaseDir, "index.tmpl"))
	if err != nil {
		panic(fmt.Errorf("%s doesn't seem to contain the templates; index.tmpl is missing: %s", err))
	}

	tmPath := filepath.Join(projectRoot, fmt.Sprintf("%s*", templateBaseDir))
	engine.LoadHTMLGlob(tmPath)

	return engine
}
