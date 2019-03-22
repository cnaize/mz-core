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

type peerConnection struct {

}

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
		log.Debug("Daemon: media request list handle: empty list")
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
		User: request.User,
		Owner: model.User{Username: d.CurrentUser.Username},
		Media: request.Media,
		Mode: request.Mode,
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

	if _, err := db.GetMediaByID(request.Media.CoreSideID); err != nil {
		log.Error("Daemon: media request handle failed: %+v", err)
		response.Error = &model.Error{Str: err.Error()}
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

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	pc.OnICEConnectionStateChange(func(connectionState ice.ConnectionState) {
		if connectionState == ice.ConnectionStateCompleted ||
			connectionState == ice.ConnectionStateFailed ||
			connectionState == ice.ConnectionStateDisconnected ||
			connectionState == ice.ConnectionStateClosed {
			// TODO:
			//  do something?
		}
		log.Info("Daemon: peer connection state changed: %s", connectionState.String())
	})

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