package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"example.com/signal"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/h264reader"
)

func main() {
	// Open MP4 file on disk.
	f, _ := os.Open("./h264.mp4")
	// Read entire mp4 into byte slice.
	freader := bufio.NewReader(f)
	content, _ := ioutil.ReadAll(freader)
	// Encode as base64.
	encoded := base64.StdEncoding.EncodeToString(content)
	f.Close()
	content = []byte{}

	// fmt.Println("len base64 mp4: ", len(encoded))

	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
			{
				URLs:           []string{"turn:aivisvn.ddns.net:3478?transport=udp"},
				Username:       "test",
				Credential:     "test",
				CredentialType: webrtc.ICECredentialTypePassword,
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
	buffer := []byte{}
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
		// buffer := []byte{}
		for {
			if isLock == 1 {
				fmt.Print("lockr")
				time.Sleep(10 * time.Millisecond)
				continue
			}
			// In nhieu qua, ko thay chu
			if len(buffer) > 300000 {
				fmt.Print("is seeding ", len(buffer))
			}
			b, err := reader.ReadBytes(END)
			if err != nil {
				log.Panicln(err)
				continue
			}
			// check if h.buffer over 500kb, remove old
			if len(buffer) > 500000 {
				buffer = buffer[len(b):]
			}
			buffer = append(buffer, b...)

		}
	}()

	// THREAD 2
	// Lock seed, then get data from h264FilePackets, empty h264FilePackets, then unlock
	go func() {
		// Wait for connection established
		<-iceConnectedCtx.Done()
		time.Sleep(100 * time.Millisecond) // start after seed thread above
		for {
			// create memory file base on h264File
			// h264FilePackets := []byte{}
			// Lock and copy to h264FilePackets, wait for thread above lock

			// isLock = 1
			// time.Sleep(5 * time.Millisecond)
			h264FilePackets := buffer
			buffer = []byte{}
			// isLock = 0 // unlock

			// Drop until find next header  0 0 0 0 1 or 0 0 0 1
			for i := 0; i < len(h264FilePackets); i++ {
				if h264FilePackets[i] == 0 {
					res := bytes.Compare(h264FilePackets[i:i+5], []byte{0x00, 0x00, 0x00, 0x00, 0x01})
					res2 := bytes.Compare(h264FilePackets[i:i+4], []byte{0x00, 0x00, 0x00, 0x01})
					// Found header
					if res == 0 || res2 == 0 {
						h264FilePackets = h264FilePackets[i:]
						if res == 0 {
							//[1:] because only 0 0 0 1 valid, and 0 0 0 0 1 is invalid
							h264FilePackets = h264FilePackets[1:]
						}
						break
					}
				}
			}

			//// -------------------FINISH process NEXT BEGIN ----------

			// fmt.Println(h264FilePackets)
			// f264, _ := os.OpenFile("h264File.h264",
			// 	os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			// defer f264.Close()
			// f264.Write(h264FilePackets)
			fmt.Println("make memory file from packet len: ", len(h264FilePackets))
			h264File := bytes.NewReader(h264FilePackets)
			h264, ivfErr := h264reader.NewReader(h264File)
			if ivfErr != nil {
				panic(ivfErr)
			}

			spsAndPpsCache := []byte{} // serve logic loop below
			for {
				nal, h264Err := h264.NextNAL()
				if h264Err == io.EOF {
					log.Println("All video frames parsed and sent")
					// runtime.GC()
					break
				} else if h264Err != nil {
					log.Panicln(h264Err)
					break
				}

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
				time.Sleep(time.Millisecond * 33)
			}
		}
	}()

	// Register data channel creation handling
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		fmt.Printf("New DataChannel %s %d\n", d.Label(), d.ID())

		// Register channel opening handling
		d.OnOpen(func() {
			fmt.Printf("Data channel '%s'-'%d' open. Random messages will now be sent to any connected DataChannels every 5 seconds\n", d.Label(), d.ID())
			// Send the message as text
			sendErr := d.SendText("data:video/mp4;base64,")
			if sendErr != nil {
				panic(sendErr)
			}
			for i := 0; i <= len(encoded); i = i + 65535 {
				// Send the message as text
				if i < len(encoded)-65535 {
					d.SendText(encoded[i : i+65535])
				} else {
					d.SendText(encoded[i:])
				}
			}
			sendErr = d.SendText("\n")
			encoded = ""
			if sendErr != nil {
				panic(sendErr)
			}
		})

		// Register text message handling
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Printf("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
		})
	})

	// Block forever
	select {}
}
