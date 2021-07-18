package main

import (
	"fmt"
	"example.com/sub"
	"time"

	"./testpack"
)

const s string = "constant"

func main() {
	// fmt.Println("Go" + "lang")
	// fmt.Println("1+1=", 1+1)
	// fmt.Println("7/3=", 7/3)
	// fmt.Println(true && false)
	// fmt.Println(true || false)
	// go func() {
    //     fmt.Println("goroutin 1")
	// 	go func ()  {
	// 		fmt.Println("goroutin 2")
	// 	}()
    // }()

	// var a = "initial"
	// fmt.Println(a)
	// var b, c int = 1, 2
	// fmt.Println(b, c)
	// var e int
	// fmt.Println(e)
	// f := "apple"
	// fmt.Println(f)
	// // testimport()
	// sub.Testimport2()
	// time.Sleep(1 * time.Millisecond)
	// // controler_thread1 = 0
	// // thread1()
	testpack.Firstfunc()

}
