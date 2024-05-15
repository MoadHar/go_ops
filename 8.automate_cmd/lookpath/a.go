package main

import (
	"fmt"
	"os/exec"
)

const (
	kubectl = "kubectl"
	git     = "git"
)

func main() {
	a, err := exec.LookPath(kubectl)
	if err != nil {
		fmt.Println("cannot find kubectl in PATH: ", err)
	} else {
		fmt.Println(a)
	}

	b, err := exec.LookPath(git)
	if err != nil {
		fmt.Println("no GIT: ", err)
	} else {
		fmt.Println(b)
	}
}
