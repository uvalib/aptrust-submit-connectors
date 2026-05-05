//
//
//

package main

type UpdatedResponse struct {
	Status string      `json:"status"`
	Data   UpdatedData `json:"data"`
}

type UpdatedData struct {
	Start      int           `json:"start"`
	TotalCount int           `json:"total_count"`
	Items      []UpdatedItem `json:"items"`
}

type UpdatedItem struct {
	GlobalId string `json:"global_id"`
	EntityId int    `json:"entity_id"`
}

type ItemResponse struct {
	Status string   `json:"status"`
	Data   ItemData `json:"data"`
}

type ItemData struct {
	Identifier string      `json:"identifier"`
	Authority  string      `json:"authority"`
	Protocol   string      `json:"protocol"`
	Latest     ItemVersion `json:"latestVersion"`
}

type ItemVersion struct {
	VersionMajor int         `json:"versionNumber"`
	VersionMinor int         `json:"versionMinorNumber"`
	Files        []ItemFiles `json:"files"`
}

type ItemFiles struct {
	File ItemDataFile `json:"dataFile"`
}

type ItemDataFile struct {
	Id       int64  `json:"id"`
	Filename string `json:"filename"`
	Filesize int64  `json:"filesize"`
	MD5      string `json:"md5"`
}

//
// end of file
//
