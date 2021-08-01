package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	// "bufio"
	// "io/ioutil"
)

func main() {
	FILENAME := "/mnt/testnfs/ev300_code_bin/h264-5mp.mp4"
	file, err := os.Open(FILENAME) // For read access.
	if err != nil {
		log.Fatal(err)
	}

	// // Read entire MP4 into byte slice.
	// reader := bufio.NewReader(file)
	// content, _ := ioutil.ReadAll(reader)

	// // Encode as base64.
	// videoEncoded := base64.StdEncoding.EncodeToString(content)
	// fmt.Println("len of binary: ", len(content))
	// fmt.Println("len of base64: ", len(videoEncoded))

	data := make([]byte, 3000)
	totalLengthBase64 := 0
	for {
		_, err := file.Read(data)
		if err != nil {
			fmt.Println(totalLengthBase64)
			// log.Fatal(err)
			break
		}
		// Encode as base64.
		videoEncoded := base64.StdEncoding.EncodeToString(data)
		totalLengthBase64 += len(videoEncoded)
	}
	fmt.Println("finish loop")
}
