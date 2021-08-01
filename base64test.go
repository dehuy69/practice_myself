package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	// Open file on disk.
	f, _ := os.Open("./download.mp4")

	// Read entire JPG into byte slice.
	reader := bufio.NewReader(f)
	content, _ := ioutil.ReadAll(reader)

	// Encode as base64.
	encoded := base64.StdEncoding.EncodeToString(content)
	fmt.Println("len base64 mp4: ", len(encoded))

	for i := 0; i <= len(encoded); i = i + 65535 {
		// Send the message as text
		if i < len(encoded)-65535 {
			// d.SendText(encoded[i : i+65535])
			fmt.Println("send from to: ", i, i+65535)
		} else {
			// d.SendText(encoded[i:])
			fmt.Println("send final from to: ", i, len(encoded))
			fmt.Println("final total: ", len(encoded)-i)
		}
	}

	// fmt.Print(encoded)
}
