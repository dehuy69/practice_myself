package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"
)

var pipeFile = "./MYFIFO"
const END = '\n'

func main() {
	// os.Remove(pipeFile)
	// err := syscall.Mkfifo(pipeFile, 0666)
	// if err != nil {
	// 	log.Fatal("Make named pipe file error:", err)
	// }
	// go scheduleWrite()
	// syscall.Mkfifo(pipeFile, 0777)
	fmt.Println("open a named pipe file for read.")
	f, err := os.OpenFile(pipeFile, os.O_RDONLY, os.ModeNamedPipe)
	fmt.Println("here 1")
	// f, err := os.Open(pipeFile)
	if err != nil {
		log.Panicln(err)
	}

	// reader := bufio.NewReader(f)

	for {
		fmt.Println("waiting...")
		time.Sleep(500 * time.Millisecond)
		b, err := bufio.NewReader(f).ReadBytes(END)
		// b := bufio.NewScanner(f).Bytes()
		if err != nil {
			log.Panicln(err)
			// continue
		}
		if (len(b) < 2) {
			continue
		} 
		b = b[:len(b)-1]

		log.Printf("read : %v \n", b)		
		log.Println("")

		// fmt.Printf("read : %v \n", len(b))		
		// fmt.Println("success next..")
		// fmt.Println(string(b))
		// time.Sleep(4 * time.Second)
	}
}
