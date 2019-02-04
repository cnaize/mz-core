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

	var in struct {
		Offset uint `form:"offset"`
		Count  uint `form:"count"`
	}

	c.ShouldBindQuery(&in)
	if in.Count == 0 || in.Count >= model.MaxResponseItemsCount {
		in.Count = model.MaxResponseItemsCount
	}

	// TODO: come back here again after sign up implementation
	var inRequest model.SearchRequest
	if err := c.ShouldBindQuery(&inRequest); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, model.SearchResponseList{
			Error: &model.Error{Str: fmt.Sprintf("input parse failed: %+v", err)},
		})
		return
	}
	inRequest.RawText = util.DecodeInStr(util.ParseInStr(inRequest.Text))

	mediaList, err := db.SearchMedia(inRequest, in.Offset, in.Count)
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

	var res model.SearchResponseList
	for _, m := range mediaList.Items {
		res.Items = append(res.Items, &model.SearchResponse{
			Owner: *self,
			Media: *m,
		})
	}
	res.AllItemsCount = mediaList.AllItemsCount

	c.JSON(http.StatusOK, res)
}

func (s *Server) handleRefreshMedia(c *gin.Context) {
	if err := s.daemon.RefreshMediaDB(); err != nil {
		if s.config.Daemon.DB.IsMediaItemNotFound(err) {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		log.Warn("Server: media db refresh failed: %+v", err)
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
		log.Warn("Server: media root add failed: %+v", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if !strings.HasSuffix(inRoot.Dir, "/") {
		inRoot.Dir += "/"
	}

	if err := db.AddMediaRoot(inRoot); err != nil {
		log.Warn("Server: media root add failed: %+v", err)
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	c.JSON(http.StatusCreated, nil)
}

func (s *Server) handleRemoveMediaRoot(c *gin.Context) {
	db := s.config.Daemon.DB

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		log.Warn("Server: media root remove failed: %+v", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	root := model.MediaRoot{Base: model.Base{ID: uint(id)}}
	if err := db.RemoveMediaRoot(root); err != nil {
		log.Warn("Server: media root remove failed: %+v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, nil)
}
