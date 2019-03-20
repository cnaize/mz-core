package model

type MediaExt string

const (
	MediaExtUnknown MediaExt = "unknown"
	MediaExtMP3     MediaExt = "mp3"
)

var SupportedMediaExtList = []MediaExt{MediaExtMP3}

type MediaAccessType string

const (
	MediaAccessTypeUnknown   MediaAccessType = "unknown"
	MediaAccessTypePrivate   MediaAccessType = "private"
	MediaAccessTypeProtected MediaAccessType = "protected"
	MediaAccessTypePublic    MediaAccessType = "public"
)

type Media struct {
	Base
	Name        string   `json:"name" form:"name" binding:"required"`
	Ext         MediaExt `json:"ext" form:"ext" binding:"required"`
	Dir         string   `json:"dir" form:"dir"`
	CoreSideID  uint     `json:"coreSideID" form:"name" binding:"required"`
	MediaRootID uint     `json:"rootID" form:"rootID" binding:"required"`
	RawPath     string   `json:"-" form:"-"` // lower path, used for search
}

type MediaList struct {
	Items         []*Media `json:"items" form:"items"`
	AllItemsCount *uint    `json:"allItemsCount,omitempty" form:"allItemsCount"`
}

type MediaRoot struct {
	Base
	Dir        string          `json:"dir" form:"dir" binding:"required" gorm:"unique_index"`
	AccessType MediaAccessType `json:"accessType" form:"accessType" binding:"required"`
	MediaCount *uint           `json:"mediaCount,omitempty" form:"mediaCount"`
}

type MediaRootList struct {
	Items []*MediaRoot `json:"items" form:"items"`
}

type MediaRequest struct {
	Base
	User      User            `json:"user" form:"-"`
	Owner     User            `json:"owner" form:"owner" binding:"required,dive"`
	Media     Media           `json:"media" form:"media" binding:"required,dive"`
	Mode      MediaAccessType `json:"mode" form:"mode"`
	WebRTCKey string          `json:"webRTCKey" form:"webRTCKey" binding:"required"`
	UserID    uint            `json:"-" form:"-"`
	OwnerID   uint            `json:"-" form:"-"`
	MediaID   uint            `json:"-" form:"-"`
}

type MediaRequestList struct {
	Items []*MediaRequest `json:"items" form:"items" binding:"required,dive"`
}

type MediaResponse struct {
	Base
	User      User            `json:"user" form:"user" binding:"required,dive"`
	Owner     User            `json:"owner" form:"-"`
	Media     Media           `json:"media" form:"media" binding:"required,dive"`
	Mode      MediaAccessType `json:"mode" form:"mode"`
	WebRTCKey string          `json:"webRTCKey" form:"webRTCKey" binding:"required"`
	Error     *Error          `json:"error,omitempty" form:"error" binding:"dive"`
	UserID    uint            `json:"-" form:"-"`
	OwnerID   uint            `json:"-" form:"-"`
	MediaID   uint            `json:"-" form:"-"`
}

type MediaResponseList struct {
	Items []*MediaResponse `json:"items" form:"items" binding:"required,dive"`
}
