package godbc

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strconv"
	"errors"
	"net/url"
	"mime/multipart"
	"encoding/json"
)

const (
	dbcAPIURL  = "http://api.dbcapi.me/api/"
	uploadURL  = "http://api.dbcapi.me/api/captcha"
	reportURL  = "http://api.dbcapi.me/api/captcha/%CAPTCHA_ID%/report"
	statusURL  = "http://api.dbcapi.me/api/captcha/%CAPTCHA_ID%"
	userURL    = "http://api.dbcapi.me/api/user"
	dbcstatURL = "http://api.dbcapi.me/api/status"
)

//CaptchaClient is used to send captcha to DBC servers
type CaptchaClient struct {
	Username string
	Password string
	SiteKey string
	SiteURL string
	Proxy string
	ProxyType string
	JSONString string
	Timeout uint8 //180 is the max
	PollRate int8 //5 seconds
	LastStatus CaptchaStatus
}

//Decode sends the captcha to dbc servers
func (c *CaptchaClient)Decode(timeout uint8) (int, error) {
	c.Timeout = timeout
	c.ParamToString()
	buffer := new(bytes.Buffer)
	writer := multipart.NewWriter(buffer)
	_ = writer.WriteField("type", "4")
	_ = writer.WriteField("password", c.Password)
	_ = writer.WriteField("token_params", c.JSONString)
	_ = writer.WriteField("username", c.Username)
	_ = writer.WriteField("swid", "0") //seems hardcoded to 0

	req, err := http.NewRequest(http.MethodPost, uploadURL, buffer)
	if err != nil {
		return 0, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "DBC/.NET v4.5") //not sure if using something else will trigger their systems
	req.Header.Add("Connection", "close")
	req.Header.Add("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
	req.Header.Add("Content-Length", string(req.ContentLength))
	req.Header.Add("Host", "api.dbcapi.me")
	cli := new(http.Client)
	resp, err := cli.Do(req)
	if err != nil {
		return 0, err
	}
	actualResponse := CaptchaStatusReal{}
	json.NewDecoder(resp.Body).Decode(&actualResponse)
	return actualResponse.CaptchaID, nil
}

//ParamToString turns variables into hardcoded json string. DBC hires cheap coders, gets cheap programs...
func (c *CaptchaClient)ParamToString() {
	if c.Proxy != "" {
		c.JSONString = `{ "proxy": "`+c.Proxy+`", "proxytype": "`+c.ProxyType+`", "googlekey": "`+c.SiteKey+`", "pageurl": "`+c.SiteURL+`" }`
	} else {
		c.JSONString = `{ "googlekey": "`+c.SiteKey+`", "pageurl": "`+c.SiteURL+`" }`
	}
}

//PollCaptcha keeps sending a get request to dbc for updated captcha results
func (c *CaptchaClient)PollCaptcha(id int) error {
	realID := strconv.Itoa(id)
	fullURL := uploadURL + "/" + realID
	//username and password are not required
	cli := new(http.Client)
	
	resp, err := cli.Get(fullURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bod, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	vals, err := url.ParseQuery(string(bod))
	if err != nil {
		return err
	}
	if len(vals) != 2 {
		cap := CaptchaStatus{}
		cap.CaptchaID = vals.Get("captcha")
		cap.Status = vals.Get("status")
		cap.IsCorrect = vals.Get("is_correct")
		cap.Text = vals.Get("text")
		c.LastStatus = cap
	} else {
		return errors.New("Dead captcha")
	}

	return nil
}

//GetText returns the text from the captcha
func (c *CaptchaClient)GetText() string {
	return c.LastStatus.Text
}

//UpdateURL is used for registration pages that may have account-based parameters in their fields
func (c *CaptchaClient)UpdateURL(url string) {
	c.SiteURL = url
}

//UpdateProxy updates the currently set proxy in the captcha client
func (c *CaptchaClient)UpdateProxy(proxy string) {
	c.Proxy = proxy
}

//UpdatePollRate updates the polling rate
func (c *CaptchaClient)UpdatePollRate(rate int8) error {
	if rate < 5 {
		return errors.New("using a pollrate lower than 5 can potentially get you banned from DBC. Pollrate not adjusted")
	}
	c.PollRate = rate
	return nil
}

//CaptchaStatus is the struct returned when polling a captchas status
type CaptchaStatus struct {
	Status string
	CaptchaID string
	IsCorrect string
	Text string
}

//CaptchaStatusReal is the actual json response when uploading captchas because their service is coded by 3rd worlders
type CaptchaStatusReal struct {
	Status int `json:"status"`
	CaptchaID int `json:"captcha"`
	IsCorrect bool `json:"is_correct"`
	Text string `json:"text"`
}

