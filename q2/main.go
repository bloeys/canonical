package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"
)

const overwriteCount = 3

var (
	fileToShred = flag.String("file", "", "./my-file.txt")
)

func main() {

	flag.Parse()
	if *fileToShred == "" {
		fmt.Println("No file specified to shred. Please use the '-file' argument to specify a file")
		return
	}

	rand.Seed(time.Now().Unix())
	err := Shred(*fileToShred)
	if err != nil {
		fmt.Println("Error shredding file. Error: " + err.Error())
	}
}

func Shred(fPath string) error {

	f, err := os.OpenFile(fPath, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	//Find file size
	fileSizeBytes, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	//Write 1MB of data until file is fully written
	const WriteBufSize = 1024 * 1024
	randData := make([]byte, WriteBufSize)

	for i := 0; i < overwriteCount; i++ {

		currLoc, err := f.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}

		for currLoc < fileSizeBytes {

			rand.Read(randData)

			//Ensure that we don't write over the file size
			remBytes := fileSizeBytes - currLoc
			if WriteBufSize > remBytes {
				currLoc += remBytes
				_, err = f.Write(randData[:remBytes])
			} else {
				currLoc += WriteBufSize
				_, err = f.Write(randData)
			}

			if err != nil {
				return err
			}
		}
	}

	f.Close()
	err = os.Remove(fPath)
	if err != nil {
		return err
	}

	return nil
}
