//
//
//

package main

import (
	"context"
	"log"

	"github.com/melbahja/got"
)

func fastDownload(url string, headers map[string]string, outfile string) error {

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

	err = dl.Start()
	if err != nil {
		log.Printf("ERROR: during start (%s)", err.Error())
		return err
	}

	return nil
}

//
// end of file
//
