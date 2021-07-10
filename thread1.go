package main

import (
	"fmt"
	"time"
)
var controler_thread1 int = 0

func thread1() {
	for true {
		if (controler_thread1 == 1) {
			fmt.Println("thread1 is running")
			time.Sleep(1 * time.Second)
		} else {
			fmt.Println("thread1 is pausing")
			time.Sleep(1 * time.Second)
			fmt.Println("loopnext")
		}
	}
}