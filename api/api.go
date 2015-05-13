package api

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bitly/go-simplejson"
)

func Request(req *http.Request) (*simplejson.Json, error) {
	httpclient := &http.Client{}
	resp, err := httpclient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		log.Printf("got response code %d - %s", resp.StatusCode, body)
		return nil, errors.New("api request returned non 200 status code")
	}
	data, err := simplejson.NewJson(body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func RequestUsingAccessTokenParameter(url string, access_token string) (
	response *http.Response, err error) {
	req_url := url + "?access_token=" + access_token
	req, err := http.NewRequest("GET", req_url, nil)
	if err != nil {
		return nil, errors.New("failed building request for " +
			req_url + ": " + err.Error())
	}

	httpclient := &http.Client{}
	if response, err = httpclient.Do(req); err != nil {
		return nil, errors.New("request failed for " +
			req_url + ": " + err.Error())
	}
	return
}
