package main

import (
	"flag"
	"log"
	"os"
)

type ServiceOptions struct {
	StartDate string
	EndDate   string
	SingleId  string
	ResumeId  string
	NoProcess bool
	NoFiles   bool
	NoSubmit  bool
}

// main entry point
func main() {

	// process the commandline options
	opts := ServiceOptions{}
	flag.StringVar(&opts.StartDate, "startdate", "", "The object set start date, YYYY-MM-DD")
	flag.StringVar(&opts.EndDate, "enddate", "", "The object set end date, YYYY-MM-DD")
	flag.StringVar(&opts.SingleId, "singleid", "", "Submit a single object")
	flag.StringVar(&opts.ResumeId, "resumeid", "", "Resume at this id in the object set")
	flag.BoolVar(&opts.NoProcess, "noprocess", false, "Count but dont generate")
	flag.BoolVar(&opts.NoFiles, "nofiles", false, "No file download")
	flag.BoolVar(&opts.NoSubmit, "nosubmit", false, "Generate but dont submit")

	flag.Parse()

	// verify we have specified something sensible...
	if len(opts.StartDate) == 0 && len(opts.EndDate) == 0 && len(opts.SingleId) == 0 && len(opts.ResumeId) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if len(opts.SingleId) != 0 && (len(opts.StartDate) != 0 || len(opts.EndDate) != 0) {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if len(opts.SingleId) == 0 && (len(opts.StartDate) != 10 || len(opts.EndDate) != 10) {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if len(opts.ResumeId) != 0 && (len(opts.StartDate) != 10 || len(opts.EndDate) != 10) {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// no submitting to APT if we do not download files
	if opts.NoFiles == true {
		opts.NoSubmit = true
	}

	// commandline options look good, issue startup message
	log.Printf("===> %s service staring up (version: %s) <===", os.Args[0], Version())

	// Get config params and use them to init service context. Any issues are fatal
	cfg := loadConfiguration()

	// get or construct the set of ID's to be processed
	idList := make([]string, 0)
	var err error
	if len(opts.SingleId) != 0 {
		idList = append(idList, opts.SingleId)
	} else {
		idList, err = getUpdatedIds(cfg, &opts)
		log.Printf("INFO: %d updated item(s) reported", len(idList))
	}

	// process the set of id's (if appropriate)
	if opts.NoProcess == false && len(idList) != 0 {
		err = processUpdatedIds(cfg, &opts, idList)
	} else {
		for ix, id := range idList {
			log.Printf("INFO: %d of %d ==> %s", ix+1, len(idList), id)
		}

	}

	if err == nil {
		log.Printf("INFO: terminate normally")
	} else {
		log.Printf("ERROR: terminate with '%s'", err.Error())
	}

}

//
// end of file
//
