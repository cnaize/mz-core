package core

import (
	"fmt"
	"github.com/cnaize/mz-common/model"
	"github.com/cnaize/mz-common/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func (s *Server) handleSearchMedia(c *gin.Context) {
	db := s.config.Daemon.DB

	text := util.DecodeInStr(util.ParseInStr(c.Param("text")))

	res, err := db.SearchMedia(text)
	if err != nil {
		if db.IsMediaItemNotFound(err) {
			c.AbortWithStatusJSON(http.StatusOK, model.MediaList{
				Error: &model.Error{Str: fmt.Sprintf("not found: %+v", err)},
			})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, model.MediaList{
			Error: &model.Error{Str: fmt.Sprintf("db failed: %+v", err)},
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

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
		if db.IsMediaItemNotFound(err) {
			c.AbortWithStatusJSON(http.StatusOK, model.MediaRootList{
				Error: &model.Error{Str: fmt.Sprintf("not found: %+v", err)},
			})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, model.MediaRootList{
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

	if rootList, err := db.GetMediaRootList(); err == nil {
		var errStr string
		for _, r := range rootList.Items {
			inDir := strings.ToLower(inRoot.Dir)
			dir := strings.ToLower(r.Dir)
			if strings.HasPrefix(inDir, dir) {
				errStr = fmt.Sprintf("path %s containes existent path %s", dir, inDir)
				break
			}
			if strings.HasPrefix(dir, inDir) {
				errStr = fmt.Sprintf("path %s already containes path %s", inDir, dir)
				break
			}
		}

		if len(errStr) > 0 {
			c.AbortWithStatusJSON(http.StatusConflict, model.MediaRoot{
				Error: &model.Error{Str: errStr},
			})
			return
		}
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

	dir := util.DecodeInStr(c.Param("dir"))

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
