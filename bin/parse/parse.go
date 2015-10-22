package main

import (
	"fmt"
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

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s [path/to/echoprint.json]\n", os.Args[0])
		os.Exit(1)
	}

	fpList, err := echoprint.ParseCodegenFile(os.Args[1])
	dieOrNah(err)

	for i, codegenFp := range fpList {
		fmt.Printf("Processing codegen %d, TrackID=%d, Version=%f, Filename=%s\n",
			i, codegenFp.TrackID, codegenFp.Metadata.Version, codegenFp.Metadata.Filename)

		fp, err := echoprint.NewFingerprint(fpList[0].Code, fpList[0].Metadata.Version)
		dieOrNah(err)

		fmt.Printf("%d Code/Time pairs", len(fp.Codes))
		// fmt.Print("\tCodes=")
		// for _, code := range fp.Codes {
		// 	fmt.Print(code)
		// 	fmt.Print(" ")
		// }
		//
		// fmt.Print("\n\tTimes=")
		// for _, time := range fp.Times {
		// 	fmt.Print(time)
		// 	fmt.Print(" ")
		// }
		fmt.Print("\n\n")
	}
}
