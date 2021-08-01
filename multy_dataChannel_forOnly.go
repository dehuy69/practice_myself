package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"

	"example.com/signal"
	"github.com/pion/webrtc/v3"
)

func main() {
	// Open file on disk.
	f, _ := os.Open("./h264.mp4")

	// Read entire JPG into byte slice.
	reader := bufio.NewReader(f)
	content, _ := ioutil.ReadAll(reader)

	// Encode as base64.
	encoded := base64.StdEncoding.EncodeToString(content)
	fmt.Println("len base64 mp4: ", len(encoded))
	// Maximum datachannel send each time is 65535

	// Everything below is the Pion WebRTC API! Thanks for using it ❤️.

	sdpChan := signal.HTTPSDPServer()
	for {
		
		fmt.Println("")
		fmt.Println("Curl an base64 SDP to start sendonly peer connection")

		recvOnlyOffer := webrtc.SessionDescription{}
		signal.Decode(<-sdpChan, &recvOnlyOffer)

		// Create a new PeerConnection
		peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
				// {
				// 	URLs:           []string{"turn:aivisvn.ddns.net:3478?transport=udp"},
				// 	Username:       "test",
				// 	Credential:     "test",
				// 	CredentialType: webrtc.ICECredentialTypePassword,
				// },
			},
		})

		if err != nil {
			panic(err)
		}
		fmt.Println("here1")
		// Set the handler for Peer connection state
		// This will notify you when the peer has connected/disconnected
		peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
			fmt.Printf("Peer Connection State has changed: %s\n", s.String())

			if s == webrtc.PeerConnectionStateFailed {
				// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
				// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
				// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
				fmt.Println("Peer Connection has gone to failed exiting")
			}
		})
		fmt.Println("here2")
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
				if sendErr != nil {
					panic(sendErr)
				}
			})

			// Register text message handling
			d.OnMessage(func(msg webrtc.DataChannelMessage) {
				fmt.Printf("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
			})
		})
		fmt.Println("here3")

		fmt.Println("here4")
		// Set the remote SessionDescription
		err = peerConnection.SetRemoteDescription(recvOnlyOffer)
		if err != nil {
			panic(err)
		}
		fmt.Println("here5")
		// Create an answer
		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			panic(err)
		}
		fmt.Println("here6")
		// Create channel that is blocked until ICE Gathering is complete
		gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
		fmt.Println("here7")
		// Sets the LocalDescription, and starts our UDP listeners
		err = peerConnection.SetLocalDescription(answer)
		if err != nil {
			panic(err)
		}
		fmt.Println("here8")
		// Block until ICE Gathering is complete, disabling trickle ICE
		// we do this because we only can exchange one signaling message
		// in a production application you should exchange ICE Candidates via OnICECandidate
		<-gatherComplete
		fmt.Println("here9")
		// Get the LocalDescription and take it to base64 so we can paste in browser
		fmt.Println(signal.Encode(*peerConnection.LocalDescription()))
	}
	// Block forever
	// select {}
}
