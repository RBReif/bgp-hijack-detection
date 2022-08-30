package main

import (
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"os"
	"strings"
)

func getRightScanner(file string) (*bufio.Scanner, error) {
	var scanner *bufio.Scanner

	//read inpput
	f, err := os.Open(file)
	if err != nil {
		return scanner, err
	}
	if strings.HasSuffix(file, "bz2") {
		bzip2Reader := bzip2.NewReader(f)
		scanner = bufio.NewScanner(bzip2Reader)
	} else {
		if strings.HasSuffix(file, "gz") {
			gzipReader, err := gzip.NewReader(f)
			if err != nil {
				fmt.Println("Could not open gz file")
				return scanner, err
			}
			scanner = bufio.NewScanner(gzipReader)
		} else {
			scanner = bufio.NewScanner(f)
		}
	}
	return scanner, nil
}
