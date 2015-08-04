package main

import (
	"fmt"
	"io"
	"flag"
	"os"
	"math"
	"path/filepath"
	"crypto/sha1"
	)

const CHUNKSIZE uint64 = 8192    // This'll work -- pretty standard size

func checkSum(pathname string, filesize int64, err error) string {

    file, err := os.Open(pathname)
    if err != nil {
       panic(err.Error())
    }

    defer file.Close()

   blocks := uint64(math.Ceil(float64(filesize) / float64(CHUNKSIZE)))

   hash := sha1.New()

   for i := uint64(0); i < blocks; i++ {
	blocksize := int(math.Min(float64(CHUNKSIZE), float64(filesize-int64(i*CHUNKSIZE))))
	buf := make([] byte, blocksize)

        file.Read(buf)
        io.WriteString(hash, string(buf))   // 'tack on' the end
   }

   return fmt.Sprintf("\n%s (%d): %x", pathname, filesize, hash.Sum(nil))
}

func walkPathNSum(path string, f os.FileInfo, err error) error {
  fmt.Printf("\n%s", checkSum(path, int64(f.Size()), err ))
  return nil
}

func main() {
	flag.Parse()
	root := flag.Arg(0)
	filepath.Walk(root, walkPathNSum)
}
