package main

import (
	"io"
	"math/rand"
	"os"
	"time"
)

const overwriteCount = 3

func main() {

	rand.Seed(time.Now().Unix())
	Shred("./test1.txt")
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
