//
//
//

package main

type UpdatedResponse []string

type ItemResponse struct {
	Item    ItemData       `json:"item"`
	Content []ItemDataFile `json:"content"`
}

type ItemData struct {
	Metadata map[string][]ItemMetadataFields `json:"metadata"`
}

type ItemMetadataFields struct {
	Value string `json:"value"`
}

type ItemDataFile struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

var titleMetadataFieldName = "dc.title"
var descriptionMetadataFieldName = "dc.description.abstract"

//
// end of file
//
