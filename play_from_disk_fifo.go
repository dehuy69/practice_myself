package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	// "go/printer"
	"io"
	"log"
	"os"
	"time"

	// "container/list"

	"example.com/signal"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"

	// "github.com/pion/webrtc/v3/pkg/media/ivfreader"
	"github.com/pion/webrtc/v3/pkg/media/oggreader"
	// "github.com/pion/webrtc/v3/pkg/media/h264reader"
)

const (
	audioFileName = "output.ogg"
	// videoFileName = "output.ivf"
	videoFileName = "stream_chn1.h264"
)

func main() {
	// Assert that we have an audio or video file
	_, err := os.Stat(videoFileName)
	haveVideoFile := !os.IsNotExist(err)

	_, err = os.Stat(audioFileName)
	haveAudioFile := !os.IsNotExist(err)

	if !haveAudioFile && !haveVideoFile {
		panic("Could not find `" + audioFileName + "` or `" + videoFileName + "`")
	}


	// Connect to FIFO pipe
	GOP := []byte{}
	sendGOP := []byte{}
	go func ()  {
		var pipeFile = "./MYFIFO"
		const END = '\n'
		fmt.Println("open a named pipe file for read.")
		f, err := os.OpenFile(pipeFile, os.O_RDONLY, os.ModeNamedPipe)
		if err != nil {
			log.Panicln(err)
		}
		oneNAL := []byte{}
		for {
			b, _ := bufio.NewReader(f).ReadBytes(END)
			// if err != nil {
			// 	log.Panicln(err)
			// 	// continue
			// }

			// case data only one charecter > pass out
			if (len(b) < 2) {
				continue
			}
			// drop last charecter \n
			b = b[:len(b)-1]

			// if not begin with 0001 > merge
			// else send prev nal, then add new one
			res := bytes.Compare(b[:3], []byte{0x00, 0x00, 0x00, 0x01})
			if (res == 1) {
				oneNAL = append(oneNAL, b...)
				continue
			} else {
				// Add prev NAL to GOP
				GOP = append(GOP, oneNAL...)
				// fmt.Println("len gop: %v", len(GOP))
				if (len(GOP) > 41000) {
					sendGOP = GOP
					GOP = []byte{}
				}
				// Add new data to begin a new NAL
				oneNAL = b
			}
			// log.Printf("read : %v \n", nalData)		
			// log.Println("")
		}	
	}()	

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

	if haveVideoFile {
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

		go func() {
			// Open a h264 file and start reading using our IVFReader
			// file, ivfErr := os.Open(videoFileName)
			// if ivfErr != nil {
			// 	panic(ivfErr)
			// }

			// h264, ivfErr := h264reader.NewReader(file)
			// if ivfErr != nil {
			// 	panic(ivfErr)
			// }

			// Wait for connection established
			<-iceConnectedCtx.Done()

			// Send our video file frame at a time. Pace our sending so we send it at the same speed it should be played back as.
			// This isn't required since the video is timestamped, but we will such much higher loss if we send all at once.
			// sleepTime := time.Millisecond * time.Duration((float32(header.TimebaseNumerator)/float32(header.TimebaseDenominator))*1000)
			
			// spsAndPpsCache := []byte{}
			for {
				// nal, h264Err := h264.NextNAL()
				// if h264Err == io.EOF {
				// 	fmt.Printf("All video frames parsed and sent")
				// 	os.Exit(0)
				// }

				// if h264Err != nil {
				// 	panic(h264Err)
				// }

				// time.Sleep(sleepTime)
				// time.Sleep(time.Millisecond * 33)
				// nal.Data = append([]byte{0x00, 0x00, 0x00, 0x01}, nal.Data...)
				// if nal.UnitType == h264reader.NalUnitTypeSPS || nal.UnitType == h264reader.NalUnitTypePPS {
				// 	spsAndPpsCache = append(spsAndPpsCache, nal.Data...)
				// 	continue
				// } else if nal.UnitType == h264reader.NalUnitTypeCodedSliceIdr {
				// 	nal.Data = append(spsAndPpsCache, nal.Data...)
				// 	spsAndPpsCache = []byte{}
				// }
				// fmt.Println(nal.Data)
				if (len(sendGOP) == 0) {
					
					continue
				}
				fmt.Println(sendGOP)
				videoTrack.WriteSample(media.Sample{Data: sendGOP, Duration: time.Second})
				sendGOP = []byte{}
				// if err = videoTrack.WriteSample(media.Sample{Data: nalData, Duration: time.Second}); err != nil {
				// 	panic(err)
				// }
			}
		}()
	}

	if haveAudioFile {
		// Create a audio track
		audioTrack, audioTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion")
		if audioTrackErr != nil {
			panic(audioTrackErr)
		}

		rtpSender, audioTrackErr := peerConnection.AddTrack(audioTrack)
		if audioTrackErr != nil {
			panic(audioTrackErr)
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

		go func() {
			// Open a IVF file and start reading using our IVFReader
			file, oggErr := os.Open(audioFileName)
			if oggErr != nil {
				panic(oggErr)
			}

			// Open on oggfile in non-checksum mode.
			ogg, _, oggErr := oggreader.NewWith(file)
			if oggErr != nil {
				panic(oggErr)
			}

			// Wait for connection established
			<-iceConnectedCtx.Done()

			// Keep track of last granule, the difference is the amount of samples in the buffer
			var lastGranule uint64
			for {
				pageData, pageHeader, oggErr := ogg.ParseNextPage()
				if oggErr == io.EOF {
					fmt.Printf("All audio pages parsed and sent")
					os.Exit(0)
				}

				if oggErr != nil {
					panic(oggErr)
				}

				// The amount of samples is the difference between the last and current timestamp
				sampleCount := float64(pageHeader.GranulePosition - lastGranule)
				lastGranule = pageHeader.GranulePosition
				sampleDuration := time.Duration((sampleCount/48000)*1000) * time.Millisecond

				if oggErr = audioTrack.WriteSample(media.Sample{Data: pageData, Duration: sampleDuration}); oggErr != nil {
					panic(oggErr)
				}

				time.Sleep(sampleDuration)
			}
		}()
	}

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

	// Block forever
	select {}
}