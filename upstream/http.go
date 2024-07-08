package upstream

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yiffyi/gorad/radhttp"
)

const baseURL = "http://dhwc.westlake.edu.cn"

var commonFormFields = map[string]string{
	"typeId":      "0",
	"phoneSystem": "WeChat",
	"version":     "5.0.0",
	"appId":       "wxc06f4dbb636009bb",
}

const REQUEST_TIMEOUT = 10 * time.Second

type Session struct {
	telPhone  string
	loginCode string
	timeId    string
	c         *http.Client
}

func CreateSession(username string, password string) (*Session, error) {
	s := Session{
		c: &http.Client{
			Timeout: REQUEST_TIMEOUT,
		},
	}

	pwdHash := md5.Sum([]byte(password))

	result, err := s.httpPost("/user/login", map[string]string{
		"telPhone": username,
		"password": strings.ToUpper(hex.EncodeToString(pwdHash[len(pwdHash)-5:])),
	})

	if err != nil {
		return nil, err
	}

	s.loginCode = result["data"].(map[string]interface{})["loginCode"].(string)
	s.telPhone = username

	return &s, nil
}

func CreateAnonymousSession() *Session {
	s := Session{
		c: &http.Client{
			Timeout: REQUEST_TIMEOUT,
		},
	}

	return &s
}

func (s *Session) httpPost(path string, data map[string]string) (map[string]interface{}, error) {
	body := url.Values{}
	for k, v := range commonFormFields {
		body.Set(k, v)
	}

	if len(s.loginCode) > 0 && len(s.telPhone) > 0 {
		body.Set("telPhone", s.telPhone)
		body.Set("loginCode", s.loginCode)
	}

	for k, v := range data {
		body.Set(k, v)
	}

	req, err := radhttp.NewURLEncodedFormRequest(baseURL+path, body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	resp, _, err := radhttp.DoAsJSON(s.c, req, &result)
	if err != nil {
		return nil, err
	}

	if err = radhttp.EnsureSuccessful(resp); err != nil {
		return result, err
	}

	if int(result["errorCode"].(float64)) != 0 {
		return result, errors.New(result["message"].(string))
	}

	return result, nil
}

func (s *Session) httpGet(path string, data map[string]string) (map[string]interface{}, error) {
	params := url.Values{}
	for k, v := range commonFormFields {
		params.Set(k, v)
	}

	if len(s.loginCode) > 0 && len(s.telPhone) > 0 {
		params.Set("telPhone", s.telPhone)
		params.Set("loginCode", s.loginCode)
	}

	for k, v := range data {
		params.Set(k, v)
	}

	req, err := radhttp.NewGetRequest(baseURL+path, params)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	resp, _, err := radhttp.DoAsJSON(s.c, req, &result)
	if err != nil {
		return nil, err
	}

	if err = radhttp.EnsureSuccessful(resp); err != nil {
		return result, err
	}

	if int(result["errorCode"].(float64)) != 0 {
		return result, errors.New(result["message"].(string))
	}

	return result, nil
}
