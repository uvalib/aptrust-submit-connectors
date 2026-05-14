//
//
//

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"
)

var bagNamePrefix = "LibraOpen-"

func createBagContents(cfg *ServiceConfig, opts *ServiceOptions, httpClient *http.Client, item *ItemResponse, rawItem []byte, nativeId string) (string, error) {

	var err error

	// the identifier we use for the bag might be the native item identifier or it might be the
	// legacy libra-open id if it exists in the metadata
	identifier := nativeId
	md, ok := item.Metadata[identifierMetadataFieldName]
	if ok {
		identifier = md[0].Value
	}

	// generate the new bag name
	bagName := fmt.Sprintf("%s%s", bagNamePrefix, identifier)

	log.Printf("INFO: creating contents for bag: %s", bagName)

	// create/cleanup the asset working directory
	assetDir := filepath.Join(cfg.ScratchFileSystem, bagName)
	_ = cleanScratchFilesystem(assetDir)
	err = os.MkdirAll(assetDir, 0755)
	if err != nil {
		return "", err
	}

	// the manifest consists of a series of file names plus the corresponding md5 fingerprint
	// if we do not have a fingerprint from the external system, we need to create one
	manifest := make(map[string]string)

	// extract the interesting info from the metadata
	md, ok = item.Metadata[descriptionMetadataFieldName]
	if ok {
		err = writeFile(filepath.Join(assetDir, descriptionFileName), []byte(md[0].Value))
		if err != nil {
			return "", err
		}
		// add to manifest, no fingerprint yet
		manifest[descriptionFileName] = ""
	}
	md, ok = item.Metadata[titleMetadataFieldName]
	if ok {
		err = writeFile(filepath.Join(assetDir, titleFileName), []byte(md[0].Value))
		if err != nil {
			return "", err
		}
		// add to manifest, no fingerprint yet
		manifest[titleFileName] = ""
	}

	err = writeFile(filepath.Join(assetDir, payloadFilename), rawItem)
	if err != nil {
		return "", err
	}

	// add to manifest, no fingerprint yet
	manifest[payloadFilename] = ""

	if opts.NoFiles == false {
		bundlesUrl, ok := item.Links["bundles"]
		if ok {
			// create our request headers
			headers := map[string]string{"Accept": "application/json"}

			// issue the request
			pl, err := httpGet(httpClient, bundlesUrl.Href, headers)
			if err != nil {
				return "", err
			}

			// process the response
			resp := BundlesResponse{}
			err = json.Unmarshal(pl, &resp)
			if err != nil {
				log.Printf("ERROR: json.Unmarshal failed (%s)", err.Error())
				return "", err
			}

			for _, bundle := range resp.EmbeddedBundles.Bundles {
				//fmt.Printf("Bundle: [%s]\n", bundle.Name)
				if bundle.Name == "ORIGINAL" {
					bitstreamsUrl, ok := bundle.Links["bitstreams"]
					if ok {
						// issue the request
						pl, err := httpGet(httpClient, bitstreamsUrl.Href, headers)
						if err != nil {
							return "", err
						}

						// process the response
						resp := BitstreamsResponse{}
						err = json.Unmarshal(pl, &resp)
						if err != nil {
							log.Printf("ERROR: json.Unmarshal failed (%s)", err.Error())
							return "", err
						}

						//fmt.Printf("%s\n", pl)

						for _, bs := range resp.EmbeddedBitstreams.Bitstreams {
							contentUrl, ok := bs.Links["content"]
							if ok {
								log.Printf("INFO: downloading %s (%s)", bs.Name, humanize.IBytes((uint64)(bs.Size)))

								// map the name if we have too
								mappedName := specialCaseNameMapper(bs.Name, nativeId)

								// try the download 3 times
								err = retry(3, 1*time.Second, func() error {
									return fastDownload(contentUrl.Href, headers, filepath.Join(assetDir, mappedName))
								})

								if err != nil {
									log.Printf("ERROR: skipping download of %s (%s)", bs.Name, err.Error())
									continue
								}

								// add to manifest, add fingerprint if we have it
								fp := ""
								if bs.Checksum.Algorithm == "MD5" {
									fp = bs.Checksum.Value
								}
								manifest[mappedName] = fp
							}
						}
					}
					break
				}
			}
		}

		//for n, v := range item.Links {

		//fmt.Printf("link [%s]: [%s]\n", n, v.Href)
		//log.Printf("INFO: downloading %d of %d, %s (%d bytes)", ix+1, len(item.Content), f.Link)

		// map the name if we have too
		//mappedName := specialCaseNameMapper(f.Name, nativeId)

		// download the file
		//b, err := httpGet(httpClient, f.Link, make(map[string]string))
		//if err != nil {
		//	return "", err
		//}

		// and write it
		//err = writeFile(filepath.Join(assetDir, f.Name), b)
		//if err != nil {
		//	return "", err
		//}

		// add to manifest, add fingerprint if we have it
		//manifest[f.Name] = ""
		//}
	}

	// generate the manifest contents
	manifestContents, err := generateManifestContent(assetDir, manifest)
	if err != nil {
		return "", err
	}

	return bagName, writeFile(filepath.Join(assetDir, manifestFilename), []byte(manifestContents))
}

//
// end of file
//
