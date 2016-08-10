package lxhttpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/layer-x/layerx-commons/lxerrors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

var DefaultRetries = 5

type client struct {
	c *http.Client
}

func newClient() *client {
	return &client{
		c: http.DefaultClient,
	}
}

var emptyBytes []byte

func GetWithUnmarshal(url string, path string, headers map[string]string, jsonObject interface{}) (*http.Response, []byte, error) {
	resp, body, err := Get(url, path, headers)
	if err != nil {
		return resp, body, err
	}
	err = json.Unmarshal(body, jsonObject)
	if err != nil {
		err = lxerrors.New("could not unmarshal body into jsonObject", err)
	}
	return resp, body, err
}

func PostWithUnmarshal(url string, path string, headers map[string]string, message, jsonObject interface{}) (*http.Response, []byte, error) {
	resp, body, err := Post(url, path, headers, message)
	if err != nil {
		return resp, body, err
	}
	err = json.Unmarshal(body, jsonObject)
	if err != nil {
		err = lxerrors.New("could not unmarshal body into jsonObject", err)
	}
	return resp, body, err
}

func Get(url string, path string, headers map[string]string) (*http.Response, []byte, error) {
	return getWithRetries(url, path, headers, DefaultRetries)
}

func getWithRetries(url string, path string, headers map[string]string, retries int) (*http.Response, []byte, error) {
	resp, respBytes, err := func() (*http.Response, []byte, error) {
		completeURL := parseURL(url, path)
		request, err := http.NewRequest("GET", completeURL, nil)
		if err != nil {
			return nil, emptyBytes, lxerrors.New("error generating get request", err)
		}
		for key, value := range headers {
			request.Header.Add(key, value)
		}
		resp, err := newClient().c.Do(request)
		if err != nil {
			return resp, emptyBytes, lxerrors.New("error performing get request", err)
		}
		respBytes, err := ioutil.ReadAll(resp.Body)
		if resp.Body != nil {
			defer resp.Body.Close()
		}
		if err != nil {
			return resp, emptyBytes, lxerrors.New("error reading get response", err)
		}

		return resp, respBytes, nil
	}()
	if err != nil && retries > 0 {
		return getWithRetries(url, path, headers, retries-1)
	}
	return resp, respBytes, err
}

func GetAsync(url string, path string, headers map[string]string) (*http.Response, error) {
	return getAsyncWithRetries(url, path, headers, DefaultRetries)
}

func getAsyncWithRetries(url string, path string, headers map[string]string, retries int) (*http.Response, error) {
	resp, err := func() (*http.Response, error) {
		completeURL := parseURL(url, path)
		request, err := http.NewRequest("GET", completeURL, nil)
		if err != nil {
			return nil, lxerrors.New("error generating get request", err)
		}
		for key, value := range headers {
			request.Header.Add(key, value)
		}
		resp, err := newClient().c.Do(request)
		if err != nil {
			return resp, lxerrors.New("error performing get request", err)
		}
		return resp, nil
	}()
	if err != nil && retries > 0 {
		return getAsyncWithRetries(url, path, headers, retries-1)
	}
	return resp, err
}

func Delete(url string, path string, headers map[string]string) (*http.Response, []byte, error) {
	return deleteWithRetries(url, path, headers, DefaultRetries)
}

func deleteWithRetries(url string, path string, headers map[string]string, retries int) (*http.Response, []byte, error) {
	resp, respBytes, err := func() (*http.Response, []byte, error) {
		completeURL := parseURL(url, path)
		request, err := http.NewRequest("DELETE", completeURL, nil)
		if err != nil {
			return nil, emptyBytes, lxerrors.New("error generating delete request", err)
		}
		for key, value := range headers {
			request.Header.Add(key, value)
		}
		resp, err := newClient().c.Do(request)
		if err != nil {
			return resp, emptyBytes, lxerrors.New("error performing delete request", err)
		}
		respBytes, err := ioutil.ReadAll(resp.Body)
		if resp.Body != nil {
			defer resp.Body.Close()
		}
		if err != nil {
			return resp, emptyBytes, lxerrors.New("error reading delete response", err)
		}

		return resp, respBytes, nil
	}()
	if err != nil && retries > 0 {
		return deleteWithRetries(url, path, headers, retries-1)
	}
	return resp, respBytes, err
}

func DeleteAsync(url string, path string, headers map[string]string) (*http.Response, error) {
	return deleteAsyncWithRetries(url, path, headers, DefaultRetries)
}

func deleteAsyncWithRetries(url string, path string, headers map[string]string, retries int) (*http.Response, error) {
	resp, err := func() (*http.Response, error) {
		completeURL := parseURL(url, path)
		request, err := http.NewRequest("DELETE", completeURL, nil)
		if err != nil {
			return nil, lxerrors.New("error generating delete request", err)
		}
		for key, value := range headers {
			request.Header.Add(key, value)
		}
		resp, err := newClient().c.Do(request)
		if err != nil {
			return resp, lxerrors.New("error performing delete request", err)
		}

		return resp, nil
	}()
	if err != nil && retries > 0 {
		return deleteAsyncWithRetries(url, path, headers, retries-1)
	}
	return resp, err
}

func Post(url string, path string, headers map[string]string, message interface{}) (*http.Response, []byte, error) {
	return postWithRetries(url, path, headers, message, DefaultRetries)
}

func postWithRetries(url string, path string, headers map[string]string, message interface{}, retries int) (*http.Response, []byte, error) {
	resp, respBytes, err := func() (*http.Response, []byte, error) {
		switch message.(type) {
		case proto.Message:
			return postPB(url, path, headers, message.(proto.Message))
		case *bytes.Buffer:
			return postBuffer(url, path, headers, message.(*bytes.Buffer))
		default:
			_, err := json.Marshal(message)
			if err != nil {
				return nil, emptyBytes, lxerrors.New("message was not of expected type `json` or `protobuf`", err)
			}
			return postJson(url, path, headers, message)
		}
	}()
	if err != nil && retries > 0 {
		return postWithRetries(url, path, headers, message, retries-1)
	}
	return resp, respBytes, err
}

func postPB(url string, path string, headers map[string]string, pb proto.Message) (*http.Response, []byte, error) {
	data, err := proto.Marshal(pb)
	if err != nil {
		return nil, emptyBytes, lxerrors.New("could not proto.Marshal mesasge", err)
	}
	return postData(url, path, headers, data)
}

func postBuffer(url string, path string, headers map[string]string, buffer *bytes.Buffer) (*http.Response, []byte, error) {
	completeURL := parseURL(url, path)
	request, err := http.NewRequest("POST", completeURL, buffer)
	if err != nil {
		return nil, emptyBytes, lxerrors.New("error generating post request", err)
	}
	for key, value := range headers {
		request.Header.Add(key, value)
	}
	resp, err := newClient().c.Do(request)
	if err != nil {
		return resp, emptyBytes, lxerrors.New("error performing post request", err)
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return resp, emptyBytes, lxerrors.New("error reading post response", err)
	}

	return resp, respBytes, nil
}

func postJson(url string, path string, headers map[string]string, jsonStruct interface{}) (*http.Response, []byte, error) {
	//err has already been caught
	data, _ := json.Marshal(jsonStruct)
	return postData(url, path, headers, data)
}

func postData(url string, path string, headers map[string]string, data []byte) (*http.Response, []byte, error) {
	completeURL := parseURL(url, path)
	request, err := http.NewRequest("POST", completeURL, bytes.NewReader(data))
	if err != nil {
		return nil, emptyBytes, lxerrors.New("error generating post request", err)
	}
	for key, value := range headers {
		request.Header.Add(key, value)
	}
	resp, err := newClient().c.Do(request)
	if err != nil {
		return resp, emptyBytes, lxerrors.New("error performing post request", err)
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return resp, emptyBytes, lxerrors.New("error reading post response", err)
	}

	return resp, respBytes, nil
}

func PostFile(url, path, fileKey, pathToFile string) (*http.Response, []byte, error) {
	completeURL := parseURL(url, path)
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile(fileKey, pathToFile)
	if err != nil {
		return nil, emptyBytes, lxerrors.New("error writing to buffer", err)
	}

	// open file handle
	fh, err := os.Open(pathToFile)
	if err != nil {
		return nil, emptyBytes, lxerrors.New("error opening file", err)
	}

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return nil, emptyBytes, lxerrors.New("error copying file to form", err)
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(completeURL, contentType, bodyBuf)
	if err != nil {
		return resp, emptyBytes, lxerrors.New("error performing post", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, body, lxerrors.New("reading response body", err)
	}

	return resp, body, nil
}

func PostAsync(url string, path string, headers map[string]string, message interface{}) (*http.Response, error) {
	return postAsyncWithRetries(url, path, headers, message, DefaultRetries)
}

func postAsyncWithRetries(url string, path string, headers map[string]string, message interface{}, retries int) (*http.Response, error) {
	resp, err := func() (*http.Response, error) {
		switch message.(type) {
		case proto.Message:
			return postAsyncPB(url, path, headers, message.(proto.Message))
		case *bytes.Buffer:
			return postAsyncBuffer(url, path, headers, message.(*bytes.Buffer))
		default:
			_, err := json.Marshal(message)
			if err != nil {
				return nil, lxerrors.New("message was not of expected type `json` or `protobuf`", err)
			}
			return postAsyncJson(url, path, headers, message)
		}
	}()
	if err != nil && retries > 0 {
		return postAsyncWithRetries(url, path, headers, message, retries-1)
	}
	return resp, err
}

func postAsyncPB(url string, path string, headers map[string]string, pb proto.Message) (*http.Response, error) {
	data, err := proto.Marshal(pb)
	if err != nil {
		return nil, lxerrors.New("could not proto.Marshal mesasge", err)
	}
	return postAsyncData(url, path, headers, data)
}

func postAsyncBuffer(url string, path string, headers map[string]string, buffer *bytes.Buffer) (*http.Response, error) {
	completeURL := parseURL(url, path)
	request, err := http.NewRequest("POST", completeURL, buffer)
	if err != nil {
		return nil, lxerrors.New("error generating post request", err)
	}
	for key, value := range headers {
		request.Header.Add(key, value)
	}
	resp, err := newClient().c.Do(request)
	if err != nil {
		return resp, lxerrors.New("error performing post request", err)
	}

	return resp, nil
}

func postAsyncJson(url string, path string, headers map[string]string, jsonStruct interface{}) (*http.Response, error) {
	//err has already been caught
	data, _ := json.Marshal(jsonStruct)
	return postAsyncData(url, path, headers, data)
}

func postAsyncData(url string, path string, headers map[string]string, data []byte) (*http.Response, error) {
	completeURL := parseURL(url, path)
	request, err := http.NewRequest("POST", completeURL, bytes.NewReader(data))
	if err != nil {
		return nil, lxerrors.New("error generating post request", err)
	}
	for key, value := range headers {
		request.Header.Add(key, value)
	}
	resp, err := newClient().c.Do(request)
	if err != nil {
		return resp, lxerrors.New("error performing post request", err)
	}

	return resp, nil
}

func PostAsyncFile(url, path, fileKey, pathToFile string) (*http.Response, error) {
	completeURL := parseURL(url, path)
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile(fileKey, pathToFile)
	if err != nil {
		return nil, lxerrors.New("error writing to buffer", err)
	}

	// open file handle
	fh, err := os.Open(pathToFile)
	if err != nil {
		return nil, lxerrors.New("error opening file", err)
	}

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return nil, lxerrors.New("error copying file to form", err)
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(completeURL, contentType, bodyBuf)
	if err != nil {
		return resp, lxerrors.New("error performing post", err)
	}

	return resp, nil
}

func parseURL(url string, path string) string {
	if !strings.HasPrefix(url, "http://") || !strings.HasPrefix(url, "https://") {
		url = fmt.Sprintf("http://%s", url)
	}
	if strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}
	if strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}
	return fmt.Sprintf("%s/%s", url, path)
}
