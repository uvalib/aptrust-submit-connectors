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
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/seqsense/s3sync/v2"
)

func submitBagContents(cfg *ServiceConfig, httpClient *http.Client, localDir string, collectionName string, bagName string) error {

	// do the submission registration
	regResponse, err := register(cfg, httpClient, collectionName)
	if err != nil {
		return err
	}

	// create appropriate s3 key...
	s3Key := filepath.Join(regResponse.DepositPath, bagName)

	// upload the assets to S3
	err = syncAssets(localDir, regResponse.DepositBucket, s3Key)
	if err != nil {
		return err
	}

	// do the submission initiate
	err = initiate(cfg, httpClient, regResponse.SubmissionIdentifier, bagName)
	if err != nil {
		return err
	}

	return nil
}

func register(cfg *ServiceConfig, httpClient *http.Client, collectionName string) (*SubmitRegisterResponse, error) {

	req := SubmitRegisterRequest{}
	req.ClientIdentifier = cfg.APTServiceClient
	req.Collection = collectionName
	pl, err := json.Marshal(req)
	if err != nil {
		log.Printf("ERROR: json.Marshal() failed (%s)", err.Error())
		return nil, err
	}

	// create our request headers
	headers := map[string]string{"Accept": "application/json"}

	// post the request
	pl, err = httpPost(httpClient, cfg.APTServiceRegister, pl, headers)
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

func initiate(cfg *ServiceConfig, httpClient *http.Client, sid string, bagName string) error {

	req := SubmitInitiateRequest{}
	req.ClientIdentifier = cfg.APTServiceClient
	req.SubmissionIdentifier = sid
	req.BagFolders = []string{bagName}

	pl, err := json.Marshal(req)
	if err != nil {
		log.Printf("ERROR: json.Marshal() failed (%s)", err.Error())
		return err
	}

	// create our request headers
	headers := map[string]string{"Accept": "application/json"}

	// post the request
	pl, err = httpPost(httpClient, cfg.APTServiceSubmit, pl, headers)
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

	// new sync manager (16 workers is the default)
	syncManager := s3sync.New(cfg, s3sync.WithParallel(32))

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

//
// end of file
//
