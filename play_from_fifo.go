package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	// "io/ioutil"
	"log"
	"os"
	"time"

	"example.com/signal"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/h264reader"
)

func main() {
	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())

	// Create a video track
	videoTrack, videoTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	if videoTrackErr != nil {
		panic(videoTrackErr)
	}

	rtpSender, videoTrackErr := peerConnection.AddTrack(videoTrack)
	if videoTrackErr != nil {
		panic(videoTrackErr)
	}

	// Read incoming RTCP packets
	// Before these packets are returned they are processed by interceptors. For things
	// like NACK this needs to be called.
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			iceConnectedCtxCancel()
		}
	})

	// Wait for the offer to be pasted
	offer := webrtc.SessionDescription{}
	signal.Decode(signal.MustReadStdin(), &offer)

	// Set the remote SessionDescription
	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		panic(err)
	}

	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Connect to FIFO pipe, continuos append, when enough length then send to video track
	h264FilePackets := []byte{}
	go func() {
		var pipeFile = "./MYFIFO"
		const END = '\n'
		fmt.Println("open a named pipe file for read.")
		f, err := os.OpenFile(pipeFile, os.O_RDONLY, os.ModeNamedPipe)
		if err != nil {
			log.Panicln(err)
		}
		reader := bufio.NewReader(f)
		// Wait for connection established
		<-iceConnectedCtx.Done()

		f264, err := os.OpenFile("h264File.h264",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
		}
		defer f264.Close()
		for {
			b, _ := reader.ReadBytes(END)
			if err != nil {
				log.Panicln(err)
				// continue
			}
			// case data only one charecter > pass out
			// if len(b) < 2 {
			// 	continue
			// }
			// drop last charecter \n
			// b = b[:len(b)-1]

			// log.Printf("receive : %v \n", b)
			// fmt.Println("")

			h264FilePackets = append(h264FilePackets, b...)
			if _, err := f264.Write(b); err != nil {
				log.Println(err)
			}
			// fmt.Println("len package: ", len(h264FilePackets))
			if len(h264FilePackets) < 200000 {
				continue
			}

			// create memory file base on h264File
			h264File := bytes.NewReader(h264FilePackets)

			// h264File, ivfErr := os.Open("h264File.h264")
			// if ivfErr != nil {
			// 	panic(ivfErr)
			// }
			//////////////////////////////////////////////////////

			h264, ivfErr := h264reader.NewReader(h264File)
			if ivfErr != nil {
				panic(ivfErr)
			}
			h264FilePackets = []byte{}

			// Send our video file frame at a time. Pace our sending so we send it at the same speed it should be played back as.
			// This isn't required since the video is timestamped, but we will such much higher loss if we send all at once.
			// sleepTime := time.Millisecond * time.Duration((float32(header.TimebaseNumerator)/float32(header.TimebaseDenominator))*1000)
			go func() {
				spsAndPpsCache := []byte{}
				for {
					nal, h264Err := h264.NextNAL()
					if h264Err == io.EOF {
						fmt.Printf("All video frames parsed and sent")
						break
					}

					if h264Err != nil {
						panic(h264Err)
					}

					time.Sleep(time.Millisecond * 33)
					nal.Data = append([]byte{0x00, 0x00, 0x00, 0x01}, nal.Data...)
					if nal.UnitType == h264reader.NalUnitTypeSPS || nal.UnitType == h264reader.NalUnitTypePPS {
						spsAndPpsCache = append(spsAndPpsCache, nal.Data...)
						continue
					} else if nal.UnitType == h264reader.NalUnitTypeCodedSliceIdr {
						nal.Data = append(spsAndPpsCache, nal.Data...)
						spsAndPpsCache = []byte{}
					}
					// fmt.Println(nal.Data)
					if h264Err = videoTrack.WriteSample(media.Sample{Data: nal.Data, Duration: time.Second}); ivfErr != nil {
						panic(ivfErr)
					}
				}
			}()

		}
	}()

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		panic(err)
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	// Output the answer in base64 so we can paste it in browser
	fmt.Println(signal.Encode(*peerConnection.LocalDescription()))

	// Block forever
	select {}
}
