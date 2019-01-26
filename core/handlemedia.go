package core

import (
	"fmt"
	"github.com/cnaize/mz-common/log"
	"github.com/cnaize/mz-common/model"
	"github.com/cnaize/mz-common/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleSearchMedia(c *gin.Context) {
	db := s.config.Daemon.DB
	self := c.MustGet("currentUser").(*model.User)

	var request model.SearchRequest
	// TODO: come back here again after sign up implementation
	if err := c.ShouldBindQuery(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, model.SearchResponseList{
			Error: &model.Error{Str: fmt.Sprintf("input parse failed: %+v", err)},
		})
		return
	}
	request.RawText = util.DecodeInStr(util.ParseInStr(request.Text))

	mediaList, err := db.SearchMedia(request)
	if err != nil {
		if db.IsMediaItemNotFound(err) {
			c.AbortWithStatusJSON(http.StatusOK, model.SearchResponseList{
				Error: &model.Error{Str: fmt.Sprintf("not found: %+v", err)},
			})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, model.SearchResponseList{
			Error: &model.Error{Str: fmt.Sprintf("db failed: %+v", err)},
		})
		return
	}

	res := model.SearchResponseList{
		Request: &request,
	}
	for _, m := range mediaList.Items {
		res.Items = append(res.Items, &model.SearchResponse{
			Owner: self,
			Media: m,
		})
	}

	c.JSON(http.StatusOK, res)
}

func (s *Server) handleRefreshMedia(c *gin.Context) {
	if err := s.daemon.RefreshMediaDB(); err != nil {
		if s.config.Daemon.DB.IsMediaItemNotFound(err) {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		log.Error("Server: media db refresh failed: %+v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusAccepted, nil)
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
		log.Error("Server: media root add failed: %+v", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if !strings.HasSuffix(inRoot.Dir, "/") {
		inRoot.Dir += "/"
	}

	if err := db.AddMediaRoot(inRoot); err != nil {
		log.Error("Server: media root add failed: %+v", err)
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	c.JSON(http.StatusCreated, nil)
}

func (s *Server) handleRemoveMediaRoot(c *gin.Context) {
	db := s.config.Daemon.DB

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		log.Error("Server: media root remove failed: %+v", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	uid := uint(id)

	root := model.MediaRoot{Base: model.Base{ID: &uid}}
	if err := db.RemoveMediaRoot(root); err != nil {
		log.Error("Server: media root remove failed: %+v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, nil)
}
