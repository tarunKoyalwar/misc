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
	TypeGzip bool
	input    string
	output   string
	decode   bool
	peek     int
)

func main() {
	flag.StringVar(&input, "input", "", "input file")
	flag.StringVar(&output, "output", "", "output file")
	flag.BoolVar(&TypeGzip, "gzip", false, "gzip output file")
	flag.BoolVar(&decode, "decode", false, "decode input file")
	flag.IntVar(&peek, "peek", 0, "dumps first N bytes of the input file(useful when its compressed)")
	flag.Parse()
	if input == "" || (output == "" && peek == 0) {
		flag.PrintDefaults()
		return
	}

	if !TypeGzip {
		// make basic assumptions
		if strings.HasSuffix(input, ".gz") {
			decode = true
			TypeGzip = true
		}

		if strings.HasSuffix(output, ".gz") {
			TypeGzip = true
		}
	}

	if !TypeGzip {
		fmt.Println("Only gzip is supported, currently")
		flag.PrintDefaults()
		return
	}

	inputFile, err := os.Open(input)
	if err != nil {
		panic(err)
	}
	inputReader := bufio.NewReader(inputFile)

	if peek > 0 {
		peekData := make([]byte, peek)
		gzipReader, err := gzip.NewReader(inputReader)
		if err != nil {
			panic(err)
		}
		n, err := gzipReader.Read(peekData)
		if err != nil {
			panic(err)
		}
		fmt.Printf("[+] Peeked %d bytes: \n%s\n", n, string(peekData))
		return
	}

	output, err := os.Create(output)
	if err != nil {
		panic(err)
	}

	pb := progressbar.New(-1)
	outputWriter := bufio.NewWriter(io.MultiWriter(output, pb))

	now := time.Now()

	if decode {
		// decode
		reader, err := gzip.NewReader(inputReader)
		if err != nil {
			panic(err)
		}
		defer reader.Close()
		_, err = io.Copy(outputWriter, reader)
		if err != nil {
			panic(err)
		}
	} else {
		// encode
		writer := gzip.NewWriter(output)
		defer writer.Close()
		_, err = io.Copy(writer, inputReader)
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("[+] Done, Time taken %s\n", time.Since(now).String())
}
