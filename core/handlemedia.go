package core

import (
	"encoding/base64"
	"fmt"
	"github.com/cnaize/mz-common/model"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"net/http"
	"strings"
)

func (s *Server) handleRefreshMedia(c *gin.Context) {
	res, err := s.daemon.RefreshMediaDB()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, model.MediaRootList{
			Error: &model.Error{Str: fmt.Sprintf("media db refresh failed: %+v", err)},
		})
		return
	}

	c.JSON(http.StatusAccepted, res)
}

func (s *Server) handleGetMediaRootList(c *gin.Context) {
	db := s.config.Daemon.DB

	res, err := db.GetMediaRootList()
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.AbortWithStatusJSON(http.StatusOK, model.MediaRootList{
				Error: &model.Error{Str: fmt.Sprintf("not found: %+v", err)},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, model.MediaRootList{
			Error: &model.Error{Str: fmt.Sprintf("db failed: %+v", err)},
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (s *Server) handleAddMediaRoot(c *gin.Context) {
	db := s.config.Daemon.DB

	var inRoot model.MediaRoot
	if err := c.ShouldBindJSON(&inRoot); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, model.MediaRoot{
			Error: &model.Error{Str: fmt.Sprintf("in model parse failed: %+v", err)},
		})
		return
	}
	if !strings.HasSuffix(inRoot.Dir, "/") {
		inRoot.Dir += "/"
	}

	if err := db.AddMediaRoot(&inRoot); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, model.MediaRoot{
			Error: &model.Error{Str: fmt.Sprintf("db failed: %+v", err)},
		})
		return
	}

	c.JSON(http.StatusCreated, inRoot)
}

func (s *Server) handleRemoveMediaRoot(c *gin.Context) {
	db := s.config.Daemon.DB

	dir, err := base64.URLEncoding.DecodeString(c.Param("dir"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, model.MediaRootList{
			Error: &model.Error{Str: fmt.Sprintf("dir parse failed: %+v", err)},
		})
		return
	}

	if err := db.RemoveMediaRoot(&model.MediaRoot{Dir: string(dir)}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, model.MediaRootList{
			Error: &model.Error{Str: fmt.Sprintf("db failed: %+v", err)},
		})
		return
	}

	res, _ := db.GetMediaRootList()
	if res == nil {
		res = &model.MediaRootList{}
	}

	c.JSON(http.StatusOK, res)
}
