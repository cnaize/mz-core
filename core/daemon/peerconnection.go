package daemon

import (
	"fmt"
	"github.com/cnaize/mz-common/model"
	"github.com/cnaize/mz-core/core/daemon/internal"
	"github.com/pions/webrtc"
	"path/filepath"
)

type peerConnection struct {
	media model.Media
	root  model.MediaRoot
	pc    *webrtc.RTCPeerConnection
	track *webrtc.RTCTrack
	done  chan struct{}
}

func newPeerConnection(media model.Media, root model.MediaRoot, pc *webrtc.RTCPeerConnection, track *webrtc.RTCTrack) *peerConnection {
	return &peerConnection{
		media: media,
		root:  root,
		pc:    pc,
		track: track,
		done:  make(chan struct{}),
	}
}

func (p *peerConnection) stream() {
	defer p.pc.Close()

	// REMOVE ME!!!
	p.media.Name = "play"

	mediaPath := fmt.Sprintf("%s.%s",
		filepath.Join(p.root.Dir, p.media.Dir, p.media.Name), string(p.media.Ext))

	gst.CreatePipeline(webrtc.Opus, p.track.Samples,
		fmt.Sprintf("uridecodebin uri=file://%s ! audioconvert ! audioresample ! queue", mediaPath)).
		Start()

	select {}

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
