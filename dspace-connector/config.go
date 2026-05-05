package main

import (
	"log"
)

// ServiceConfig defines all the service configuration parameters
type ServiceConfig struct {

	// APTrust submission service configuration
	APTServiceRegister string // url for APTrust submission registration
	APTServiceSubmit   string // url for APTrust submit
	APTServiceClient   string // client identifier for APTrust submit

	// API configuration
	ApiEndpoint string
	//ApiToken            string
	ApiUpdatedPathQuery string
	ApiItemPathQuery    string
	//ApiFilePathQuery    string

	// other configuration
	ScratchFileSystem string
	HTTPTimeout       int // in seconds
}

// loadConfiguration will load the service configuration from env/cmdline
// and return a pointer to it. Any failures are fatal.
func loadConfiguration() *ServiceConfig {

	var cfg ServiceConfig

	// APTrust submission service configuration
	cfg.APTServiceRegister = ensureSetAndNonEmpty("APT_REGISTER_URL")
	cfg.APTServiceSubmit = ensureSetAndNonEmpty("APT_SUBMIT_URL")
	cfg.APTServiceClient = ensureSetAndNonEmpty("APT_CLIENT_ID")

	// API configuration
	cfg.ApiEndpoint = ensureSetAndNonEmpty("API_ENDPOINT")
	//cfg.ApiToken = ensureSetAndNonEmpty("API_TOKEN")
	cfg.ApiUpdatedPathQuery = ensureSetAndNonEmpty("API_UPDATED_PATH")
	cfg.ApiItemPathQuery = ensureSetAndNonEmpty("API_ITEM_PATH")
	//cfg.ApiFilePathQuery = ensureSetAndNonEmpty("API_FILE_PATH")

	// other configuration
	cfg.ScratchFileSystem = ensureSetAndNonEmpty("SCRATCH_FS")
	cfg.HTTPTimeout = envToInt("HTTP_TIMEOUT")

	// APTrust submission service configuration
	log.Printf("[CONFIG] APTServiceRegister  = [%s]", cfg.APTServiceRegister)
	log.Printf("[CONFIG] APTServiceSubmit    = [%s]", cfg.APTServiceSubmit)
	log.Printf("[CONFIG] APTServiceClient    = [%s]", cfg.APTServiceClient)

	// API configuration
	log.Printf("[CONFIG] ApiEndpoint         = [%s]", cfg.ApiEndpoint)
	//log.Printf("[CONFIG] ApiToken            = [REDACTED]")
	log.Printf("[CONFIG] ApiUpdatedPathQuery = [%s]", cfg.ApiUpdatedPathQuery)
	log.Printf("[CONFIG] ApiItemPathQuery    = [%s]", cfg.ApiItemPathQuery)
	//log.Printf("[CONFIG] ApiFilePathQuery    = [%s]", cfg.ApiFilePathQuery)

	// other configuration
	log.Printf("[CONFIG] ScratchFileSystem   = [%s]", cfg.ScratchFileSystem)
	log.Printf("[CONFIG] HTTPTimeout         = [%d]", cfg.HTTPTimeout)

	return &cfg
}
