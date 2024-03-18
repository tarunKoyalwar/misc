package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	gzip "github.com/klauspost/pgzip"

	"github.com/schollz/progressbar/v3"
)

var (
	input        string
	output       string
	isGzip       bool
	delimPrefix  string
	excludeDelim bool
	swapFile     string
)

func main() {
	flag.StringVar(&input, "input", "", "input file")
	flag.StringVar(&output, "output", "", "output file")
	flag.BoolVar(&isGzip, "gzip", false, "gzip output file")
	flag.StringVar(&delimPrefix, "delimPrefix", "", "Prefix of delimiter line until where input needs to be swapped")
	flag.BoolVar(&excludeDelim, "excludeDelim", false, "Exclude delimiter line from input")
	flag.StringVar(&swapFile, "swapFile", "", "file to swap with input")

	flag.Parse()

	if input == "" || output == "" || swapFile == "" {
		flag.PrintDefaults()
		return
	}
	if delimPrefix == "" {
		fmt.Println("delimPrefix is required")
		flag.PrintDefaults()
		return
	}

	// Read input file
	inputFile, err := os.Open(input)
	if err != nil {
		panic(err)
	}
	inputReader := bufio.NewReader(inputFile)

	// Read swap file
	swapped, err := os.ReadFile(swapFile)
	if err != nil {
		panic(err)
	}

	// output file
	outputFile, err := os.Create(output)
	if err != nil {
		panic(err)
	}

	var pb *progressbar.ProgressBar

	// if it is gzip file wrap it with gzip reader
	if isGzip {
		gzReader, err := gzip.NewReader(inputReader)
		if err != nil {
			panic(err)
		}
		inputReader = bufio.NewReader(gzReader)
		pb = progressbar.New(-1)
	} else {
		info, _ := inputFile.Stat()
		pb = progressbar.New64(info.Size())
	}

	outputWriter := io.MultiWriter(outputFile, pb)

	useCopy := false

	fmt.Printf("[+] Reading input file [%s]\n", input)

	read := 0

	now := time.Now()
	// Read input file line by line
	for {
		line, isPrefix, err := inputReader.ReadLine()
		read += len(line)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		if strings.HasPrefix(string(line), delimPrefix) {
			// commit all data from swap file
			_, err := outputWriter.Write(swapped)
			if err != nil {
				panic(err)
			}
			// we have reached the delimiter line
			if !excludeDelim {
				// write delimiter line to output file
				_, err := outputWriter.Write(line)
				if err != nil {
					panic(err)
				}
			}
			_, _ = outputWriter.Write([]byte("\n"))
			// if we already reached delim line then we can use copy
			useCopy = true
			break
		}

		if isPrefix {
			fmt.Printf("Line is prefix\n")
			// commit all data from swap file
			_, err := outputWriter.Write(swapped)
			if err != nil {
				panic(err)
			}
			useCopy = true
			break
			// this is small line so after this point it will be copied instead of read line by line
		}
	}

	if useCopy {
		// copy rest of the input file to output file
		_, err := io.Copy(outputWriter, inputReader)
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("\n\n[+] Done, Time taken %s\n", time.Since(now).String())
}
