package main

import (
	"bufio"
	"bytes"
	// "container/list"
	"fmt"
	"log"
	"os"
	// "time"
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
	// nalQueue := list.New()

	// go func () {
	// 	for {
	// 	if (nalQueue.Len() > 0) {
	// 		e := nalQueue.Front() // First element
	// 		fmt.Println("from queue : %v \n", e.Value)
		
	// 		nalQueue.Remove(e) // Dequeue
	// 	} else {
	// 		fmt.Println("queue empty")
	// 		time.Sleep(1 * time.Second)
	// 		continue
	// 	}
	// }
	// }()

	fmt.Println("open a named pipe file for read.")
	f, err := os.OpenFile(pipeFile, os.O_RDONLY, os.ModeNamedPipe)
	// f, err := os.Open(pipeFile)
	if err != nil {
		log.Panicln(err)
	}

	// reader := bufio.NewReader(f)
	oneNAL := []byte{}
	// GOP := []interface{}{"one", oneNAL}
	for {
		// fmt.Println("waiting...")
		b, err := bufio.NewReader(f).ReadBytes(END)
		// b := bufio.NewScanner(f).Bytes()
		if err != nil {
			log.Panicln(err)
			// continue
		}
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
		} else {
			log.Printf("send : %v \n", oneNAL)
			// nalQueue.PushBack(oneNAL)		
			log.Println("")
			oneNAL = b
		}
		


		// log.Printf("read : %v \n", b)		
		// log.Println("")

		// fmt.Printf("read : %v \n", len(b))		
		// fmt.Println("success next..")
		// fmt.Println(string(b))
		// time.Sleep(4 * time.Second)
	}
}
