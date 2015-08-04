package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	//	"strings"
	//	"time"
)

const CHUNKSIZE uint64 = 8192 // This'll work -- pretty standard size
var wrkQueue = make(chan string)
var outQueue = make(chan string)

func fmtFileInfo(pathname string, fi os.FileInfo, err error) string {
	//	var filesize int64 = fi.Size()
	//	var tmpString = fmt.Sprintf("%s,%d,%x", pathname, filesize, 0)
	return fmt.Sprintf("%s", pathname)
}

func checkSum(pathname string, fi os.FileInfo, err error) string {
	var filesize int64 = fi.Size()

	file, err := os.Open(pathname)
	if err != nil {
		panic(err.Error())
	}

	defer file.Close()

	blocks := uint64(math.Ceil(float64(filesize) / float64(CHUNKSIZE)))

	hash := sha1.New()

	for i := uint64(0); i < blocks; i++ {
		blocksize := int(math.Min(float64(CHUNKSIZE), float64(filesize-int64(i*CHUNKSIZE))))
		buf := make([]byte, blocksize)

		file.Read(buf)
		io.WriteString(hash, string(buf)) // 'tack on' the end
	}

	return fmt.Sprintf("%s,%d,%x", pathname, filesize, hash.Sum(nil))
}

func walkPathNSum(path string, f os.FileInfo, err error) error {
	//	wrkQueue <- checkSum(path, f, err)
	wrkQueue <- fmtFileInfo(path, f, err)
	return nil
}

func Worker(i int, inq chan string, outq chan string, isMD5 bool) {
	var ckString string
	//	var string_components []string
	//	var filesize int64

	for {
		ckString = <-inq
		if len(ckString) == 0 {
			break
		}
		//		string_components := strings.Split(ckString, ",")
		//		filetype := string_components[0]
		//	filepath := string_components[0]
		//		fmt.Sscanf(string_components[2], "%d", &filesize)
		fs, err := os.Stat(ckString)
		if err != nil {
			panic(err.Error())
		}
		outq <- checkSum(ckString, fs, err)
		//		time.Sleep(time.Second * 1)
	}
}

func Outputter(outq chan string) {
	var outString string

	for {
		outString = <-outq
		if len(outString) == 0 {
			break
		}
		fmt.Println("\n", outString)
		//		time.Sleep(time.Second * 1)
	}
}

func main() {
	var pause string
	// Assumes that the first argument is a FQDN
	flag.Parse()
	root := flag.Arg(0)
	go Worker(1, wrkQueue, outQueue, false)
	go Outputter(outQueue)

	filepath.Walk(root, walkPathNSum)
	fmt.Scanln(&pause)
	wrkQueue <- ""
}
