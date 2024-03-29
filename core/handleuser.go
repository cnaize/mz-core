package core

import (
	"github.com/cnaize/mz-common/log"
	"github.com/cnaize/mz-common/model"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Server) handleSetUser(c *gin.Context) {
	var in struct {
		Username string `json:"username" form:"username" binding:"required"`
		Token    string `json:"token" form:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&in); err != nil {
		log.Warn("Server: current user set failed: %+v", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	user := model.User{
		Username: in.Username,
		Token:    in.Token,
	}
	log.Debug("Server: Daemon: user set: %s", user.Username)

	if err := s.daemon.StartFeedLoop(user); err != nil {
		log.Error("Server: Daemon: feed loop start failed: %+v", err)
	}

	c.Status(http.StatusAccepted)
}

func (s *Server) handleCheckUser(c *gin.Context) {
	user := s.daemon.CurrentUser

	if user.Username == "" || user.Token == "" {
		log.Warn("Server: user set failed: current user is empty: %#v", user)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Set("user", user)
}
