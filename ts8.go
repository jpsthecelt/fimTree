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
	"strings"
	"time"
)

type fInfo struct {
	name    string
	sz      int64
	mode    os.FileMode
	modTime time.Time
}

type checksumWorkerFunction func(int, string, string, *fInfo, error) string

const CHUNKSIZE uint64 = 8192

var wrkQueue = make(chan *fInfo)
var outQueue = make(chan string)
var hostname = os.Getenv("HOSTNAME")

func checkSumSHA1(threadID int, pathname string, hname string, fi *fInfo, err error) string {
	var filesize int64 = fi.sz

	file, err := os.Open(pathname)
	if err != nil {
		return fmt.Sprintf("(SHA1)-%s:%s,%s, %d,%x,%x,thread:%d", hname, pathname, err, -1, -1, fi.modTime, threadID)
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

	return fmt.Sprintf("(SHA1)-%s:%s,%d,%x,%x,thread:%d", hname, pathname, filesize, hash.Sum(nil), fi.modTime, threadID)
}

func checkSumMD5(threadID int, pathname string, hname string, fi *fInfo, err error) string {
	var filesize int64 = fi.sz

	file, err := os.Open(pathname)
	if err != nil {
		return fmt.Sprintf("(MD5)-%s:%s,%s %d,%x,%x,thread:%d", hname, pathname, err, -1, -1, fi.modTime, threadID)
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

	return fmt.Sprintf("(MD5)-%s:%s,%d,%x,%x,thread:%d", hname, pathname, filesize, hash.Sum(nil), fi.modTime, threadID)
}

// This just 'walks' through the filesystem, grabbing fileInfo information; queueing up to the 'Work' input
func walkPathNSum(pathname string, f os.FileInfo, err error) error {
	wrkQueue <- &fInfo{name: pathname, sz: f.Size(), mode: f.Mode(), modTime: f.ModTime()}
	return nil
}

// This worker function grabs a string from the input Q, uses the checksumWorkerFunction pointer to checksum the file and sends that to the putput Q
func Worker(i int, inq chan *fInfo, outq chan string, cwf checksumWorkerFunction ) {
	var ckString *fInfo
	var err error

	for {
		ckString = <-inq
		if ckString == nil {
			break
		}
		outq <- cwf(i, ckString.name, hostname, ckString, err)
	}
}

// This outputs the calculated checksum string to the appropriate entity (today, the console; tomorrow a DB)
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
//	var pause string
	var numberCpus = runtime.NumCPU()

	nPtr := flag.Int("cpuLimit", 0, "an int")

	// Assumes that the first argument is a FQDN, no '~' and uses '/'s vs. '\'s
	flag.Parse()
	root := flag.Arg(0)

	allArgs := strings.ToLower(fmt.Sprintln(os.Args[1:]))

	if *nPtr > 0 {
		runtime.GOMAXPROCS(*nPtr)
		fmt.Println("\nWorker threads: changed from ", numberCpus, " to ", *nPtr)
	} else {
		*nPtr = numberCpus
		runtime.GOMAXPROCS(numberCpus)
		fmt.Println("\nWorker threads: ", numberCpus)
	}

	if runtime.GOOS == "windows" {
		hostname = os.Getenv("COMPUTERNAME")
	} else {
		hostname = os.Getenv("HOSTNAME")
	}
	fmt.Println("\n(apparent) Hostname = ", hostname)

	// spawn workers
	for i := 0; i < *nPtr; i++ {
		if strings.Contains(allArgs, "md5") {
			go Worker(i, wrkQueue, outQueue, checkSumMD5)
		} else {
			go Worker(i, wrkQueue, outQueue, checkSumSHA1)
		}
	}
	go Outputter(outQueue)

	filepath.Walk(root, walkPathNSum)

//	fmt.Println("\nPress ENTER to continue"0)
//	fmt.Scanln(&pause)

	fmt.Println("\n")
	for i := 0; i < *nPtr; i++ {
		wrkQueue <- nil
	}
	outQueue <- ""
}
