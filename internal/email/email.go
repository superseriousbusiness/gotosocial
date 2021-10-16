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

package email

import (
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"path/filepath"

	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// Sender contains functions for sending emails to instance users/new signups.
type Sender interface {
	// SendConfirmEmail sends a 'please confirm your email' style email to the given toAddress, with the given data.
	SendConfirmEmail(toAddress string, data ConfirmData) error

	// SendResetEmail sends a 'reset your password' style email to the given toAddress, with the given data.
	SendResetEmail(toAddress string, data ResetData) error

	// ExecuteTemplate returns templated HTML using the given templateName and data. Mostly you won't need to call this,
	// and can just call one of the 'Send' functions instead (which calls this under the hood anyway).
	ExecuteTemplate(templateName string, data interface{}) (string, error)
}

// NewSender returns a new email Sender interface with the given configuration, or an error if something goes wrong.
func NewSender(cfg *config.Config) (Sender, error) {
	t, err := loadTemplates(cfg)
	if err != nil {
		return nil, err
	}

	auth := smtp.PlainAuth("", cfg.SMTPConfig.Username, cfg.SMTPConfig.Password, cfg.SMTPConfig.Host)

	return &sender{
		hostAddress: fmt.Sprintf("%s:%d", cfg.SMTPConfig.Host, cfg.SMTPConfig.Port),
		from:        cfg.SMTPConfig.From,
		auth:        auth,
		template:    t,
	}, nil
}

type sender struct {
	hostAddress string
	from        string
	auth        smtp.Auth
	template    *template.Template
}

// loadTemplates loads html templates for use in emails
func loadTemplates(cfg *config.Config) (*template.Template, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current working directory: %s", err)
	}

	// look for all templates that start with 'email_'
	tmPath := filepath.Join(cwd, fmt.Sprintf("%semail_*", cfg.TemplateConfig.BaseDir))
	return template.ParseGlob(tmPath)
}
