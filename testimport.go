package main

import (
	"fmt"
	"time"
)

func testimport() {
	x := 1
	go func ()  {
		fmt.Println("this is import func")
		for {
			fmt.Println("this is loop 2")
			x++
			if (x == 3) {break}
			time.Sleep(1)
		}
		
	}()

}

func main() {
	testimport()
	fmt.Println("this is main func")
	time.Sleep(10*time.Millisecond)
}