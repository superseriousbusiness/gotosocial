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

package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gotosocial/gotosocial/internal/db"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// getLog will try to set the logrus log level to the
// desired level specified by the user with the --log-level flag
func getLog(c *cli.Context) (*logrus.Logger, error) {
	log := logrus.New()
	logLevel, err := logrus.ParseLevel(c.String("log-level"))
	if err != nil {
		return nil, err
	}
	log.SetLevel(logLevel)
	return log, nil
}

// Run starts the gotosocial server
func Run(c *cli.Context) error {
	log, err := getLog(c)
	if err != nil {
		return fmt.Errorf("error creating logger: %s", err)
	}

	ctx := context.Background()
	dbConfig := &db.Config{
		Type:            "POSTGRES",
		Address:         "",
		Port:            5432,
		User:            "",
		Password:        "whatever",
		Database:        "postgres",
		ApplicationName: "gotosocial",
	}
	dbService, err := db.NewService(ctx, dbConfig, log)
	if err != nil {
		return err
	}

	// catch shutdown signals from the operating system
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	sig := <-sigs
	log.Infof("received signal %s, shutting down", sig)

	// close down all running services in order
	if err := dbService.Stop(ctx); err != nil {
		return fmt.Errorf("error closing dbservice: %s", err)
	}

	log.Info("done! exiting...")
	return nil
}
