package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/AudioAddict/go-echoprint/echoprint"
)

func dieOrNah(err error) {
	if err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Println(err)
	os.Exit(1)
}

var codegenPath = flag.String("path", "", "path to codegen file to match")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}

	flag.Parse()
	if *codegenPath == "" {
		flag.Usage()
	}

	codegenList, err := echoprint.ParseCodegenFile(*codegenPath)
	dieOrNah(err)

	err = echoprint.DBConnect()
	dieOrNah(err)
	defer echoprint.DBDisconnect()

	allMatches := echoprint.MatchAll(codegenList)

	for group, matches := range allMatches {
		log.Println("Matches for group ", group)
		for _, match := range matches {
			log.Printf("\t%+v", match)
		}
	}
}
