//
//
//

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var bagNamePrefix = "LibraOpen-"

func createBagContents(cfg *ServiceConfig, opts *ServiceOptions, httpClient *http.Client, item *ItemResponse, rawItem []byte, nativeId string) (string, error) {

	var err error

	// the identifier we use for the bag might be the native item identifier or it might be the
	// legacy libra-open id if it exists in the metadata
	identifier := nativeId
	md, ok := item.Item.Metadata[identifierMetadataFieldName]
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
	md, ok = item.Item.Metadata[descriptionMetadataFieldName]
	if ok {
		err = writeFile(filepath.Join(assetDir, descriptionFileName), []byte(md[0].Value))
		if err != nil {
			return "", err
		}
		// add to manifest, no fingerprint yet
		manifest[descriptionFileName] = ""
	}
	md, ok = item.Item.Metadata[titleMetadataFieldName]
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
		for _, f := range item.Content {

			log.Printf("INFO: downloading %s", f.Link)

			// download the file
			b, err := httpGet(httpClient, f.Link, make(map[string]string))
			if err != nil {
				return "", err
			}

			// and write it
			err = writeFile(filepath.Join(assetDir, f.Name), b)
			if err != nil {
				return "", err
			}

			// add to manifest, add fingerprint if we have it
			manifest[f.Name] = ""
		}
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
