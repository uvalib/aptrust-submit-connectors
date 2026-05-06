package main

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
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

	de, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, d := range de {
		err = os.RemoveAll(path.Join(dir, d.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

func generateManifestContent(assetDir string, manifest map[string]string) (string, error) {

	content := ""
	var err error
	for name, fp := range manifest {
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

//
// end of file
//
