package daemon

import (
	"fmt"
	"github.com/cnaize/mz-common/log"
	"github.com/cnaize/mz-common/model"
	"net/http"
	"time"
)

func (d *Daemon) mediaFeedLoop() {
	log.Info("Daemon: media feed loop run")
	for {
		d.handleMediaRequestList()
		time.Sleep(1 * time.Second)
	}
}

func (d *Daemon) handleMediaRequestList() {
	req := d.baseReq.Clone()

	var res model.MediaRequestList
	resp, _, errs := req.Get(fmt.Sprintf("%s/v1/media/requests", d.config.CenterBaseURL)).EndStruct(&res)
	if (resp != nil && resp.StatusCode == http.StatusNotFound) || len(res.Items) == 0 {
		return
	}
	if errs != nil {
		log.Warn("Daemon: media request list handle failed: %+v", errs)
		return
	}

	for _, r := range res.Items {
		r := *r

		go d.handleMediaRequest(r)
	}
}

func (d *Daemon) handleMediaRequest(request model.MediaRequest) {
	db := d.DB
	req := d.baseReq.Clone()

	// TODO:
	//  check if we have connections count > limit

	response := model.MediaResponse{
		MediaRequest: request,
	}

	defer func() {
		resp, _, errs := req.Post(fmt.Sprintf("%s/v1/media/responses", d.config.CenterBaseURL)).
			Type("json").
			SendStruct(&response).
			End()
		if (resp != nil && resp.StatusCode != http.StatusCreated) || errs != nil {
			log.Warn("Daemon: media request handle failed: %+v", errs)
			return
		}
	}()

	if request.Owner.Username != d.CurrentUser.Username {
		errStr := fmt.Sprintf("request owner and current user mismatch: %s != %s",
			request.Owner.Username, d.CurrentUser.Username)
		log.Error("Daemon: " + errStr)
		response.Error = &model.Error{Str: errStr}
		return
	}

	// TODO:
	//  handle "protected" mode

	if request.Mode == model.MediaAccessTypePrivate && request.User.Username != d.CurrentUser.Username {
		errStr := fmt.Sprintf("media owner %s: access denied for user %s",
			d.CurrentUser.Username, request.User.Username)
		log.Error("Daemon: " + errStr)
		response.Error = &model.Error{Str: errStr}
		return
	}

	media, err := db.GetMediaByID(request.Media.CoreSideID)
	if err != nil {
		log.Error("Daemon: media request handle failed: %+v", err)
		response.Error = &model.Error{Str: err.Error()}
		return
	}

	mediaRoot, err := db.GetMediaRootByID(media.MediaRootID)
	if err != nil {
		log.Error("Daemon: media request handle failed: %+v", err)
		response.Error = &model.Error{Str: err.Error()}
		return
	}
	if request.Mode != mediaRoot.AccessType {
		errStr := fmt.Sprintf("request mode and media root access type mismatch: %s != %s",
			request.Mode, mediaRoot.AccessType)
		log.Error("Daemon: " + errStr)
		response.Error = &model.Error{Str: errStr}
		return
	}

	pc := newPeerConnection(media, mediaRoot)

	response.WebRTCKey, err = pc.Open(request.WebRTCKey)
	if err != nil {
		log.Error("Daemon: media request handle failed: %+v", err)
		response.Error = &model.Error{Str: err.Error()}
		return
	}
}
