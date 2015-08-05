package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	//	"strings"
	"time"
)

type fInfo struct {
	name    string
	sz      int64
	mode    os.FileMode
	modTime time.Time
}

const CHUNKSIZE uint64 = 8192 // This'll work -- pretty standard size
var wrkQueue = make(chan *fInfo)
var outQueue = make(chan string)

func fmtFileInfo(pathname string, fi os.FileInfo, err error) *fInfo {
	f := &fInfo{name: pathname, sz: fi.Size(), mode: fi.Mode(), modTime: fi.ModTime()}
	return f
}

func checkSum(threadID int, pathname string, fi *fInfo, err error) string {
	var filesize int64 = fi.sz

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

	return fmt.Sprintf("%s,%d,%x,thread:%d", pathname, filesize, hash.Sum(nil), threadID)
}

func walkPathNSum(path string, f os.FileInfo, err error) error {
	wrkQueue <- fmtFileInfo(path, f, err)
	return nil
}

func Worker(i int, inq chan *fInfo, outq chan string, isMD5 bool) {
	var ckString *fInfo
	var err error

	for {
		ckString = <-inq
		if ckString == nil {
			break
		}
		outq <- checkSum(i, ckString.name, ckString, err)
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
	// Assumes that the first argument is a FQDN, no '~' and uses '/'s vs. '\'s
	flag.Parse()
	root := flag.Arg(0)
	ncpu := runtime.NumCPU()
	fmt.Println("\nWorking with %d CPUs/threads", ncpu)
	runtime.GOMAXPROCS(ncpu)

	// spawn workers
	for i := 0; i < ncpu; i++ {
		go Worker(i, wrkQueue, outQueue, false)
	}
	go Outputter(outQueue)

	filepath.Walk(root, walkPathNSum)
	//	fmt.Println("\nPress ENTER to continue"0)
	fmt.Scanln(&pause)
	fmt.Println("\n")
	for i := 0; i < ncpu; i++ {
		wrkQueue <- nil
	}
	outQueue <- ""
}
