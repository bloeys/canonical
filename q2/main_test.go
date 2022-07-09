package main

import (
	"bytes"
	"io"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestShred(t *testing.T) {

	rand.Seed(time.Now().Unix())

	checkFile := func(fPath string) {

		copyFName, hardlinkFName, fBytes := copyAndHardlink(fPath)
		sizeBeforeShred := int64(len(fBytes))
		defer os.Remove(hardlinkFName)

		err := Shred(copyFName)
		fatalOnErr(t, err)

		//Check that file was deleted
		_, err = os.Open(copyFName)
		if err == nil {
			t.Fatalf("Managed to open file after shred, which means file '%s' wasn't deleted by shred", fPath)
		}

		//The original is deleted but we have access to the hardlink which we
		//can use to analyze how shred did
		tempFile, err := os.Open(hardlinkFName)
		fatalOnErr(t, err)
		defer tempFile.Close()

		sizeAfterShred, err := tempFile.Seek(0, io.SeekEnd)
		fatalOnErr(t, err)

		//Check size hasn't changed
		sizeDiff := sizeAfterShred - sizeBeforeShred
		if sizeDiff != 0 {
			t.Fatalf("File size changed which means file '%s' wasn't written over properly. Size difference: %d", fPath, sizeDiff)
		}

		//Check contents changed by sampling the beginning 10% of the file
		_, err = tempFile.Seek(0, io.SeekStart)
		fatalOnErr(t, err)

		var fBeginBytes []byte = make([]byte, sizeBeforeShred/10)
		_, err = tempFile.Read(fBeginBytes)
		fatalOnErr(t, err)

		if bytes.Equal(fBeginBytes, fBytes[:len(fBeginBytes)]) {
			t.Fatalf("Shredded file contents hasn't been overwritten in file '%s'", fPath)
		}
	}

	//Test1 is a text file smalle than the write buffer used by shred.
	//So we expect shred to adjust its buffer size before the first write
	checkFile("./test1.txt")

	//test2 is a 1 MB+1 byte sized binary file filled with zeros.
	//Shred must write it correctly and also adjust its write buffer on the second write to not over shoot.
	checkFile("./test2.txt")

}

func copyAndHardlink(fPath string) (copyFName, hardlinkFName string, fileBytes []byte) {

	copyFName = fPath + ".copy"
	hardlinkFName = fPath + ".copy.hardlink"

	//Create a copy to work with so we don't destroy the original file
	b, err := os.ReadFile(fPath)
	if err != nil {
		panic(err.Error())
	}
	fileBytes = b

	err = os.WriteFile(copyFName, b, 0666)
	if err != nil {
		panic(err.Error())
	}

	//Create a hardlink of the file we want to shred then shred the original
	os.Remove(hardlinkFName)
	err = os.Link(copyFName, hardlinkFName)
	if err != nil {
		panic(err.Error())
	}

	return
}

func fatalOnErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err.Error())
	}
}
