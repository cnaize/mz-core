package daemon

import (
	"encoding/json"
	"fmt"
	"github.com/cnaize/mz-common/log"
	"github.com/cnaize/mz-common/model"
	"github.com/cnaize/mz-common/util"
	"github.com/pion/webrtc/v2"
	"math/rand"
)

type peerConnection struct {
	media model.Media
	root  model.MediaRoot
	pc    *webrtc.PeerConnection
	done  chan struct{}
}

func newPeerConnection(media model.Media, root model.MediaRoot) *peerConnection {
	return &peerConnection{
		media: media,
		root:  root,
		done:  make(chan struct{}, 1),
	}
}

func (p *peerConnection) Open(inRTCKey string) (string, error) {
	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
			{URLs: []string{"stun:stun1.l.google.com:19302"}},
			{URLs: []string{"stun:stun2.l.google.com:19302"}},
			{URLs: []string{"stun:stun3.l.google.com:19302"}},
			{URLs: []string{"stun:stun4.l.google.com:19302"}},
		},
	}

	// Create a new RTCPeerConnection
	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {

		return "", err
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	pc.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		if connectionState == webrtc.ICEConnectionStateConnected {
			go p.stream()
		}

		if connectionState == webrtc.ICEConnectionStateCompleted ||
			connectionState == webrtc.ICEConnectionStateFailed ||
			connectionState == webrtc.ICEConnectionStateDisconnected ||
			connectionState == webrtc.ICEConnectionStateClosed {
			p.Done()
		}
		log.Info("Daemon: peer connection state changed: %s", connectionState.String())
	})

	// Create a audio track
	r := rand.Uint32()
	audioTrack, err := pc.NewTrack(webrtc.DefaultPayloadTypeOpus, r, "audio", fmt.Sprintf("pion-%d", r))
	if err != nil {
		return "", err
	}
	_, err = pc.AddTrack(audioTrack)
	if err != nil {
		return "", err
	}

	// Wait for the offer to be pasted
	var offer webrtc.SessionDescription
	if err := json.Unmarshal([]byte(util.DecodeInStr(inRTCKey)), &offer); err != nil {
		return "", err
	}
	// Set the remote SessionDescription
	if err := pc.SetRemoteDescription(offer); err != nil {
		return "", err
	}

	// Sets the LocalDescription, and starts our UDP listeners
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		return "", err
	}
	outRTCKey, err := json.Marshal(&answer)
	if err != nil {
		return "", err
	}

	return util.EncodeOutStr(string(outRTCKey)), nil
}

func (p *peerConnection) Done() {
	p.done <- struct{}{}
}

func (p *peerConnection) stream() {
	defer p.pc.Close()

	// REMOVE ME!!!
	p.media.Name = "play"

	//mediaPath := fmt.Sprintf("%s.%s",
	//	filepath.Join(p.root.Dir, p.media.Dir, p.media.Name), string(p.media.Ext))



/*	gst.CreatePipeline(webrtc.Opus, p.track.Samples,
		fmt.Sprintf("uridecodebin uri=file://%s ! audioconvert ! audioresample ! queue", mediaPath)).
		Start()

	select {}
*/

/*	mediaPath := fmt.Sprintf("%s.%s",
		filepath.Join(p.root.Dir, p.media.Dir, p.media.Name), string(p.media.Ext))

	f, err := os.Open(mediaPath)
	if err != nil {
		log.Error("Peer connection: media stream failed: %+v", err)
		return
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Error("Peer connection: media stream failed: %+v", err)
		return
	}
	defer streamer.Close()

	defer log.Info("Peer connection: audio stream finished")

	tickD := time.Second / 100

	bufferSize := format.SampleRate.N(tickD)
	samples := make([][2]float64, bufferSize)
	buf := make([]byte, bufferSize*4)

	for {
		select {
		case <-p.done:
			return
		default:
			n, _ := streamer.Stream(samples)
			//if n == len(samples) && ok {
			//	log.Debug("Case 1")
			//} else if 0 < n && n < len(samples) && ok {
			//	log.Debug("Case 2")
			//} else if n == 0 && !ok {
			//	log.Debug("Case 3")
			//}
			if streamer.Err() != nil {
				log.Warn("Peer connection: stream failed: %+v", streamer.Err())
				p.done <- struct{}{}
				return
			}

			fillSamplesBuf(samples, buf)

			p.track.Samples <- media.RTCSample{
				Data:    buf,
				Samples: uint32(n),
			}

			//log.Debug("Peer connection: samples sent: %d", n)
		}
	}
*/
}

func fillSamplesBuf(samples [][2]float64, buf []byte) {
	for i := range samples {
		for c := range samples[i] {
			val := samples[i][c]
			if val < -1 {
				val = -1
			}
			if val > +1 {
				val = +1
			}
			valInt16 := int16(val * (1<<15 - 1))
			low := byte(valInt16)
			high := byte(valInt16 >> 8)
			buf[i*4+c*2+0] = low
			buf[i*4+c*2+1] = high
		}
	}
}
