package middleware

import (
	"github.com/cnaize/mz-common/model"
	"github.com/cnaize/mz-core/db"
	"github.com/gin-gonic/gin"
)

func SetCurrentUser(db db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := "cnaize"
		c.Set("currentUser", &model.User{
			Username: &username,
		})
	}
}
