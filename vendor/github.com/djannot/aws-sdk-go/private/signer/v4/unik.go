// Intercept s3 requests for the UnikHub
package v4

import (
	"bytes"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

type RequestToValidate struct {
	Pass   string      `json:"pass"`
	Method string      `json:"method"`
	Path   string      `json:"path"`
	Query  url.Values  `json:"query"`
	Header http.Header `json:"headers"`
}

type ValidationResponse struct {
	Message     string `json:"message"`
	AccessKeyID string `json:"access_key_id"`
	Region      string `json:"region"`
	Bucket      string `json:"bucket"`
}

type RequestToSign struct {
	RequestToValidate  RequestToValidate `json:"request_to_validate"`
	FormattedShortTime string            `json:"formatted_short_time"`
	ServiceName        string            `json:"service_name"`
	StringToSign       string            `json:"string_to_sign"`
}

type SignatureResponse struct {
	Signature   []byte `json:"signature"`
	Err   string `json:"err"`
}

// Validate the request with the UnikHub
func (v4 *signer) validateRequest(s3AuthProxyUrl string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}
	// Send the API request to the UnikHub
	authReq, err := http.NewRequest("GET", s3AuthProxyUrl+"/aws_info", nil)
	if err != nil {
		return err
	}
	authReq.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(authReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var validationResponse ValidationResponse
	err = decoder.Decode(&validationResponse)
	if err != nil {
		return err
	}
	// If the response code is 200, then the request is validated by the UnikHub
	if resp.StatusCode == 200 {
		// The s3 region and bucket aren't known by the UnikHubClient. They are provided by the UnikHub
		v4.CredValues.AccessKeyID = validationResponse.AccessKeyID
		v4.Region = validationResponse.Region
		newURL := strings.Replace(v4.Request.URL.String(), "AWSREGION", validationResponse.Region, 1)
		v4.Request.URL, err = url.Parse(newURL)
		if err != nil {
			err = errors.New("Can't replace the Aws Region in the request")
			return err
		}
		newURL = strings.Replace(v4.Request.URL.String(), "AWSBUCKET", validationResponse.Bucket, 1)
		v4.Request.URL, err = url.Parse(newURL)
		if err != nil {
			err = errors.New("Can't replace the Aws Bucket in the request")
			return err
		}
		v4.pass = v4.Request.Header.Get("X-Amz-Meta-Unik-Password")

		// Remove the X-Amz-Meta-Unik-Password and X-Amz-Meta-Unik-Email headers because they shouldn't be stored with the /bucket/user/image/version object
		v4.Request.Header.Del("X-Amz-Meta-Unik-Password")
	} else {
		err = errors.New(validationResponse.Message)
		return err
	}
	return nil
}

// Get a signature from the UnikHub
func (v4 *signer) getSignature(s3AuthProxyUrl string) error {
	// Get the URL and parse it (to get the Path)
	u, err := url.Parse(v4.Request.URL.String())
	if err != nil {
		return err
	}
	pass := v4.pass

	// Prepare the data to send to the UnikHub
	requestToValidate := RequestToValidate{
		Pass:   pass,
		Method: v4.Request.Method,
		Path:   u.Path,
		Query:  u.Query(),
		Header: v4.Request.Header,
	}

	// Prepare the data to send to the UnikHub
	requestToSign := RequestToSign{
		RequestToValidate:  requestToValidate,
		FormattedShortTime: v4.formattedShortTime,
		ServiceName:        v4.ServiceName,
		StringToSign:       v4.stringToSign,
	}
	j, err := json.Marshal(requestToSign)
	if err != nil {
		return err
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}
	// Send the API request to the UnikHub
	authReq, err := http.NewRequest("POST", s3AuthProxyUrl+"/sign", bytes.NewBuffer(j))
	if err != nil {
		return err
	}
	authReq.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(authReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var signatureResponse SignatureResponse
	err = decoder.Decode(&signatureResponse)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		err = errors.New(signatureResponse.Err)
		return err
	}
	//v4.CredValues.AccessKeyID = awsCredentials.AccessKeyID
	v4.signature = hex.EncodeToString(signatureResponse.Signature)
	return nil
}
