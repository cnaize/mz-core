package core

import (
	"github.com/cnaize/mz-common/log"
	"github.com/cnaize/mz-common/model"
	"github.com/cnaize/mz-common/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Server) handleSetUser(c *gin.Context) {
	var inUser model.User
	if err := c.ShouldBindJSON(&inUser); err != nil {
		log.Warn("Server: user set failed: %+v", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	xtoken := util.DecodeInStr(c.Query("xtoken"))
	if xtoken == "" {
		log.Warn("Server: xtoken set failed: empty token")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	inUser.XToken = xtoken
	s.config.Daemon.CurrentUser = &inUser

	c.Status(http.StatusAccepted)
}

func (s *Server) handleSetCurrentUser(c *gin.Context) {
	if s.config.Daemon.CurrentUser == nil {
		log.Warn("Server: current user set failed: user is nil")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Set("currentUser", s.config.Daemon.CurrentUser)
}
