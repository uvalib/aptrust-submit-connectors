//
//
//

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

func getUpdatedIds(cfg *ServiceConfig, opts *ServiceOptions) ([]string, error) {

	// create our HTTP client
	httpClient := newHttpClient(1, cfg.HTTPTimeout)
	// important, cleanup properly
	defer httpClient.CloseIdleConnections()

	// create our request headers
	headers := map[string]string{"Accept": "application/json"}

	// fill in the needed components
	url := fmt.Sprintf("%s/%s", cfg.ApiEndpoint, cfg.ApiUpdatedPathQuery)
	url = strings.Replace(url, "{{:startdate}}", opts.StartDate, 1)
	url = strings.Replace(url, "{{:enddate}}", opts.EndDate, 1)

	// issue the request
	pl, err := httpGet(httpClient, url, headers)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("%s\n", pl)

	// process the response
	resp := UpdatedResponse{}
	err = json.Unmarshal(pl, &resp)
	if err != nil {
		log.Printf("ERROR: json.Unmarshal failed (%s)", err.Error())
		return nil, err
	}

	return resp, nil
}

//
// end of file
//
