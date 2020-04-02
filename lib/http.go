package lib

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

//type Request struct {
//	Url         UrlType
//	Method      string
//	Headers     map[string]string
//	ContentType string
//	Body        []byte
//}
//
//type Response struct {
//	Url         UrlType
//	Status      int
//	Body        []byte
//	Headers     map[string]string
//	ContentType string
//}
//
//type UrlType struct {
//	Scheme   string
//	Domain   string
//	Host     string
//	Port     string
//	Path     string
//	Query    string
//	Fragment string
//}

var (
	client           *http.Client
	clientNoRedirect *http.Client
)

func InitHttpClient(ThreadsNum int, DownProxy string, Timeout time.Duration) error {
	tr := &http.Transport{
		MaxIdleConns:        0,
		MaxIdleConnsPerHost: ThreadsNum * 2,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
	if DownProxy != "" {
		u, err := url.Parse(DownProxy)
		if err != nil {
			return err
		}
		tr.Proxy = http.ProxyURL(u)
	}
	client = &http.Client{
		Transport: tr,
		Timeout:   Timeout * time.Second,
	}
	clientNoRedirect = &http.Client{
		Transport: tr,
		Timeout:   Timeout * time.Second,
	}
	clientNoRedirect.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return nil
}

func DoRequest(req *http.Request, redirect bool) (*Response, error) {
	if req.Body == nil || req.Body == http.NoBody {
	} else {
		req.Header.Set("Content-Length", strconv.Itoa(int(req.ContentLength)))
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	var oResp *http.Response
	var err error
	if redirect {
		oResp, err = client.Do(req)
	} else {
		oResp, err = clientNoRedirect.Do(req)
	}
	if err != nil {
		return nil, err
	}
	resp, err := ParseResponse(oResp)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func parseUrl(u *url.URL) *UrlType {
	nu := &UrlType{}
	nu.Scheme = u.Scheme
	nu.Domain = u.Hostname()
	nu.Host = u.Host
	nu.Port = u.Port()
	nu.Path = u.EscapedPath()
	nu.Query = u.RawQuery
	nu.Fragment = u.Fragment
	return nu
}

func ParseRequest(oReq *http.Request) (*Request, error) {
	req := &Request{}
	req.Method = oReq.Method
	req.Url = parseUrl(oReq.URL)
	header := make(map[string]string)
	for k := range oReq.Header {
		header[k] = oReq.Header.Get(k)
	}
	req.Headers = header
	req.ContentType = oReq.Header.Get("Content-Type")
	if oReq.Body == nil || oReq.Body == http.NoBody {
	} else {
		data, err := ioutil.ReadAll(oReq.Body)
		if err != nil {
			return nil, err
		}
		req.Body = data
		oReq.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	}
	return req, nil
}

func ParseResponse(oResp *http.Response) (*Response, error) {
	var resp Response
	header := make(map[string]string)
	resp.Status = int32(oResp.StatusCode)
	resp.Url = parseUrl(oResp.Request.URL)
	for k := range oResp.Header {
		header[k] = oResp.Header.Get(k)
	}
	resp.Headers = header
	resp.ContentType = oResp.Header.Get("Content-Type")
	body, err := getRespBody(oResp)
	if err != nil {
		return nil, err
	}
	resp.Body = body
	return &resp, nil
}

func getRespBody(oResp *http.Response) ([]byte, error) {
	var body []byte
	if oResp.Header.Get("Content-Encoding") == "gzip" {
		gr, _ := gzip.NewReader(oResp.Body)
		defer gr.Close()
		for {
			buf := make([]byte, 1024)
			n, err := gr.Read(buf)
			if err != nil && err != io.EOF {
				//utils.Logger.Error(err)
				return nil, err
			}
			if n == 0 {
				break
			}
			body = append(body, buf...)
		}
	} else {
		raw, err := ioutil.ReadAll(oResp.Body)
		if err != nil {
			//utils.Logger.Error(err)
			return nil, err
		}
		defer oResp.Body.Close()
		body = raw
	}
	return body, nil
}
