//
//
//

package main

type SearchResponse struct {
	Embedded SearchResults `json:"_embedded"`
}

type SearchResults struct {
	SearchResults EmbeddedObjects `json:"searchResult"`
}

type EmbeddedObjects struct {
	EmbeddedObjects Objects     `json:"_embedded"`
	ResultsPage     ResultsPage `json:"page"`
}

type Objects struct {
	Objects []EmbeddedObject `json:"objects"`
}

type EmbeddedObject struct {
	IndexableObject IndexableObject `json:"_embedded"`
}

type IndexableObject struct {
	Object Object `json:"indexableObject"`
}

type Object struct {
	Id   string `json:"id"`
	Uuid string `json:"uuid"`
}

type ResultsPage struct {
	PageNumber int `json:"number"`
	PageSize   int `json:"size"`
	TotalCount int `json:"totalElements"`
	TotalPages int `json:"totalPages"`
}

type ItemResponse struct {
	Metadata map[string][]ItemMetadataFields `json:"metadata"`
	Links    map[string]LinkHref             `json:"_links"`
}

type ItemMetadataFields struct {
	Value string `json:"value"`
}

type LinkHref struct {
	Href string `json:"href"`
}

type BundlesResponse struct {
	EmbeddedBundles Bundles `json:"_embedded"`
}

type Bundles struct {
	Bundles []Bundle `json:"bundles"`
}

type Bundle struct {
	Name  string              `json:"name"`
	Links map[string]LinkHref `json:"_links"`
}

type BitstreamsResponse struct {
	EmbeddedBitstreams Bitstreams `json:"_embedded"`
}

type Bitstreams struct {
	Bitstreams []Bitstream `json:"bitstreams"`
}

type Bitstream struct {
	Name     string              `json:"name"`
	Size     int64               `json:"sizeBytes"`
	Checksum BitstreamChecksum   `json:"checkSum"`
	Links    map[string]LinkHref `json:"_links"`
}

type BitstreamChecksum struct {
	Algorithm string `json:"checkSumAlgorithm"`
	Value     string `json:"value"`
}

var titleMetadataFieldName = "dc.title"
var descriptionMetadataFieldName = "dc.description.abstract"
var identifierMetadataFieldName = "dc.identifier"

//
// end of file
//
