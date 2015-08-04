package main

import (
	"fmt"
	"io"
	"flag"
	"os"
	"math"
	"time"
	"path/filepath"
	"crypto/sha1"
//	"crypto/md5"
	)

const CHUNKSIZE uint64 = 8192    // This'll work -- pretty standard size
var Hash  = sha1.New()
var C chan string

func checkSum(pathname string, filesize int64, err error) string {

    file, err := os.Open(pathname)
    if err != nil {
       panic(err.Error())
    }

    defer file.Close()

   blocks := uint64(math.Ceil(float64(filesize) / float64(CHUNKSIZE)))

//   Hash := sha1.New()

   for i := uint64(0); i < blocks; i++ {
	blocksize := int(math.Min(float64(CHUNKSIZE), float64(filesize-int64(i*CHUNKSIZE))))
	buf := make([] byte, blocksize)

        file.Read(buf)
        io.WriteString(Hash, string(buf))   // 'tack on' the end
   }

   return fmt.Sprintf("%s (%d): %x", pathname, filesize, Hash.Sum(nil))
}

func outputCksumString (c chan string) {
  for {
        msg := <- c
	fmt.Println("\n", msg)
	time.Sleep(time.Second * 1)
      }
}

func walkPathNSum(path string, f os.FileInfo, err error) error {
  C <- fmt.Sprintf("%s", checkSum(path, int64(f.Size()), err ))
  return nil
}

func main() {
//        var C chan string = make(chan string)
	C = make(chan string)

//	boolPtr := flag.Bool("md5", false, "a bool")
	flag.Parse()
	root := flag.Arg(0)

	// Look for md5 flag, otherwise use sha1
//	if (*boolPtr) {
//	  Hash := md5.New()
//	} else  {
//	  Hash := sha1.New()
//	}

	go outputCksumString(C)

	filepath.Walk(root, walkPathNSum)

	var input string
	fmt.Scanln(&input)
}
