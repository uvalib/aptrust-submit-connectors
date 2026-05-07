//
//
//

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

func processUpdatedIds(cfg *ServiceConfig, opts *ServiceOptions, ids []string) error {

	log.Printf("INFO: processing %d updated item(s)", len(ids))

	// create our HTTP client
	httpClient := newHttpClient(1, cfg.HTTPTimeout)
	// important, cleanup properly
	defer httpClient.CloseIdleConnections()

	for ix, id := range ids {

		log.Printf("INFO: processing %d of %d (id: %s)", ix+1, len(ids), id)

		url := fmt.Sprintf("%s/%s", cfg.ApiEndpoint, cfg.ApiItemPathQuery)

		// fill in the needed components
		url = strings.Replace(url, "{{:id}}", id, 1)

		// create our request headers
		headers := map[string]string{"Accept": "application/json", "X-Dataverse-key": cfg.ApiToken}

		// issue the request
		pl, err := httpGet(httpClient, url, headers)
		if err != nil {
			//return err
			continue
		}

		// process the response
		resp := ItemResponse{}
		err = json.Unmarshal(pl, &resp)
		if err != nil {
			log.Printf("ERROR: json.Unmarshal failed (%s)", err.Error())
			return err
		}

		// not sure why this would be the case
		if resp.Status != "OK" {
			err = fmt.Errorf("received response '%s'", resp.Status)
			log.Printf("ERROR: unexpected response (%s)", err.Error())
			return err
		}

		//fmt.Printf("%s\n", pl)
		//fmt.Printf("found %d fields\n", len(resp.Data.Latest.Metadata.Citation.Fields))

		// generate the bag contents
		bagName, err := createBagContents(cfg, opts, httpClient, &resp, pl, id)
		if err != nil {
			return err
		}

		// do we actually want to submit this bag?
		if opts.NoSubmit == false {
			workDir := filepath.Join(cfg.ScratchFileSystem, bagName)
			log.Printf("INFO: submitting bag: %s", bagName)
			err = submitBagContents(cfg, httpClient, workDir, bagName)
			if err != nil {
				return err
			}
			_ = cleanScratchFilesystem(workDir)
		} else {
			log.Printf("INFO: not submitting bag: %s", bagName)
		}
	}

	return nil
}

//
// end of file
//
