package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

const pipeFile = "./MYFIFO"
const END = '\n'
const h264File = "h264File.h264"

func main() {
	fmt.Println("open a named pipe file for read.")
	f, err := os.OpenFile(pipeFile, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		log.Panicln(err)
	}
	reader := bufio.NewReader(f)

	f264, err := os.OpenFile(h264File,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f264.Close()

	for {
		b, err := reader.ReadBytes(END)
		// b := bufio.NewScanner(f).Bytes()
		if err != nil {
			log.Panicln(err)
			// continue
		}

		// if len(b) < 2 {
		// 	continue
		// }
		// b = b[:len(b)-1]

		log.Printf("read : %v \n", b)
		log.Println("")
		
		if _, err := f264.Write(b); err != nil {
			log.Println(err)
		}
	}
}
