package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
)

// var errRE = regexp.MustCompile(`(?i)error`)
// var errRE = regexp.MustCompile(`(?i)google`)
var errRE = regexp.MustCompile(`(?i)inertia`)

//var errRE = regexp.MustCompile(`(?i)unable`)

func main() {
	var s *bufio.Scanner
	switch len(os.Args) {
	case 1:
		log.Printf("no file specified, using STDIN")
		log.Println(os.Args[0])
		s = bufio.NewScanner(os.Stdin)
	case 2:
		f, err := os.Open(os.Args[1])
		if err != nil {
			log.Println(err)
			os.Exit(2)
		}
		s = bufio.NewScanner(f)
	default:
		log.Println("too many arguments provided")
		os.Exit(3)
	}
	for s.Scan() {
		line := s.Bytes()
		println("read stdin lines: ", line)
		if errRE.Match(line) {
			fmt.Printf(">>>> %s\n", line)
		}
	}
	if err := s.Err(); err != nil {
		log.Println("Error: ", err)
		os.Exit(1)
	}
}
