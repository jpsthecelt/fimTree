// Calculate checksums for each file in a directory-tree
package main

import (
	"crypto/md5"
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

const CHUNKSIZE uint64 = 8192

var wrkQueue = make(chan *fInfo)
var outQueue = make(chan string)

func checkSumSHA1(threadID int, pathname string, fi *fInfo, err error) string {
	var filesize int64 = fi.sz
	var hostname = "grasskeet-ubS"

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

	return fmt.Sprintf("%s:%s,%d,%x,%x,thread:%d", hostname, pathname, filesize, hash.Sum(nil), fi.modTime, threadID)
}

func checkSumMD5(threadID int, pathname string, fi *fInfo, err error) string {
	var filesize int64 = fi.sz

	file, err := os.Open(pathname)
	if err != nil {
		panic(err.Error())
	}

	defer file.Close()

	blocks := uint64(math.Ceil(float64(filesize) / float64(CHUNKSIZE)))

	hash := md5.New()

	for i := uint64(0); i < blocks; i++ {
		blocksize := int(math.Min(float64(CHUNKSIZE), float64(filesize-int64(i*CHUNKSIZE))))
		buf := make([]byte, blocksize)

		file.Read(buf)
		io.WriteString(hash, string(buf)) // 'tack on' the end
	}

	return fmt.Sprintf("%s,%d,%x,%x,thread:%d", pathname, filesize, hash.Sum(nil), fi.modTime, threadID)
}

func walkPathNSum(pathname string, f os.FileInfo, err error) error {
	wrkQueue <- &fInfo{name: pathname, sz: f.Size(), mode: f.Mode(), modTime: f.ModTime()}
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
		if isMD5 {
			outq <- checkSumMD5(i, ckString.name, ckString, err)
		} else {
			outq <- checkSumSHA1(i, ckString.name, ckString, err)
		}
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
	var hostname = os.Getenv("HOSTNAME")

	// Assumes that the first argument is a FQDN, no '~' and uses '/'s vs. '\'s
	flag.Parse()
	root := flag.Arg(0)
	ncpu := runtime.NumCPU()
	fmt.Println("\nHostname = %s", hostname)
	fmt.Println("\nWorker threads: ", ncpu)
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
