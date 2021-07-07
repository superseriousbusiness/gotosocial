package router

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// loadTemplates loads html templates for use by the given engine
func loadTemplates(cfg *config.Config, engine *gin.Engine) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %s", err)
	}

	tmPath := filepath.Join(cwd, fmt.Sprintf("%s*", cfg.TemplateConfig.BaseDir))

	engine.LoadHTMLGlob(tmPath)
	return nil
}
