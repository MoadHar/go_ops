package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	useProd = flag.Bool("prod", false, "Use a production endpoint")
	useDev  = flag.Bool("dev", false, "Use developement endpoint")
	help    = flag.Bool("help", false, "display help text")
)

func main() {
	fmt.Println("in b/b.go")
	flag.Parse()
	authors := flag.Args()
	fmt.Println(authors)
	if *help {
		flag.PrintDefaults()
		return
	}
	switch {
	case *useDev && *useProd:
		log.Println("Error: --prod and --dev cannot be both set")
		flag.PrintDefaults()
		os.Exit(1)
	case !(*useProd || *useDev):
		log.Println("Error: either --prod or --dev must be set")
		flag.PrintDefaults()
		os.Exit(1)
	}
}
