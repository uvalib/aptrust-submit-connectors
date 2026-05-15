//
//
//

package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

var authorizationToken string
var xsrfEndpoint = "authn/status"
var authEndpoint = "authn/login"
var dspaceXsrfCookie = "DSPACE-XSRF-COOKIE"
var xsrfToken = "X-XSRF-TOKEN"

var ErrAuthenticationFailure = errors.New("authentication failure")

func authenticate(cfg *ServiceConfig) error {

	log.Printf("INFO: authenticating...")

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: time.Duration(cfg.HTTPTimeout) * time.Second,
	}

	// endpoints we will be using
	xsrfUrl := fmt.Sprintf("%s/%s", cfg.ApiEndpoint, xsrfEndpoint)
	authUrl := fmt.Sprintf("%s/%s", cfg.ApiEndpoint, authEndpoint)

	response, err := client.Get(xsrfUrl)
	if err != nil {
		log.Printf("ERROR: xsrfEndpoint returns error (%s)", err.Error())
		return ErrAuthenticationFailure
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Printf("ERROR: xsrfEndpoint returns %d (%s)", response.StatusCode, response.Status)
		return ErrAuthenticationFailure
	}

	// extract the xsrf cookie
	cookieStr := ""
	for _, cookie := range response.Cookies() {
		if cookie.Name == dspaceXsrfCookie {
			cookieStr = cookie.Value
			break
		}
	}

	if len(cookieStr) == 0 {
		log.Printf("ERROR: did not get %s token", dspaceXsrfCookie)
		return ErrAuthenticationFailure
	}

	headers := map[string]string{
		xsrfToken:      cookieStr,
		"Content-Type": "application/x-www-form-urlencoded",
	}

	formBody := url.Values{
		"user":     []string{cfg.ApiUser},
		"password": []string{cfg.ApiPassword},
	}
	formStr := formBody.Encode()
	reader := bytes.NewReader([]byte(formStr))

	//fmt.Printf("FORM [%s]\n", formStr)

	req, err := http.NewRequest("POST", authUrl, reader)
	if err != nil {
		log.Printf("ERROR: POST %s failed with error (%s)", authUrl, err)
		return ErrAuthenticationFailure
	}

	// add the needed headers
	for k, v := range headers {
		//fmt.Printf("Header [%s] = [%s]\n", k, v)
		req.Header.Add(k, v)
	}

	response, err = client.Do(req)
	if err != nil {
		log.Printf("ERROR: POST %s failed with error (%s)", authUrl, err)
		return ErrAuthenticationFailure
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Printf("ERROR: authEndpoint returns %d (%s)", response.StatusCode, response.Status)
		return ErrAuthenticationFailure
	}

	if len(response.Header["Authorization"]) != 0 {
		authorizationToken = response.Header["Authorization"][0]
	}

	if len(authorizationToken) == 0 {
		log.Printf("ERROR: did not get authorization header")
		return ErrAuthenticationFailure
	}

	log.Printf("INFO: authenticate success")
	return nil
}

func addAuthHeader(headers map[string]string) map[string]string {

	headers["Authorization"] = authorizationToken
	return headers
}

//
// end of file
//
