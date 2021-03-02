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

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gotosocial/gotosocial/internal/db"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
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
		panic(err)
	}

	// catch shutdown signals from the operating system
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	sig := <-sigs
	log.Infof("received signal %s, shutting down", sig)

	// close down all running services in order
	if err := dbService.Stop(ctx); err != nil {
		log.Errorf("error closing dbservice: %s", err)
	}

	log.Info("done! exiting...")
}
