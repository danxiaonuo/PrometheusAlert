package controllers

import (
	"PrometheusAlert/models"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// messageBody is ifly message
type iflymessage struct {
	tid       string `json:"tid"`       //模板id
	data      string `json:"data"`      //消息体
	telephone string `json:"telephone"` //告警接收短信号码
	time      string `json:"time"`      //告警发送时间
	Code      string `json:"Code"`      //告警消息
}

// Postiflymessage sends sms message by iflymessage
func Postiflymessage(Messages string, PhoneNumbers, logsign string) string {
	open := beego.AppConfig.String("open-iflydx")
	if open != "1" {
		logs.Info(logsign, "[iflydx]", "七陌短信接口未配置为开启状态，请先配置open-iflydx为1")
		return "陌短信接口未配置为开启状态，请先配置open-iflydx为1"
	}

	// generate url
	urlPath := beego.AppConfig.String("IFLY_URL")
	iflytid := beego.AppConfig.String("IFLY_TID")
	iflytype := beego.AppConfig.String("IFLY_TYPE")

	//当前时间
	now := time.Now().Format("2006-01-02 15:04:05")

	// construct body
	u := iflymessage{
		telephone: PhoneNumbers,
		time:      now,
		Code:      Messages,
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(u)
	logs.Info(logsign, "[iflydx]", b)

	var tr *http.Transport
	if proxyUrl := beego.AppConfig.String("proxy"); proxyUrl != "" {
		proxy := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(proxyUrl)
		}
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           proxy,
		}
	} else {
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	client := &http.Client{Transport: tr}
	// need set headers
	xheader := `{"Code":"` + Messages + `","time":"` + now + `"}`
	// logs.Info(logsign, "[长这样]", xheader)
	payload := strings.NewReader(url.Values{"tid": {iflytid}, "telephone": {PhoneNumbers}, "data": {xheader}, "type": {iflytype}}.Encode())
	// logs.Info(logsign, "[进来了]", payload)
	req, err := http.NewRequest("POST", urlPath, payload)
	// set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	resp, err := client.Do(req)

	if err != nil {
		logs.Error(logsign, "[iflydx]", err.Error())
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error(logsign, "[iflydx]", err.Error())
	}
	models.AlertToCounter.WithLabelValues("iflydx").Add(1)
	ChartsJson.Iflydx += 1
	logs.Info(logsign, "[iflydx]", string(result))
	return string(result)
}
