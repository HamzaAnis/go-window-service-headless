package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Hello, What's your favourite number?")
	var i int
	fmt.Scanf("%d\n", &i)
	fmt.Println("Ah I like ", i, " too.")
	time.Sleep(6 * time.Second)
}
