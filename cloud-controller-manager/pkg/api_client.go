package pkg

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"k8s.io/klog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
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

func getKeyCloakToken(requestedSubject, tokenClientId, clientSecret, keycloakUrl string, ic *InCloud) (string, error) {
	//return ic.KeycloakToken, nil
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
				return "Bearer " + token.AccessToken, nil
			} else {
				klog.Errorf("error to Unmarshal(body, &token): %v", err3)
				return "", err3
			}
		} else {
			klog.Errorf("error to read all res1.Body %v", err2)
			return "", err2
		}
	} else {
		if err1 != nil {
			klog.Errorf("post request keycloak err: %v", err1)
			return "", err1
		}
		if res != nil {
			klog.Errorf("nil res or not ok status code, code: %d", res.StatusCode)
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
	klog.Infof("describeLoadBalancer requestUrl is %v,token is %v", reqUrl, token)
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		klog.Errorf("Request error %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		klog.Errorf("Response error %v", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		klog.Errorf("Get response body fail %v", err)
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		klog.Errorf("response not ok:%v,%v", res.StatusCode, string(body))
		return nil, fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result []LoadBalancer
	klog.Infof("result is:%v,%v ", res.StatusCode, string(body))
	err = json.Unmarshal(body, &result)
	if err != nil {
		klog.Errorf("Unmarshal body fail: %v", err)
		return nil, err
	}
	if nil != result && len(result) > 0 {
		return &result[0], nil
	} else {
		return nil, errors.New("lb is empty")
	}
}

func modifyLoadBalancer(url, token, slbId, slbName string) (*SlbResponse, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	reqUrl := url + "/" + slbId
	requestMap := make(map[string]string)
	requestMap["slbName"] = slbName
	slbNameByte, err := json.Marshal(&requestMap)
	if nil != err {
		klog.Errorf("servers conver to bytes error %v", err)
		return nil, err
	}
	klog.Infof("modifyLoadBalancer requestUrl is %v,requestBody is%v, token is %v", reqUrl, string(slbNameByte), token)
	req, err := http.NewRequest("PUT", reqUrl, bytes.NewReader(slbNameByte))
	if err != nil {
		klog.Errorf("Request error %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		klog.Errorf("Response error %v", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		klog.Errorf("Get response body fail %v", err)
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		klog.Errorf("response not ok:%v,%v", res.StatusCode, string(body))
		return nil, fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result SlbResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		klog.Errorf("Unmarshal body fail: %v", err)
		return nil, err
	}
	return &result, nil
}

func deleteLoadBalancer(url, token, slbId string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	reqUrl := url + "/" + slbId
	klog.Infof("deleteLoadBalancer requestUrl is %v,token is%v", reqUrl, token)
	req, err := http.NewRequest("DELETE", reqUrl, nil)
	if err != nil {
		klog.Errorf("Request error %v", err)
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		klog.Errorf("Response error %v", err)
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		klog.Errorf("Get response body fail %v", err)
		return err
	}
	if res.StatusCode != http.StatusAccepted {
		klog.Errorf("response not ok:%v,%v", res.StatusCode, string(body))
		return fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result BackendList
	err = json.Unmarshal(body, &result)
	if err != nil {
		klog.Errorf("Unmarshal body fail: %v", err)
		return err
	}
	if result.code != strconv.Itoa(http.StatusAccepted) {
		return errors.New("deleteLb fail," + result.Message)
	}
	return nil
}

func describeListenersBySlbId(url, token, slbId string) ([]Listener, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	reqUrl := url + "/" + slbId + "/listeners"
	klog.Infof("describeListenersBySlbId requestUrl is %v,token is %v", reqUrl, token)
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		klog.Errorf("Request error %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		klog.Errorf("Response error %v", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		klog.Errorf("Get response body fail %v", err)
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		klog.Errorf("response not ok:%v,%v", res.StatusCode, string(body))
		return nil, fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result []Listener
	err = json.Unmarshal(body, &result)
	if err != nil {
		klog.Errorf("Unmarshal body fail: %v", err)
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
	klog.Infof("describeListenerByListnerId requestUrl is %v,token is %v", reqUrl, token)
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		klog.Errorf("Request error %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		klog.Errorf("Response error %v", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		klog.Errorf("Get response body fail %v", err)
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		klog.Errorf("response not ok:%v,%v", res.StatusCode, string(body))
		return nil, fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result Listener
	err = json.Unmarshal(body, &result)
	if err != nil {
		klog.Errorf("Unmarshal body fail: %v", err)
		return nil, err
	}
	return &result, nil
}

func createListener(url, token string, opts CreateListenerOpts) (*Listener, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	reqUrl := url + "/" + opts.SLBId + "/listeners/"
	klog.Infof("createListener requestUrl:%v,token:%v", reqUrl, token)
	serversByte, err := json.Marshal(&opts)
	if nil != err {
		klog.Errorf("opts conver to bytes error %v", err)
		return nil, err
	}
	klog.Infof("requestBody is : %v", string(serversByte))
	req, err := http.NewRequest("POST", reqUrl, bytes.NewReader(serversByte))
	if err != nil {
		klog.Errorf("Request error %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		klog.Errorf("Response error %v", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		klog.Errorf("Get response body fail %v", err)
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		klog.Errorf("response not ok:%v,%v", res.StatusCode, string(body))
		return nil, fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result Listener
	err = json.Unmarshal(body, &result)
	if err != nil {
		klog.Errorf("Unmarshal body fail: %v", err)
		return nil, err
	}
	return &result, nil
}

func modifyListener(url, token, listenerid string, opts CreateListenerOpts) (*Listener, error) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	reqUrl := url + "/" + opts.SLBId + "/listeners/" + listenerid
	klog.Infof("modifyListener requestUrl:%v,token%v", reqUrl, token)
	serversByte, err := json.Marshal(&opts)
	if nil != err {
		klog.Errorf("opts conver to bytes error %v", err)
		return nil, err
	}
	klog.Infof("requestBody is : %v", string(serversByte))
	req, err := http.NewRequest("PUT", reqUrl, bytes.NewReader(serversByte))
	if err != nil {
		klog.Errorf("Request error %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		klog.Errorf("Response error %v", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		klog.Errorf("Get response body fail %v", err)
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		klog.Errorf("response not ok:%v,%v", res.StatusCode, string(body))
		return nil, fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result Listener
	err = json.Unmarshal(body, &result)
	if err != nil {
		klog.Errorf("Unmarshal body fail: %v", err)
		return nil, err
	}
	return &result, nil
}

func deleteListener(url, token, slbId, listnerId string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	reqUrl := url + "/" + slbId + "/listeners/" + listnerId
	klog.Infof("deleteListener requestUrl:%v,token:%v", reqUrl, token)
	req, err := http.NewRequest("DELETE", reqUrl, nil)
	if err != nil {
		klog.Errorf("Request error %v", err)
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		klog.Errorf("Response error %v", err)
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		klog.Errorf("Get response body fail %v", err)
		return err
	}
	if res.StatusCode != http.StatusNoContent {
		klog.Errorf("response not ok:%v, %v", res.StatusCode, string(body))
		return fmt.Errorf("response not ok %d", res.StatusCode)
	}
	return nil
}

func createBackend(url, token string, opts CreateBackendOpts) (*BackendList, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	reqUrl := url + "/" + opts.SLBId + "/listeners/" + opts.ListenerId + "/members"
	klog.Infof("createBackend requestUrl:%v,token:%v", reqUrl, token)
	serversByte, err := json.Marshal(&opts.Servers)
	klog.Infof("requestBody is : %v", string(serversByte))
	if nil != err {
		klog.Errorf("servers conver to bytes error %v", err)
		return nil, err
	}
	req, err := http.NewRequest("POST", reqUrl, bytes.NewReader(serversByte))
	if err != nil {
		klog.Errorf("Request error %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		klog.Errorf("Response error %v", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		klog.Errorf("Get response body fail %v", err)
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		klog.Errorf("response not ok:%v,%v", res.StatusCode, string(body))
		return nil, fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result BackendList
	err = json.Unmarshal(body, &result)
	if err != nil {
		klog.Errorf("Unmarshal body fail: %v", err)
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
	klog.Infof("describeBackendservers requestUrl:%v, token:%v", reqUrl, token)
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		klog.Errorf("Request error %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		klog.Errorf("Response error %v", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		klog.Errorf("Get response body fail %v", err)
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		klog.Errorf("response not ok:%v,%v", res.StatusCode, string(body))
		return nil, fmt.Errorf("response not ok %d", res.StatusCode)
	}
	var result []Backend
	err = json.Unmarshal(body, &result)
	if err != nil {
		klog.Errorf("Unmarshal body fail: %v", err)
		return nil, err
	}
	return result, nil

}

func removeBackendServers(slburl, token, slbId, listnerId string, backendIdList []string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	bks := strings.Join(backendIdList, "\",\"")
	reqUrl, _ := url.Parse(slburl + "/" + slbId + "/listeners/" + listnerId + "/members" + "?backendIdList=[\"" + bks + "\"]")
	reqUrl.RawQuery = reqUrl.Query().Encode()
	klog.Infof("removeBackendServers requestUrl:%v, token:%v", reqUrl, token)
	req, err := http.NewRequest("DELETE", reqUrl.String(), nil)
	if err != nil {
		klog.Errorf("Request error %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC1123))
	res, err := client.Do(req)
	if err != nil {
		klog.Errorf("Response error %v", err)
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		klog.Errorf("Get response body fail %v", err)
		return err
	}
	if res.StatusCode != http.StatusOK {
		klog.Errorf("response not ok:%v, %v", res.StatusCode, string(body))
		return fmt.Errorf("response not ok %d", res.StatusCode)
	}
	return nil
}
