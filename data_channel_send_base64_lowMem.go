package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	// "log"
	"os"

	"example.com/signal"
	"github.com/pion/webrtc/v3"
)

func main() {
	// Everything below is the Pion WebRTC API! Thanks for using it ❤️.

	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}
	defer func() {
		if cErr := peerConnection.Close(); cErr != nil {
			fmt.Printf("cannot close peerConnection: %v\n", cErr)
		}
	}()

	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Peer Connection State has changed: %s\n", s.String())

		if s == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			fmt.Println("Peer Connection has gone to failed exiting")
			os.Exit(0)
		}
	})

	// Register data channel creation handling
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		fmt.Printf("New DataChannel %s %d\n", d.Label(), d.ID())

		// Register channel opening handling
		d.OnOpen(func() {
			iterReadAndSend(d)
		})

		// Register text message handling
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Printf("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
		})
	})

	// Wait for the offer to be pasted
	offer := webrtc.SessionDescription{}
	signal.Decode(signal.MustReadStdin(), &offer)

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	// Create an answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
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

func readAllAndSend(d *webrtc.DataChannel) {
	// Open file on disk.
	f, _ := os.Open("./h264.mp4")

	// Read entire JPG into byte slice.
	reader := bufio.NewReader(f)
	content, _ := ioutil.ReadAll(reader)

	// Encode as base64.
	encoded := base64.StdEncoding.EncodeToString(content)
	fmt.Println("len base64 mp4: ", len(encoded))
	// Maximum datachannel send each time is 65535
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
}

func iterReadAndSend(d *webrtc.DataChannel) {
	// Open file on disk.
	file, _ := os.Open("./h264.mp4")
	data := make([]byte, 3000)
	totalLengthBase64 := 0

	sendErr := d.SendText("data:video/mp4;base64,")
	if sendErr != nil {
		panic(sendErr)
	}
	for {
		_, err := file.Read(data)
		if err != nil {
			fmt.Println(totalLengthBase64)
			// log.Fatal(err)
			break
		}
		// Encode as base64.
		videoEncoded := base64.StdEncoding.EncodeToString(data)
		d.SendText(videoEncoded)
	}
	fmt.Println("herer")
	sendErr = d.SendText("\n")
	if sendErr != nil {
		panic(sendErr)
	}

}
