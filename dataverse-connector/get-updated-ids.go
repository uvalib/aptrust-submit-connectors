//
//
//

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
)

func getUpdatedIds(cfg *ServiceConfig, opts *ServiceOptions) ([]string, error) {

	log.Printf("INFO: getting latest items...")

	// create our HTTP client
	httpClient := newHttpClient(1, cfg.HTTPTimeout)
	// important, cleanup properly
	defer httpClient.CloseIdleConnections()

	currentRow := 0
	perPage := 250

	// create our request headers
	headers := map[string]string{"Accept": "application/json", "X-Dataverse-key": cfg.ApiToken}

	// our response slice
	ids := make([]string, 0)

	for {
		// fill in the needed components
		url := fmt.Sprintf("%s/%s", cfg.ApiEndpoint, cfg.ApiUpdatedPathQuery)
		url = strings.Replace(url, "{{:startrow}}", strconv.Itoa(currentRow), 1)
		url = strings.Replace(url, "{{:perpage}}", strconv.Itoa(perPage), 1)
		url = strings.Replace(url, "{{:startdate}}", opts.StartDate, 1)
		url = strings.Replace(url, "{{:enddate}}", opts.EndDate, 1)

		// issue the request
		pl, err := httpGet(httpClient, url, headers)
		if err != nil {
			return nil, err
		}

		// process the response
		resp := UpdatedResponse{}
		err = json.Unmarshal(pl, &resp)
		if err != nil {
			log.Printf("ERROR: json.Unmarshal failed (%s)", err.Error())
			return nil, err
		}

		// not sure why this would be the case
		if resp.Status != "OK" {
			err = fmt.Errorf("received response '%s'", resp.Status)
			log.Printf("ERROR: unexpected response (%s)", err.Error())
			return nil, err
		}

		// create the response list
		for _, i := range resp.Data.Items {
			ids = append(ids, strconv.Itoa(i.EntityId))
		}

		// increment the row count
		currentRow += len(resp.Data.Items)

		// are we done?
		if resp.Data.TotalCount <= currentRow {
			break
		}
	}

	return ids, nil
}

//
// end of file
//
