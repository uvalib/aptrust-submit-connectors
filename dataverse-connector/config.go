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

}

// loadConfiguration will load the service configuration from env/cmdline
// and return a pointer to it. Any failures are fatal.
func loadConfiguration() *ServiceConfig {

	var cfg ServiceConfig

	// APTrust submission service configuration
	cfg.APTServiceRegister = ensureSetAndNonEmpty("APT_REGISTER_URL")
	cfg.APTServiceSubmit = ensureSetAndNonEmpty("APT_SUBMIT_URL")
	cfg.APTServiceClient = ensureSetAndNonEmpty("APT_CLIENT_ID")

	// APTrust submission service configuration
	log.Printf("[CONFIG] APTServiceRegister = [%s]", cfg.APTServiceRegister)
	log.Printf("[CONFIG] APTServiceSubmit   = [%s]", cfg.APTServiceSubmit)
	log.Printf("[CONFIG] APTServiceClient   = [%s]", cfg.APTServiceClient)

	return &cfg
}
