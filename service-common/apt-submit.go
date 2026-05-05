//
//
//

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/seqsense/s3sync/v2"
)

type Logger struct {
}

func (l *Logger) Logf(format string, v ...any) {
	//log.Printf("INFO: "+format, v...)
}

func submitContent(baseUrl string, cmd string, environment string, poll bool, cid string, localDir string) error {

	// create our HTTP client
	httpClient := newHttpClient(1, 10)
	// important, cleanup properly
	defer httpClient.CloseIdleConnections()

	// get the list of bags to submit
	bagList := makeBagList(cmd, localDir)

	// do the submission registration
	regResponse, err := register(baseUrl, environment, cid, httpClient)
	if err != nil {
		return err
	}

	// create appropriate s3 key...
	s3Key := regResponse.DepositPath
	if cmd == "submit-bag" {
		s3Key = filepath.Join(regResponse.DepositPath, bagList[0])
	}

	// upload the assets to S3
	err = syncAssets(localDir, regResponse.DepositBucket, s3Key)
	if err != nil {
		return err
	}

	// do the submission initiate
	err = initiate(baseUrl, environment, cid, regResponse.SubmissionIdentifier, bagList, httpClient)
	if err != nil {
		return err
	}

	// we submitted a tree, poll for submission status
	//if cmd == "submit-tree" {
	return getSubmitStatus(baseUrl, environment, poll, regResponse.SubmissionIdentifier)
	//}

	// we submitted a bag, poll for bag status
	//return getBagStatus(baseUrl, environment, poll, bagList[0])
}

func register(baseUrl string, environment string, cid string, httpClient *http.Client) (*SubmitRegisterResponse, error) {

	// construct the URL
	url := baseUrl + "/register/" + environment

	req := SubmitRegisterRequest{}
	req.ClientIdentifier = cid
	pl, err := json.Marshal(req)
	if err != nil {
		log.Printf("ERROR: json.Marshal() failed (%s)", err.Error())
		return nil, err
	}

	// post the request
	pl, err = httpPost(httpClient, url, pl, "application/json")
	if err != nil {
		log.Printf("ERROR: received [%s] (%s)", string(pl), err.Error())
		return nil, err
	}

	// and process the response
	resp := SubmitRegisterResponse{}
	err = json.Unmarshal(pl, &resp)
	if err != nil {
		log.Printf("ERROR: json.Unmarshal failed (%s)", err.Error())
		return nil, err
	}

	log.Printf("INFO: submit register complete (%s)...", resp.SubmissionIdentifier)
	return &resp, nil
}

func initiate(baseUrl string, environment string, cid string, sid string, bagList []string, httpClient *http.Client) error {

	// construct the URL
	url := baseUrl + "/initiate/" + environment

	req := SubmitInitiateRequest{}
	req.ClientIdentifier = cid
	req.SubmissionIdentifier = sid
	req.BagFolders = bagList

	pl, err := json.Marshal(req)
	if err != nil {
		log.Printf("ERROR: json.Marshal() failed (%s)", err.Error())
		return err
	}

	// post the request
	pl, err = httpPost(httpClient, url, pl, "application/json")
	if err != nil {
		log.Printf("ERROR: received [%s] (%s)", string(pl), err.Error())
		return err
	}

	// and process the response
	resp := SubmitInitiateResponse{}
	err = json.Unmarshal(pl, &resp)
	if err != nil {
		log.Printf("ERROR: json.Unmarshal failed (%s)", err.Error())
		return err
	}

	log.Printf("INFO: submit initiate complete...")
	return nil
}

func syncAssets(local string, bucket string, path string) error {

	log.Printf("INFO: syncing assets...")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}

	// set the logger cos we dont want the standard logging
	s3sync.SetLogger(&Logger{})

	// new sync manager
	syncManager := s3sync.New(cfg, s3sync.WithParallel(5))

	// our destination location
	source := fmt.Sprintf("s3://%s/%s", bucket, path)
	log.Printf("INFO: sync from [%s] -> [%s]", local, source)

	start := time.Now()
	err = syncManager.Sync(context.TODO(), local, source)
	if err != nil {
		log.Printf("ERROR: sync error (%s)", err.Error())
		return err
	}

	stats := syncManager.GetStatistics()
	duration := time.Since(start)
	log.Printf("INFO: sync completed (elapsed %0.2f seconds)", duration.Seconds())
	log.Printf("INFO: %d bytes written, %d files uploaded, %d files deleted", stats.Bytes, stats.Files, stats.DeletedFiles)

	return nil
}

// if we are submitting a tree, we need to locate all the directories. If we are submitting a bag
// the bag name is the basename of the localdir
func makeBagList(cmd string, localDir string) []string {

	bagList := make([]string, 0)

	if cmd == "submit-bag" {
		bagList = append(bagList, filepath.Base(localDir))
		return bagList
	}

	entries, err := os.ReadDir(localDir)
	if err != nil {
		log.Printf("ERROR: read dir (%s)", err.Error())
		return bagList
	}

	for _, entry := range entries {
		if entry.IsDir() { // Check if the entry is a directory
			bagList = append(bagList, filepath.Base(entry.Name()))
		}
	}
	return bagList
}

//
// end of file
//
