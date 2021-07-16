package main

import (
	"fmt"
	"time"
)


type Stream struct{

}
func (s *Stream) thread1() {    
    for {
		fmt.Println("this is thread 1 of stream")
		time.Sleep(1 * time.Second)
	}
}
func (s *Stream) thread2() {    
	for {
		fmt.Println("this is thread 2 of stream")
		time.Sleep(1 * time.Second)
	}
}
func (s *Stream) Init() {    
    go s.thread1()
	go s.thread2()
}

func main() {
    stream := new(Stream)
    stream.Init()
	select{}
}