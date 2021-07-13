package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"runtime"

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

	// Seeding thread:
	// Connect to FIFO pipe, continuos append,
	h264FilePackets := []byte{}
	p_h264FilePackets := &h264FilePackets

	var pipeFile = "./MYFIFO"

	// Wait for connection established
	<-iceConnectedCtx.Done()

	fmt.Println("open a named pipe file for read.")
	VIDEO_STREAM_PIPE, err := os.OpenFile(pipeFile, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		log.Panicln(err)
	}
	defer VIDEO_STREAM_PIPE.Close()
	reader := bufio.NewReader(VIDEO_STREAM_PIPE)
	const END = '\n'
	isLock := 0 // unlock at at begin
	go func() {

		for {
			if isLock == 1 {
				fmt.Print("lockr")
				time.Sleep(10 * time.Millisecond)
				continue
			}
			fmt.Print("is seeding ", len(*p_h264FilePackets))
			b, err := reader.ReadBytes(END)
			if err != nil {
				log.Panicln(err)
				continue
			}
			*p_h264FilePackets = append(*p_h264FilePackets, b...)

		}
	}()

	// THREAD 2
	// Lock seed, then get data from h264FilePackets, empty h264FilePackets, then unlock
	go func() {
		// Wait for connection established
		<-iceConnectedCtx.Done()
		time.Sleep(1000 * time.Millisecond) // start after seed thread above
		for {
			// create memory file base on h264File
			// if len too small, wait until it larger
			if len(*p_h264FilePackets) < 1000 {
				continue
			}
			// Lock and process h264FilePackets, wait for thread above lock
			isLock = 1
			time.Sleep(10 * time.Millisecond)

			// if packet does not begin with 00 00 00 01 , then seek to header, take data from that seek
			// res := bytes.Compare(h264FilePackets[:3], []byte{0x00, 0x00, 0x00, 0x01})
			// if res != 0 {
			// 	h264FilePackets = append([]byte{0x00, 0x00, 0x00, 0x01}, h264FilePackets...)
			// }

			// Get next packet in named pipe until next begin file header 00 00 00 01
			fmt.Println("GET next packet")
			reader = bufio.NewReader(VIDEO_STREAM_PIPE)
			nextHeader := []byte{}
			for {
				nextHeader, _ = reader.ReadBytes(END)
				fmt.Print(nextHeader)
				if len(nextHeader) <= 4 {
					h264FilePackets = append(*p_h264FilePackets, nextHeader...)
					continue
				}
				res := bytes.Compare(nextHeader[:5], []byte{0x00, 0x00, 0x00, 0x00, 0x01})
				if res != 0 {
					h264FilePackets = append(*p_h264FilePackets, nextHeader...)
					continue
				}
				break // appear next header frame
			}
			//// -------------------FINISH process NEXT BEGIN ----------

			h264File := bytes.NewReader(*p_h264FilePackets)
			h264, ivfErr := h264reader.NewReader(h264File)
			if ivfErr != nil {
				panic(ivfErr)
			}
			// add nextFrame to public var h264FilePackets,
			//[1:] because only 0 0 0 1 valid, and 0 0 0 0 1 is invalid 
			*p_h264FilePackets = nextHeader[1:]
			isLock = 0 // unlock

			spsAndPpsCache := []byte{} // serve logic loop below
			for {
				nal, h264Err := h264.NextNAL()
				if h264Err == io.EOF {
					fmt.Printf("All video frames parsed and sent")
					break
				} else if h264Err != nil {
					log.Panicln(h264Err)
					// Track file error
					// flog, _ := os.Create("panic.log")
					// defer flog.Close()
					// h264File.WriteTo(flog)
					os.Exit(0)
				}
				time.Sleep(time.Millisecond * 25)
				// fmt.Println(nal.Data)

				nal.Data = append([]byte{0x00, 0x00, 0x00, 0x01}, nal.Data...)
				if nal.UnitType == h264reader.NalUnitTypeSPS || nal.UnitType == h264reader.NalUnitTypePPS {
					spsAndPpsCache = append(spsAndPpsCache, nal.Data...)
					continue
				} else if nal.UnitType == h264reader.NalUnitTypeCodedSliceIdr {
					nal.Data = append(spsAndPpsCache, nal.Data...)
					spsAndPpsCache = []byte{}
				}
				if h264Err = videoTrack.WriteSample(media.Sample{Data: nal.Data, Duration: time.Second}); ivfErr != nil {
					panic(ivfErr)
				}
			}
		}
	}()

	// Block forever
	select {}
}
