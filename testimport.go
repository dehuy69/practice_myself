package main

import "fmt"
import "time"

func testimport() {
	fmt.Println("this is import func")
}

func main() {
	for {
		fmt.Println("this is loop 1")
		x := 1
		for {
			fmt.Println("this is loop 2")
			x++
			if (x == 3) {break}
			time.Sleep(1)
		}
	}
}