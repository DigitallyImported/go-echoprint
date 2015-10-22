package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"

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

	runtime.GOMAXPROCS(runtime.NumCPU())

	fpList, err := echoprint.ParseCodegenFile(os.Args[1])
	dieOrNah(err)

	mm, err := echoprint.NewMatchMaker()
	dieOrNah(err)

	var wg sync.WaitGroup
	for _, codegenFp := range fpList {
		wg.Add(1)

		go func(codegenFp echoprint.CodegenFp) {
			defer wg.Done()

			log.Printf("Processing codegen TrackID=%d, Version=%f, Filename=%s\n",
				codegenFp.TrackID, codegenFp.Metadata.Version, codegenFp.Metadata.Filename)

			fp, err := echoprint.NewFingerprint(codegenFp.Code, codegenFp.Metadata.Version)
			dieOrNah(err)

			// log.Printf("%d Code/Time pairs\n", len(fp.Codes))

			matches, err := mm.Match(fp)
			dieOrNah(err)

			log.Println("Number of matches found:", len(matches))
		}(codegenFp)
		//dieOrNah(err)
	}

	wg.Wait()
}
