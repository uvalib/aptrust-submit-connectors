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
	Dryrun    bool
}

// main entry point
func main() {

	// process the commandline options
	opts := ServiceOptions{}
	flag.StringVar(&opts.StartDate, "startdate", "", "The object set start date")
	flag.StringVar(&opts.EndDate, "enddate", "", "The object set end date")
	flag.StringVar(&opts.SingleId, "singleid", "", "Submit a single object")
	flag.BoolVar(&opts.Dryrun, "dryrun", false, "Show but dont do")

	flag.Parse()

	// verify we have specified something sensible...
	if len(opts.StartDate) == 0 && len(opts.EndDate) == 0 && len(opts.SingleId) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if len(opts.SingleId) != 0 && (len(opts.StartDate) != 0 || len(opts.EndDate) != 0) {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if len(opts.SingleId) == 0 && (len(opts.StartDate) == 0 || len(opts.EndDate) == 0) {
		flag.PrintDefaults()
		os.Exit(1)
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
		var ids []string
		ids, err = getIdList(cfg, &opts)
		if err == nil {
			idList = append(idList, ids...)
		}
	}

	// process the set of id's
	if len(idList) != 0 {
		err = processIdList(cfg, &opts, idList)
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
