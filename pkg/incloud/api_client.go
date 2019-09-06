package incloud

import (
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"gitserver/kubernetes/inspur-cloud-controller-manager/pkg/loadbalance"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func GetKeyCloakToken(requestedSubject, tokenClientId, clientSecret, keycloakUrl string) (string, error) {
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

//http://cn-north-3.10.110.25.123.xip.io/slb/v1/slbs?slbName=123
//按slb id查询用户的slb
func DescribeLoadBalancers(url, token, slbId string) (*loadbalance.LoadBalancerSpec, error) {
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
	var result []loadbalance.LoadBalancerSpec
	err = xml.Unmarshal(body, &result)
	if err != nil {
		glog.Errorf("Unmarshal body fail: %v", err)
		return nil, err
	}
	return &result[0], nil
}

func DescribeListeners(url, token, slbId string) (*[]loadbalance.LisenerSpec, error) {
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
	var result []loadbalance.LisenerSpec
	err = xml.Unmarshal(body, &result)
	if err != nil {
		glog.Errorf("Unmarshal body fail: %v", err)
		return nil, err
	}
	return &result, nil
}
