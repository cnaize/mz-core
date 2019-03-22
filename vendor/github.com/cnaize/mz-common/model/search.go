package model

const (
	MaxResponseItemsCount           = 200
	MaxRequestItemsPerRequestCount  = 100
	MaxResponseItemsPerRequestCount = MaxResponseItemsCount / 10
)

type SearchRequest struct {
	Base
	Text    string          `json:"text" form:"text" binding:"required"`
	Mode    MediaAccessType `json:"mode" form:"mode"`
	UserID  uint            `json:"-" form:"-"`
	RawText string          `json:"-" form:"-"`
}

type SearchRequestList struct {
	Items []*SearchRequest `json:"items" form:"items"`
}

type SearchResponse struct {
	Base
	Owner           User  `json:"owner,omitempty" form:"owner" binding:"required"`
	Media           Media `json:"media,omitempty" form:"media" binding:"required"`
	UserID          uint  `json:"-" form:"-"` // actually this is owner.ID (done for db.Model().Related())
	MediaID         uint  `json:"-" form:"-"`
	SearchRequestID uint  `json:"-" form:"-"`
}

type SearchResponseList struct {
	Items []*SearchResponse `json:"items" form:"items" binding:"required,dive"`
}
