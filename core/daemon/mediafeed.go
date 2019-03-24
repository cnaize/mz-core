package daemon

import (
	"encoding/json"
	"fmt"
	"github.com/cnaize/mz-common/log"
	"github.com/cnaize/mz-common/model"
	"github.com/cnaize/mz-common/util"
	"github.com/pions/webrtc"
	"github.com/pions/webrtc/pkg/ice"
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

	// Setup the codecs you want to use.
	// We'll use the default ones but you can also define your own
	webrtc.RegisterDefaultCodecs()

	// Prepare the configuration
	config := webrtc.RTCConfiguration{
		IceServers: []webrtc.RTCIceServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
			{URLs: []string{"stun:stun1.l.google.com:19302"}},
			{URLs: []string{"stun:stun2.l.google.com:19302"}},
			{URLs: []string{"stun:stun3.l.google.com:19302"}},
			{URLs: []string{"stun:stun4.l.google.com:19302"}},
		},
	}

	// Create a new RTCPeerConnection
	pc, err := webrtc.New(config)
	if err != nil {
		log.Error("Daemon: media request handle failed: %+v", err)
		response.Error = &model.Error{Str: err.Error()}
		return
	}

	// Create a audio track
	opusTrack, err := pc.NewRTCSampleTrack(webrtc.DefaultPayloadTypeOpus, "audio", "pion1")
	if err != nil {
		log.Error("Daemon: media request handle failed: %+v", err)
		response.Error = &model.Error{Str: err.Error()}
		return
	}
	_, err = pc.AddTrack(opusTrack)
	if err != nil {
		log.Error("Daemon: media request handle failed: %+v", err)
		response.Error = &model.Error{Str: err.Error()}
		return
	}

	peer := newPeerConnection(media, mediaRoot, pc, opusTrack)

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	pc.OnICEConnectionStateChange(func(connectionState ice.ConnectionState) {
		if connectionState == ice.ConnectionStateConnected {
			go peer.stream()
		}

		if connectionState == ice.ConnectionStateCompleted ||
			connectionState == ice.ConnectionStateFailed ||
			connectionState == ice.ConnectionStateDisconnected ||
			connectionState == ice.ConnectionStateClosed {
			peer.done <- struct{}{}
		}
		log.Info("Daemon: peer connection state changed: %s", connectionState.String())
	})

	// Wait for the offer to be pasted
	var offer webrtc.RTCSessionDescription
	if err := json.Unmarshal([]byte(util.DecodeInStr(request.WebRTCKey)), &offer); err != nil {
		log.Error("Daemon: media request handle failed: %+v", err)
		response.Error = &model.Error{Str: err.Error()}
		return
	}

	// Set the remote SessionDescription
	if err := pc.SetRemoteDescription(offer); err != nil {
		log.Error("Daemon: media request handle failed: %+v", err)
		response.Error = &model.Error{Str: err.Error()}
		return
	}

	// Sets the LocalDescription, and starts our UDP listeners
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		log.Error("Daemon: media request handle failed: %+v", err)
		response.Error = &model.Error{Str: err.Error()}
		return
	}
	webRTCKey, err := json.Marshal(&answer)
	if err != nil {
		log.Error("Daemon: media request handle failed: %+v", err)
		response.Error = &model.Error{Str: err.Error()}
		return
	}

	response.WebRTCKey = util.EncodeOutStr(string(webRTCKey))
}
