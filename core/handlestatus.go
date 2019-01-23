package core

import (
	"github.com/cnaize/mz-common/model"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Server) handleStatus(c *gin.Context) {
	c.JSON(http.StatusOK, model.CoreStatus{
		Version: s.config.Daemon.Version,
	})
}
