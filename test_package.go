package testpack

import (
	"fmt"
	"time"
)

func Firstfunc() {
	x := 1

	fmt.Println("this is import func")
	for {
		fmt.Println("this is loop 2")
		x++
		if x == 3 {
			break
		}
		time.Sleep(1)
	}

}
