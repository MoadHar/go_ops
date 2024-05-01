package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("yes go")
	data, err := os.ReadFile("go.mod")
	if err != nil {
		fmt.Println("err reading: ", err)
	}
	fmt.Println(string(data))
}
