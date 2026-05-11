package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

var maxHttpRetries = 3
var httpRetrySleepTime = 100 * time.Millisecond

func newHttpClient(maxConnections int, timeout int) *http.Client {

	defaultTransport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 15 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
		MaxIdleConns:        maxConnections,
		MaxIdleConnsPerHost: maxConnections,
	}

	return &http.Client{
		Transport: defaultTransport,
		Timeout:   time.Duration(timeout) * time.Second,
	}
}

func httpGet(client *http.Client, url string, headers map[string]string) ([]byte, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("ERROR: GET %s failed with error (%s)", url, err)
		return nil, err
	}

	// add the needed headers
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	return httpSend(client, req)
}

func httpPost(client *http.Client, url string, payload []byte, headers map[string]string) ([]byte, error) {

	reader := bytes.NewReader(payload)
	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		log.Printf("ERROR: POST %s failed with error (%s)", url, err)
		return nil, err
	}

	// add the needed headers
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	return httpSend(client, req)
}

func httpSend(client *http.Client, req *http.Request) ([]byte, error) {

	var response *http.Response
	var err error
	url := req.URL.String()
	count := 0
	for {
		//start := time.Now()
		response, err = client.Do(req)
		//duration := time.Since(start)
		//log.Printf("INFO: %s %s (elapsed %0.2f seconds)", req.Method, url, duration.Seconds())

		count++
		if err != nil {
			if canRetry(err) == false {
				log.Printf("ERROR: %s %s failed with error (%s)", req.Method, url, err)
				return nil, err
			}

			// break when tried too many times
			if count >= maxHttpRetries {
				return nil, err
			}

			log.Printf("ERROR: %s %s failed with error, retrying (%s)", req.Method, url, err)

			// sleep for a bit before retrying
			time.Sleep(httpRetrySleepTime)
		} else {

			defer response.Body.Close()

			if response.StatusCode >= 300 {
				logLevel := "ERROR"
				// log StatusNotFound as informational instead of as an error
				if response.StatusCode == http.StatusNotFound {
					logLevel = "INFO"
				}
				log.Printf("%s: %s %s failed with status %d", logLevel, req.Method, url, response.StatusCode)

				body, _ := io.ReadAll(response.Body)

				return body, fmt.Errorf("request returns HTTP %d", response.StatusCode)
			}

			body, err := io.ReadAll(response.Body)
			if err != nil {
				log.Printf("ERROR: %s %s failed with error (%s)", req.Method, url, err)
				return nil, err
			}
			//fmt.Printf( body )
			return body, nil
		}
	}
}

// examines the error and decides if it can be retried
func canRetry(err error) bool {

	if strings.Contains(err.Error(), "operation timed out") == true {
		return true
	}

	if strings.Contains(err.Error(), "Client.Timeout") == true {
		return true
	}

	if strings.Contains(err.Error(), "write: broken pipe") == true {
		return true
	}

	if strings.Contains(err.Error(), "no such host") == true {
		return true
	}

	if strings.Contains(err.Error(), "network is down") == true {
		return true
	}

	if strings.Contains(err.Error(), "TLS handshake timeout") == true {
		return true
	}

	return false
}

//
// end of file
//
