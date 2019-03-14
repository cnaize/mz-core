package model

const (
	MaxResponseItemsCount           = 200
	MaxRequestItemsPerRequestCount  = 100
	MaxResponseItemsPerRequestCount = MaxResponseItemsCount / 10
)

type SearchRequest struct {
	Base
	Text    string `json:"text" gorm:"unique_index" form:"text" binding:"required"`
	RawText string `json:"-"`
}

type SearchRequestList struct {
	Items         []*SearchRequest `json:"items" form:"items"`
	AllItemsCount *uint            `json:"allItemsCount,omitempty" form:"allItemsCount"`
}

type SearchResponse struct {
	Base
	Owner           User  `json:"owner,omitempty" form:"owner" binding:"required"`
	Media           Media `json:"media,omitempty" form:"media" binding:"required"`
	UserID          uint  `json:"-"`
	MediaID         uint  `json:"-"`
	SearchRequestID uint  `json:"-"`
}

type SearchResponseList struct {
	Items         []*SearchResponse `json:"items" form:"items" binding:"required,dive"`
	AllItemsCount *uint             `json:"allItemsCount,omitempty" form:"allItemsCount"`
}
