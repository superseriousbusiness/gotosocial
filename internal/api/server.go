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

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gotosocial/gotosocial/internal/config"
	"github.com/sirupsen/logrus"
)

type Server interface {
	AttachHTTPHandler(method string, path string, handler http.HandlerFunc)
	AttachGinHandler(method string, path string, handler gin.HandlerFunc)
	// AttachMiddleware(handler gin.HandlerFunc)
	GetAPIGroup() *gin.RouterGroup
	Start()
	Stop()
}

type AddsRoutes interface {
	AddRoutes(s Server) error
}

type server struct {
	APIGroup *gin.RouterGroup
	logger   *logrus.Logger
	engine   *gin.Engine
}

func (s *server) GetAPIGroup() *gin.RouterGroup {
	return s.APIGroup
}

func (s *server) Start() {
	// todo: start gracefully
	if err := s.engine.Run(); err != nil {
		s.logger.Panicf("server error: %s", err)
	}
}

func (s *server) Stop() {
	// todo: shut down gracefully
}

func (s *server) AttachHTTPHandler(method string, path string, handler http.HandlerFunc) {
	s.engine.Handle(method, path, gin.WrapH(handler))
}

func (s *server) AttachGinHandler(method string, path string, handler gin.HandlerFunc) {
	s.engine.Handle(method, path, handler)
}

func New(config *config.Config, logger *logrus.Logger) Server {
	engine := gin.New()
	return &server{
		APIGroup: engine.Group("/api").Group("/v1"),
		logger:   logger,
		engine:   engine,
	}
}
