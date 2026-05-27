//
//
//

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/melbahja/got"
)

var ErrBadDownloadSize = fmt.Errorf("bad download size")

func fastDownload(url string, size uint64, headers map[string]string, outfile string) error {

	dl := got.NewDownload(context.Background(), url, outfile)
	dl.Header = make([]got.GotHeader, 0)
	for k, v := range headers {
		dl.Header = append(dl.Header, got.GotHeader{Key: k, Value: v})
	}

	err := dl.Init()
	if err != nil {
		log.Printf("ERROR: during init (%s)", err.Error())
		return err
	}

	// verify the reported size is as expected
	if dl.TotalSize() != size {
		log.Printf("ERROR: expected %d bytes, reported %d bytes", size, dl.TotalSize())
		return ErrBadDownloadSize
	}

	err = dl.Start()
	if err != nil {
		log.Printf("ERROR: during start (%s)", err.Error())
		return err
	}

	stat, err := os.Stat(outfile)
	if err != nil {
		log.Printf("ERROR: during stat (%s)", err.Error())
		return err
	}

	if stat.Size() != int64(size) {
		log.Printf("ERROR: expected %d bytes, received %d bytes", size, stat.Size())
		return ErrBadDownloadSize
	}

	return nil
}

//
// end of file
//
