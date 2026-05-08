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
	"strconv"
	"strings"
)

func createBagContents(cfg *ServiceConfig, opts *ServiceOptions, httpClient *http.Client, item *ItemResponse, rawItem []byte, nativeId string) (string, error) {

	var err error

	// generate the new bag name
	bagName := fmt.Sprintf("%s-%s-%sv%d.%d",
		item.Data.Protocol,
		strings.Replace(item.Data.Authority, ".", "-", -1),
		strings.Replace(strings.ToLower(item.Data.Identifier), "/", "-", -1),
		item.Data.Latest.VersionMajor,
		item.Data.Latest.VersionMinor)

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
	for _, mdf := range item.Data.Latest.Metadata.Citation.Fields {
		if mdf.Name == titleMetadataFieldName {
			s := strings.Trim(string(mdf.Value), "\"")
			err = writeFile(filepath.Join(assetDir, titleFileName), []byte(s))
			if err != nil {
				return "", err
			}
			// add to manifest, no fingerprint yet
			manifest[titleFileName] = ""
		}
		if mdf.Name == descriptionMetadataFieldName {

			value := make([]ItemCitationCompoundDescriptionField, 1)
			err = json.Unmarshal(mdf.Value, &value)
			if err != nil {
				log.Printf("ERROR: json.Unmarshal failed (%s)", err.Error())
				return "", err
			}
			s := strings.Trim(string(value[0].DV.Value), "\"")
			err = writeFile(filepath.Join(assetDir, descriptionFileName), []byte(s))
			if err != nil {
				return "", err
			}
			// add to manifest, no fingerprint yet
			manifest[descriptionFileName] = ""
		}
	}

	err = writeFile(filepath.Join(assetDir, payloadFilename), rawItem)
	if err != nil {
		return "", err
	}

	// add to manifest, no fingerprint yet
	manifest[payloadFilename] = ""

	if opts.NoFiles == false {
		// create our request headers
		headers := map[string]string{"X-Dataverse-key": cfg.ApiToken}

		for ix, f := range item.Data.Latest.Files {

			url := fmt.Sprintf("%s/%s", cfg.ApiEndpoint, cfg.ApiFilePathQuery)
			// fill in the needed components
			url = strings.Replace(url, "{{:id}}", strconv.Itoa(f.File.Id), 1)

			// seems to be the case sometimes...
			if f.File.Filesize <= 0 {
				log.Printf("WARNING: suspect filesize for %s (%d), ignoring this file", f.File.Filename, f.File.Filesize)
				continue
			}

			log.Printf("INFO: downloading %d of %d, %s (%d bytes)", ix+1, len(item.Data.Latest.Files), f.File.Filename, f.File.Filesize)

			// map the name if we have too
			mappedName := specialCaseNameMapper(f.File.Filename, nativeId)

			err = fastDownload(url, headers, filepath.Join(assetDir, mappedName))
			if err != nil {
				return "", err
			}

			fp := f.File.MD5
			if len(f.File.OriginalName) != 0 {

				// this is a special case where the original file was transformed in some manner by
				// the external service so the fingerprint is unreliable
				log.Printf("WARNING: likely derived asset, ignoring fingerprint for %s", mappedName)
				fp = ""
			}
			// add to manifest, add fingerprint if we have it
			manifest[mappedName] = fp
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
