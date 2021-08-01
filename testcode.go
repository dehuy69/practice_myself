package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"os/exec"
)

func headerLocate(h264FilePackets []byte) (int, int){
	firstHeaderLocation := 0
	lastHeaderLocation := 0
	// Find first header  0 0 0 0 1 or 0 0 0 1
	for i := 0; i < len(h264FilePackets); i++ {
		if h264FilePackets[i] == 0 {
			res := bytes.Compare(h264FilePackets[i:i+5], []byte{0x00, 0x00, 0x00, 0x00, 0x01})
			res2 := bytes.Compare(h264FilePackets[i:i+4], []byte{0x00, 0x00, 0x00, 0x01})
			// Found header
			if res == 0 || res2 == 0 {
				// fmt.Println(i)
				firstHeaderLocation = i
				break
			}
		}
	}
	fmt.Println("here")
	// Find last header  0 0 0 0 1 or 0 0 0 1
	for i := len(h264FilePackets) - 4; i >= 0; i-- {
		
		if h264FilePackets[i] == 0 {
			fmt.Println(i)
			fmt.Println(h264FilePackets[i:i+1])
			// res := bytes.Compare(h264FilePackets[i:i+5], []byte{0x00, 0x00, 0x00, 0x00, 0x01})
			// Find 0001 first, if appear then check 00001
			res2 := bytes.Compare(h264FilePackets[i:i+4], []byte{0x00, 0x00, 0x00, 0x01})
			// Found header
			if res2 == 0 {
				res := bytes.Compare(h264FilePackets[i-1:i+4], []byte{0x00, 0x00, 0x00, 0x00, 0x01})				
				if res == 0 {lastHeaderLocation = i-1} else {lastHeaderLocation = i}
				break
			}
		}
	}
	return firstHeaderLocation, lastHeaderLocation
}

func jsontest() {
	// jsonMenu := {"Name": "Ed", "Text": "Knock knock."}
	jsonStream := `{"MenuName":"sendVideo", "Param":"h264.mp4"}`
	// fmt.Print(jsonStream)
	type Request struct {
		MenuName, Param string
	}
	dec := json.NewDecoder(strings.NewReader(jsonStream))
	for {
		var m Request
		if err := dec.Decode(&m); err == io.EOF {
			// log.Fatal(err)
			break
		} else if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: %s\n", m.MenuName, m.Param)
	}
}


func testexec() {

    cmd := exec.Command("firefox")

    err := cmd.Run()

    if err != nil {
        log.Fatal(err)
    }
}
func main() {
	// testexec()
	prevB := []byte{}
	res := bytes.Compare(prevB, []byte{})
	fmt.Print(res)
	fmt.Print([]byte("\n"))
}
