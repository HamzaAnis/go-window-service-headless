package main

import (
	"fmt"
)

func getUniqueNodes1(a []string, b []string) {
	result := make([]string, 0, 11)
	for _, v := range a {
		exist := false
		for _, w := range b {
			if v == w {
				exist = true
			}
		}
		if !exist {
			result = append(result, v)
		}
	}
	fmt.Println(result) // [F5 F7 C6 G5]
}
func main() {

	y := []string{"c", "o", "d", "x"}
	z := []string{"c", "l", "m", "d", "a"}
	getUniqueNodes1(y, z)
}
