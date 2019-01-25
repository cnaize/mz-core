package core

import (
	"fmt"
	"github.com/cnaize/mz-common/log"
	"github.com/cnaize/mz-core/core/daemon"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	config Config
	daemon *daemon.Daemon
	router *gin.Engine
}

func New(config Config) *Server {
	r := gin.Default()
	r.Use(cors.Default())

	s := &Server{
		config: config,
		daemon: daemon.New(config.Daemon),
		router: r,
	}

	r.GET("/", s.handleStatus)
	v1 := r.Group("/v1")
	{
		media := v1.Group("/media")
		{
			media.POST("/refresh", s.handleRefreshMedia)
			media.GET("/search/:text", s.handleSearchMedia)

			roots := media.Group("/roots")
			{
				roots.GET("", s.handleGetMediaRootList)
				roots.POST("", s.handleAddMediaRoot)
				roots.DELETE("/:dir", s.handleRemoveMediaRoot)
			}
		}
	}

	return s
}

func (s *Server) Run() error {
	if err := s.daemon.Run(); err != nil {
		return fmt.Errorf("run failed: %+v", err)
	}

	log.Info("MuzeZone Core: running server on port: %d", s.config.Port)
	return s.router.Run(fmt.Sprintf(":%d", s.config.Port))
}
