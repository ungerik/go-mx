package main

import (
	"flag"
)

var (
	help    bool
	verbose bool
)

func printHelp() {
	flag.PrintDefaults()
}

func main() {
	flag.BoolVar(&help, "h", false, "show this help")
	flag.BoolVar(&help, "help", false, "show this help")
	flag.BoolVar(&verbose, "v", false, "verbose output")
	flag.Parse()

	if help || len(flag.Args()) == 0 || flag.Arg(0) == "help" {
		printHelp()
		return
	}

	cmd := flag.Arg(0)
	switch cmd {
	case "gen-routes":
		println("gen-routes")
	case "jsx-to-mx":
		println("jsx-to-mx")
	case "tsx-to-mx":
		println("tsx-to-mx")
	}
}
