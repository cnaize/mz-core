package core

import (
	"github.com/cnaize/mz-common/log"
	"github.com/cnaize/mz-common/model"
	"github.com/cnaize/mz-common/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleSearchMedia(c *gin.Context) {
	db := s.daemon.DB
	user := c.MustGet("user").(model.User)

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
		log.Warn("Server: media search failed: %+v", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	inRequest.RawText = util.DecodeInStr(util.ParseInStr(inRequest.Text))

	mediaList, err := db.SearchMedia(model.MediaAccessTypePrivate, inRequest, in.Offset, in.Count)
	if err != nil {
		if db.IsMediaItemNotFound(err) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		log.Error("Server: media search failed: %+v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var res model.SearchResponseList
	for _, m := range mediaList.Items {
		res.Items = append(res.Items, &model.SearchResponse{
			Owner: user,
			Media: *m,
		})
	}
	res.AllItemsCount = mediaList.AllItemsCount

	c.JSON(http.StatusOK, res)
}

func (s *Server) handleRefreshMedia(c *gin.Context) {
	db := s.daemon.DB

	if err := s.daemon.RefreshMediaDB(); err != nil {
		if db.IsMediaItemNotFound(err) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		log.Error("Server: media db refresh failed: %+v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusAccepted, nil)
}

func (s *Server) handleGetMediaRootList(c *gin.Context) {
	db := s.daemon.DB

	res, err := db.GetMediaRootList()
	if err != nil {
		if db.IsMediaItemNotFound(err) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		log.Error("Server: media root list get failed: %+v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (s *Server) handleAddMediaRoot(c *gin.Context) {
	db := s.daemon.DB

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
	db := s.daemon.DB

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		log.Warn("Server: media root remove failed: %+v", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	root := model.MediaRoot{Base: model.Base{ID: uint(id)}}
	if err := db.RemoveMediaRoot(root); err != nil {
		if db.IsMediaItemNotFound(err) {
			log.Warn("Server: media root remove failed: %+v", err)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		log.Error("Server: media root remove failed: %+v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, nil)
}
