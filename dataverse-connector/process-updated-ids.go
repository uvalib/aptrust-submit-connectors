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

func processUpdatedIds(cfg *ServiceConfig, opts *ServiceOptions, ids []string) error {

	log.Printf("INFO: processing %d updated id(s)", len(ids))

	// create our HTTP client
	httpClient := newHttpClient(1, cfg.HTTPTimeout)
	// important, cleanup properly
	defer httpClient.CloseIdleConnections()

	for _, id := range ids {
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

		// generate the new bag name
		bagName := fmt.Sprintf("%s-%s-%sv%d.%d",
			resp.Data.Protocol,
			strings.Replace(resp.Data.Authority, ".", "-", -1),
			strings.Replace(strings.ToLower(resp.Data.Identifier), "/", "-", -1),
			resp.Data.Latest.VersionMajor,
			resp.Data.Latest.VersionMinor)

		//fmt.Printf("received %d file(s)\n", len(resp.Data.Latest.Files))
		fmt.Printf("bag name: %s\n", bagName)
	}

	return nil
}

//
// end of file
//
