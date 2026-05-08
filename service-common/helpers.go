package main

import (
	"crypto/md5"
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"
	"slices"
)

func fatalIfError(err error) {
	if err != nil {
		log.Fatalf("FATAL ERROR: %s", err.Error())
	}
}

func md5Checksum(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("ERROR: reading [%s] (%s)", filename, err.Error())
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(data)), nil
}

func writeFile(filename string, buffer []byte) error {

	err := os.WriteFile(filename, buffer, 0644)
	if err != nil {
		log.Printf("ERROR: writing [%s] (%s)", filename, err.Error())
		return err
	}
	return nil
}

func cleanScratchFilesystem(dir string) error {
	return os.RemoveAll(dir)
}

func generateManifestContent(assetDir string, manifest map[string]string) (string, error) {

	// process our manifest in name order
	keys := slices.Collect(maps.Keys(manifest))
	slices.Sort(keys)

	content := ""
	var err error
	for _, name := range keys {
		fp := manifest[name]
		// do we need to calculate the fingerprint
		if len(fp) == 0 {
			fp, err = md5Checksum(filepath.Join(assetDir, name))
			if err != nil {
				return "", err
			}
		}
		content += fmt.Sprintf("%s %s\n", fp, name)
	}
	return content, nil
}

func specialCaseNameMapper(filename string, specialPrefix string) string {

	specialFiles := []string{payloadFilename, descriptionFileName, titleFileName, manifestFilename}
	if slices.Contains(specialFiles, filename) == true {
		newName := fmt.Sprintf("%s-%s", specialPrefix, filename)
		log.Printf("WARNING: mapping filename [%s] -> [%s]", filename, newName)
		return newName
	}
	return filename
}

//
// end of file
//
