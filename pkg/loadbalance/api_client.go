package loadbalance

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"strconv"
)

type CreateOptsBuilder interface {
	ToListenerCreateMap() (map[string]interface{}, error)
}

type keycloakToken struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int32  `json:"expires_in"`
	RefreshExpiresIn int32  `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int32  `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
}

func getKeyCloakToken(requestedSubject, tokenClientId, clientSecret, keycloakUrl string) (string, error) {
	var grantType = "urn:ietf:params:oauth:grant-type:token-exchange"
	var requestTokenType = "urn:ietf:params:oauth:token-type:refresh_token"
	var audience = "console"
	strReq := "grant_type=" + grantType + "&client_id=" + tokenClientId + "&client_secret=" + clientSecret +
		"&request_token_type=" + requestTokenType + "&requested_subject=" + requestedSubject + "&audience=" + audience
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	res, err1 := client.Post(keycloakUrl, "application/x-www-form-urlencoded", strings.NewReader(strReq))
	if res != nil && res.StatusCode == 200 {
		defer res.Body.Close()
		body, err2 := ioutil.ReadAll(res.Body)
		if err2 == nil {
			var token keycloakToken
			if err3 := json.Unmarshal(body, &token); err3 == nil {
				glog.Info("token is " + token.AccessToken)
				return "Bearer " + token.AccessToken, nil
			} else {
				glog.Errorf("error to Unmarshal(body, &token): %v", err3)
				return "", err3
			}
		} else {
			glog.Errorf("error to read all res1.Body %v", err2)
			return "", err2
		}
	} else {
		if err1 != nil {
			glog.Errorf("post request keycloak err: %v", err1)
			return "", err1
		}
		if res != nil {
			glog.Errorf("nil res or not ok status code, code: %d", res.StatusCode)
			defer res.Body.Close()
		}
		return "", errors.New("nil res or not ok status code")
	}
}

//http://cn-north-3.10.110.25.123.xip.io/slb/v1/slbs?slbId=123
//按slb id查询用户的slb
func describeLoadBalancer(url, token, slbId string) (*LoadBalancer, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	reqUrl := url + "?slbId=" + slbId
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		glog.Errorf("Request error %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("Response error %v", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("Get response body fail %v", err)
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		glog.Errorf("response not ok %v", res.StatusCode)
		return nil, fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result []LoadBalancer
	err = xml.Unmarshal(body, &result)
	if err != nil {
		glog.Errorf("Unmarshal body fail: %v", err)
		return nil, err
	}
	return &result[0], nil
}

func modifyLoadBalancer(url, token, slbId , slbName string)(*SlbResponse,error){
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	reqUrl := url + "/" + slbId
	requestMap := make(map[string]string)
	requestMap["slbName"] = slbName
	slbNameByte,err := json.Marshal(&requestMap)
	if nil != err {
		glog.Errorf("servers conver to bytes error %v", err)
		return nil,err
	}
	req, err := http.NewRequest("PUT", reqUrl, bytes.NewReader(slbNameByte))
	if err != nil {
		glog.Errorf("Request error %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("Response error %v", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("Get response body fail %v", err)
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		glog.Errorf("response not ok %v", res.StatusCode)
		return nil, fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result SlbResponse
	err = xml.Unmarshal(body, &result)
	if err != nil {
		glog.Errorf("Unmarshal body fail: %v", err)
		return nil, err
	}
	return &result, nil
}

func deleteLoadBalancer(url, token, slbId string)error{
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	reqUrl := url + "/" + slbId
	req, err := http.NewRequest("DELETE", reqUrl, nil)
	if err != nil {
		glog.Errorf("Request error %v", err)
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("Response error %v", err)
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("Get response body fail %v", err)
		return err
	}
	if res.StatusCode != http.StatusOK {
		glog.Errorf("response not ok %v", res.StatusCode)
		return fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result BackendList
	err = xml.Unmarshal(body, &result)
	if err != nil {
		glog.Errorf("Unmarshal body fail: %v", err)
		return  err
	}
	if result.code != strconv.Itoa(http.StatusAccepted) {
		return errors.New("deleteLb fail,"+result.Message)
	}
	return nil
}

func describeListenersBySlbId(url, token, slbId string) ([]Listener, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	reqUrl := url + "/" + slbId + "/listeners"
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		glog.Errorf("Request error %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("Response error %v", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("Get response body fail %v", err)
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		glog.Errorf("response not ok %v", res.StatusCode)
		return nil, fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result []Listener
	err = xml.Unmarshal(body, &result)
	if err != nil {
		glog.Errorf("Unmarshal body fail: %v", err)
		return nil, err
	}
	return result, nil
}

func describeListenerByListnerId(url, token, slbId, listnerId string) (*Listener, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	reqUrl := url + "/" + slbId + "/listeners/" + listnerId
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		glog.Errorf("Request error %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("Response error %v", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("Get response body fail %v", err)
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		glog.Errorf("response not ok %v", res.StatusCode)
		return nil, fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result Listener
	err = xml.Unmarshal(body, &result)
	if err != nil {
		glog.Errorf("Unmarshal body fail: %v", err)
		return nil, err
	}
	return &result, nil
}

func createListener(url, token string, opts CreateListenerOpts) (*Listener, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	reqUrl := url + "/" + opts.SLBId+"/listeners/"
	serversByte,err := json.Marshal(&opts)
	if nil != err {
		glog.Errorf("opts conver to bytes error %v", err)
		return nil,err
	}
	req, err := http.NewRequest("POST", reqUrl, bytes.NewReader(serversByte))
	if err != nil {
		glog.Errorf("Request error %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("Response error %v", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("Get response body fail %v", err)
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		glog.Errorf("response not ok %v", res.StatusCode)
		return nil, fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result Listener
	err = xml.Unmarshal(body, &result)
	if err != nil {
		glog.Errorf("Unmarshal body fail: %v", err)
		return nil, err
	}
	return &result, nil
}

func modifyListener(url, token, listenerid string, opts CreateListenerOpts) (*Listener, error) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	reqUrl := url + "/" + opts.SLBId+"/listeners/"+listenerid
	serversByte,err := json.Marshal(&opts)
	if nil != err {
		glog.Errorf("opts conver to bytes error %v", err)
		return nil,err
	}
	req, err := http.NewRequest("PUT", reqUrl, bytes.NewReader(serversByte))
	if err != nil {
		glog.Errorf("Request error %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("Response error %v", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("Get response body fail %v", err)
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		glog.Errorf("response not ok %v", res.StatusCode)
		return nil, fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result Listener
	err = xml.Unmarshal(body, &result)
	if err != nil {
		glog.Errorf("Unmarshal body fail: %v", err)
		return nil, err
	}
	return &result, nil
}

func deleteListener(url, token, slbId, listnerId string)error{
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	reqUrl := url + "/" + slbId + "/listeners/" + listnerId
	req, err := http.NewRequest("DELETE", reqUrl, nil)
	if err != nil {
		glog.Errorf("Request error %v", err)
		return  err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("Response error %v", err)
		return  err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("Get response body fail %v", err)
		return  err
	}
	if res.StatusCode != http.StatusOK {
		glog.Errorf("response not ok %v", res.StatusCode)
		return  fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result BackendList
	err = xml.Unmarshal(body, &result)
	if err != nil {
		glog.Errorf("Unmarshal body fail: %v", err)
		return  err
	}
	if result.code != strconv.Itoa(http.StatusNoContent) {
		glog.Errorf("delete listener fail: %v",result.Message)
		return errors.New(result.Message)
	}
	return nil
}

func createBackend(url, token string, opts CreateBackendOpts) (*BackendList, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	reqUrl := url + "/" + opts.SLBId + "/listeners/" + opts.ListenerId + "/members"
	serversByte, err := json.Marshal(&opts.Servers)
	if nil != err {
		glog.Errorf("servers conver to bytes error %v", err)
		return nil, err
	}
	req, err := http.NewRequest("POST", reqUrl, bytes.NewReader(serversByte))
	if err != nil {
		glog.Errorf("Request error %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("Response error %v", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("Get response body fail %v", err)
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		glog.Errorf("response not ok %v", res.StatusCode)
		return nil, fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result BackendList
	err = xml.Unmarshal(body, &result)
	if err != nil {
		glog.Errorf("Unmarshal body fail: %v", err)
		return nil, err
	}
	return &result, nil
}

func describeBackendservers(url, token, slbId, listnerId string) ([]Backend, error) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	reqUrl := url + "/" + slbId + "/listeners/" + listnerId + "/members"
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		glog.Errorf("Request error %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("Response error %v", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("Get response body fail %v", err)
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		glog.Errorf("response not ok %v", res.StatusCode)
		return nil, fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result []Backend
	err = xml.Unmarshal(body, &result)
	if err != nil {
		glog.Errorf("Unmarshal body fail: %v", err)
		return nil, err
	}
	return result, nil

}

func removeBackendServers(url, token, slbId , listnerId string, backendIdList []string) error {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	backendByte,err := json.Marshal(&backendIdList)
	if err != nil {
		glog.Errorf("parse json error %v", err)
		return err
	}
	reqUrl := url + "/" + slbId + "/listeners/" + listnerId + "/members" + "?backendIdList=" + string(backendByte)
	req, err := http.NewRequest("DELETE", reqUrl, nil)
	if err != nil {
		glog.Errorf("Request error %v", err)
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("Response error %v", err)
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("Get response body fail %v", err)
		return  err
	}
	if res.StatusCode != http.StatusOK {
		glog.Errorf("response not ok %v", res.StatusCode)
		return fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result BackendList
	err = xml.Unmarshal(body, &result)
	if err != nil {
		glog.Errorf("Unmarshal body fail: %v", err)
		return  err
	}
	if result.code != strconv.Itoa(http.StatusOK) {
		glog.Errorf("Delete backend fail: %v", result.Message)
		return errors.New(result.Message)
	}
	return nil
}
