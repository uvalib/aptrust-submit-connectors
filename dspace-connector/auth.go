//
// based mainly on information here: https://github.com/DSpace/RestContract/blob/main/authentication.md
//

package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

// our active authorization token
var authorizationToken string
var xsrfToken string
var authorizationTokenRenewTime = time.Now()
var authorizationTokenLife = 20 * time.Minute
var authorizationTokenRenewUrl string

var xsrfEndpoint = "authn/status"
var authEndpoint = "authn/login"
var dspaceXsrfCookie = "DSPACE-XSRF-COOKIE"
var xsrfTokenName = "X-XSRF-TOKEN"

var ErrAuthenticationFailure = errors.New("authentication failure")

func authenticate(cfg *ServiceConfig) error {

	log.Printf("INFO: authenticating...")

	client := &http.Client{}

	// endpoints we will be using
	xsrfUrl := fmt.Sprintf("%s/%s", cfg.ApiEndpoint, xsrfEndpoint)
	authorizationTokenRenewUrl = fmt.Sprintf("%s/%s", cfg.ApiEndpoint, authEndpoint)

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
	for _, cookie := range response.Cookies() {
		if cookie.Name == dspaceXsrfCookie {
			xsrfToken = cookie.Value
			break
		}
	}

	if len(xsrfToken) == 0 {
		log.Printf("ERROR: did not get %s token", dspaceXsrfCookie)
		return ErrAuthenticationFailure
	}

	headers := map[string]string{
		xsrfTokenName:  xsrfToken,
		"Content-Type": "application/x-www-form-urlencoded",
	}

	formBody := url.Values{
		"user":     []string{cfg.ApiUser},
		"password": []string{cfg.ApiPassword},
	}
	formStr := formBody.Encode()
	reader := bytes.NewReader([]byte(formStr))

	//fmt.Printf("FORM [%s]\n", formStr)

	req, err := http.NewRequest("POST", authorizationTokenRenewUrl, reader)
	if err != nil {
		log.Printf("ERROR: POST %s failed with error (%s)", authorizationTokenRenewUrl, err.Error())
		return ErrAuthenticationFailure
	}

	// add the needed headers
	for k, v := range headers {
		//fmt.Printf("Header [%s] = [%s]\n", k, v)
		req.Header.Add(k, v)
	}

	// add the expected cookie
	cookie := http.Cookie{
		Name:     dspaceXsrfCookie,
		Value:    xsrfToken,
		MaxAge:   3600,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	}
	req.AddCookie(&cookie)

	response, err = client.Do(req)
	if err != nil {
		log.Printf("ERROR: POST %s failed with error (%s)", authorizationTokenRenewUrl, err.Error())
		return ErrAuthenticationFailure
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Printf("ERROR: authEndpoint returns %d", response.StatusCode)
		return ErrAuthenticationFailure
	}

	if len(response.Header["Authorization"]) != 0 {
		authorizationToken = response.Header["Authorization"][0]
	}

	if len(authorizationToken) == 0 {
		log.Printf("ERROR: did not get authorization header")
		return ErrAuthenticationFailure
	}

	authorizationTokenRenewTime = time.Now().Add(authorizationTokenLife)
	log.Printf("INFO: authenticate success")
	return nil
}

// providing a valid auth token that might expire soon (default 30 minutes) to generate a
// fresh new one
func renewAuthToken() error {

	log.Printf("INFO: renewing authorization token...")

	req, err := http.NewRequest("POST", authorizationTokenRenewUrl, nil)
	if err != nil {
		log.Printf("ERROR: POST %s failed with error (%s)", authorizationTokenRenewUrl, err.Error())
		return ErrAuthenticationFailure
	}

	// add the needed headers
	req.Header.Add(xsrfTokenName, xsrfToken)
	req.Header.Add("Authorization", authorizationToken)

	// add the expected cookie
	cookie := http.Cookie{
		Name:     dspaceXsrfCookie,
		Value:    xsrfToken,
		MaxAge:   3600,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	}
	req.AddCookie(&cookie)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Printf("ERROR: POST %s failed with error (%s)", authorizationTokenRenewUrl, err.Error())
		return ErrAuthenticationFailure
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Printf("ERROR: authEndpoint returns %d", response.StatusCode)
		return ErrAuthenticationFailure
	}

	if len(response.Header["Authorization"]) != 0 {
		authorizationToken = response.Header["Authorization"][0]
	}

	if len(authorizationToken) == 0 {
		log.Printf("ERROR: did not get authorization header")
		return ErrAuthenticationFailure
	}

	authorizationTokenRenewTime = time.Now().Add(authorizationTokenLife)
	log.Printf("INFO: renew auth token success")
	return nil
}

func addAuthHeader(headers map[string]string) map[string]string {

	// time to renew the auth token
	if authorizationTokenRenewTime.Before(time.Now()) == true {
		_ = renewAuthToken()
	}

	headers["Authorization"] = authorizationToken
	return headers
}

//
// end of file
//
