package core

import (
	"fmt"
	"github.com/cnaize/mz-common/log"
	"github.com/cnaize/mz-common/model"
	"github.com/cnaize/mz-core/core/daemon"
	"github.com/cnaize/mz-core/db"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

type Server struct {
	config Config
	daemon *daemon.Daemon
	router *gin.Engine
}

func New(config Config, db db.DB) *Server {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:    []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowAllOrigins: true,
		MaxAge:          12 * time.Hour,
	}))

	s := &Server{
		config: config,
		daemon: daemon.New(config.Daemon, db),
		router: r,
	}

	r.GET("/", s.handleStatus)
	v1 := r.Group("/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("/set", s.handleSetUser)
		}

		media := v1.Group("/media", s.handleCheckUser)
		{
			media.POST("/refresh", s.handleRefreshMedia)

			roots := media.Group("/roots")
			{
				roots.GET("", s.handleGetMediaRootList)
				roots.POST("", s.handleAddMediaRoot)
				roots.DELETE("/:id", s.handleRemoveMediaRoot)
			}
		}
	}

	return s
}

func (s *Server) Run() error {
	if err := s.daemon.Run(); err != nil {
		return fmt.Errorf("run failed: %+v", err)
	}

	// TODO: REMOVE IT!!!
	s.daemon.StartFeedLoop(model.User{
		Username: "ni",
		//Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VybmFtZSI6Im5pIn0.E04Xxz7ROycss7bo8mGQ8BHZd4_lGIbAc4H9wlXTAIY",
		Token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VybmFtZSI6Im5pIn0.NGN7R9p3RYVjcBnTiRlFUqzGig1UL5aOMFhCQCCpHQY",
	})

	log.Info("MuzeZone Core: running server on port: %d", s.config.Port)
	return s.router.Run(fmt.Sprintf(":%d", s.config.Port))
}
