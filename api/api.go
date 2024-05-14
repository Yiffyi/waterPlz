package api

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const baseURL = "http://dhwc.westlake.edu.cn"

var commonFormFields = map[string]string{
	"typeId":      "0",
	"phoneSystem": "WeChat",
	"version":     "5.0.0",
	"appId":       "wxc06f4dbb636009bb",
}

type Session struct {
	telPhone  string
	loginCode string
	timeId    string
	c         *http.Client
}

func CreateSession(username string, password string) (*Session, error) {
	s := Session{
		c: &http.Client{},
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

	req, err := http.NewRequest(http.MethodPost, baseURL+path, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, err
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

	u, err := url.ParseRequestURI(baseURL)
	if err != nil {
		panic(err)
	}
	u.Path = path
	u.RawQuery = params.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, err
	}

	if int(result["errorCode"].(float64)) != 0 {
		return result, errors.New(result["message"].(string))
	}

	return result, nil
}

func (s *Session) UserInfo() (map[string]interface{}, error) {
	result, err := s.httpGet("/user/info", nil)
	if err != nil {
		return nil, err
	}

	return result["data"].(map[string]interface{}), nil
}

func (s *Session) CreateOrder(SN string) error {
	result, err := s.httpPost("/order/downRate/snWater", map[string]string{
		"deviceSncode": SN,
	})

	fmt.Println(result)
	if err != nil {
		if result != nil {
			if int(result["errorCode"].(float64)) == 307 {
				s.timeId = result["data"].(map[string]interface{})["timeIds"].(string)
				return err
			} else {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (s *Session) CloseOrder(SN string) error {
	_, err := s.httpPost("/order/send/closeOrder", map[string]string{
		"deviceSncode": SN,
		"timeIds":      s.timeId,
	})

	return err
}
