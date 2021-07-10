package main

import (
	"fmt"
	"example.com/sub"
)

const s string = "constant"

func submain() {
	// fmt.Println("Go" + "lang")
	// fmt.Println("1+1=", 1+1)
	// fmt.Println("7/3=", 7/3)
	// fmt.Println(true && false)
	// fmt.Println(true || false)

	var a = "initial"
	fmt.Println(a)
	var b, c int = 1, 2
	fmt.Println(b, c)
	var e int
	fmt.Println(e)
	f := "apple"
	fmt.Println(f)
	testimport()
	sub.Testimport2()
	controler_thread1 = 0
	thread1()
}
