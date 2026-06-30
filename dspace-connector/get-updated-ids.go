//
//
//

package main

import (
	"encoding/json"
	//"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
)

func getUpdatedIds(cfg *ServiceConfig, opts *ServiceOptions) ([]string, error) {

	log.Printf("INFO: getting items between [%s] and [%s]", opts.StartDate, opts.EndDate)

	// create our HTTP client
	httpClient := newHttpClient(1, cfg.HTTPTimeout)
	// important, cleanup properly
	defer httpClient.CloseIdleConnections()

	currentPage := 0
	perPage := 100

	// create our request headers
	headers := map[string]string{"Accept": "application/json"}
	headers = addAuthHeader(headers)

	// our response slice
	ids := make([]string, 0)

	for {

		// fill in the needed components
		url := fmt.Sprintf("%s/%s", cfg.ApiEndpoint, cfg.ApiUpdatedPathQuery)
		url = strings.Replace(url, "{{:startpage}}", strconv.Itoa(currentPage), 1)
		url = strings.Replace(url, "{{:perpage}}", strconv.Itoa(perPage), 1)
		url = strings.Replace(url, "{{:startdate}}", opts.StartDate, 1)
		url = strings.Replace(url, "{{:enddate}}", opts.EndDate, 1)

		// issue the request
		pl, err := httpGet(httpClient, url, headers)
		if err != nil {
			return nil, err
		}

		//fmt.Printf("%s\n", pl)

		// process the response
		resp := SearchResponse{}
		err = json.Unmarshal(pl, &resp)
		if err != nil {
			log.Printf("ERROR: json.Unmarshal failed (%s)", err.Error())
			return nil, err
		}

		log.Printf("INFO: received %d object(s)", len(resp.Embedded.SearchResults.EmbeddedObjects.Objects))

		log.Printf("INFO: page %d of %d, total count %d", resp.Embedded.SearchResults.ResultsPage.PageNumber, resp.Embedded.SearchResults.ResultsPage.TotalPages, resp.Embedded.SearchResults.ResultsPage.TotalCount)

		// populate the response list
		for _, i := range resp.Embedded.SearchResults.EmbeddedObjects.Objects {
			ids = append(ids, i.IndexableObject.Object.Uuid)
		}

		// increment the page number
		currentPage += 1

		// are we done?
		if resp.Embedded.SearchResults.ResultsPage.TotalPages <= currentPage {
			break
		}
	}

	return ids, nil
}

//
// end of file
//
