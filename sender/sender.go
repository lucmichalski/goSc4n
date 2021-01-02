package sender

import (
	"crypto/tls"
	"net/http"

	"errors"
	"fmt"
	"github.com/jaeles-project/jaeles/utils"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"io/ioutil"
	"math/rand"
	//"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jaeles-project/jaeles/libs"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

// JustSend just sending request
func JustSend(options libs.Options, req libs.Request) (res libs.Response, err error) {
	if req.Method == "" {
		req.Method = "GET"
	}
	method := req.Method
	url := req.URL
	body := req.Body
	headers := GetHeaders(req)
	proxy := options.Proxy

	// override proxy
	if req.Proxy != "" && req.Proxy != "blank" {
		proxy = req.Proxy
	}

	timeout := options.Timeout
	if req.Timeout > 0 {
		timeout = req.Timeout
	}

	//disableCompress := false
	//if len(headers) > 0 && strings.Contains(headers["Accept-Encoding"], "gzip") {
	//	disableCompress = true
	//}

	// update it again
	var newHeader []map[string]string
	for k, v := range headers {
		element := make(map[string]string)
		element[k] = v
		newHeader = append(newHeader, element)
	}
	req.Headers = newHeader

	// disable log when retry
	logger := logrus.New()
	if !options.Debug {
		logger.Out = ioutil.Discard
	}

	request := fasthttp.AcquireRequest()
	//client := resty.New()
	//client.SetLogger(logger)
	tlsCfg := &tls.Config{
		Renegotiation:            tls.RenegotiateOnceAsClient,
		PreferServerCipherSuites: true,
		InsecureSkipVerify:       true,
	}

	if proxy != "" {
		// some times burp reject default cipher
		tlsCfg = &tls.Config{
			CipherSuites: []uint16{
				tls.TLS_RSA_WITH_RC4_128_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			},
			Renegotiation:            tls.RenegotiateOnceAsClient,
			PreferServerCipherSuites: true,
			InsecureSkipVerify:       true,
		}
	}
	client := &fasthttp.Client{
		MaxIdleConnDuration: time.Duration(timeout) * time.Second,
		MaxConnsPerHost: 1000,
		ReadTimeout: time.Duration(timeout) * time.Second,
		WriteTimeout: time.Duration(timeout) * time.Second,
		TLSConfig: tlsCfg,
	}
	//client.SetTransport(&http.Transport{
	//	MaxIdleConns:          100,
	//	MaxConnsPerHost:       1000,
	//	IdleConnTimeout:       time.Duration(timeout) * time.Second,
	//	ExpectContinueTimeout: time.Duration(timeout) * time.Second,
	//	ResponseHeaderTimeout: time.Duration(timeout) * time.Second,
	//	TLSHandshakeTimeout:   time.Duration(timeout) * time.Second,
	//	DisableCompression:    disableCompress,
	//	DisableKeepAlives:     true,
	//	TLSClientConfig:       tlsCfg,
	//})

	if proxy != "" {
		client.Dial = fasthttpproxy.FasthttpSocksDialer(proxy)
	}

	for key,headerValue := range headers{
		request.Header.Add(key,headerValue)
	}
	//client.SetHeaders(headers)
	request.Header.Set("Connection","close")

	//no need anymore because that Client supports automatic retry on idempotent requests' failure. until ReadTimeout end
	//if options.Retry > 0 {
	//	client.SetRetryCount(options.Retry)
	//}
	//client.SetTimeout(time.Duration(timeout) * time.Second)
	//client.SetRetryWaitTime(time.Duration(timeout/2) * time.Second)
	//client.SetRetryMaxWaitTime(time.Duration(timeout) * time.Second)
	//timeStart := time.Now()
	// redirect policyfalse
	if req.Redirect == false {
		client.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
			// keep the header the same
			// client.SetHeaders(headers)

			res.StatusCode = req.Response.StatusCode
			res.Status = req.Response.Status
			resp := req.Response
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				utils.ErrorF("%v", err)
			}
			bodyString := string(bodyBytes)
			resLength := len(bodyString)
			// format the headers
			var resHeaders []map[string]string
			for k, v := range resp.Header {
				element := make(map[string]string)
				//fmt.Printf("%v: %v\n", k, v)
				element[k] = strings.Join(v[:], "")
				resLength += len(fmt.Sprintf("%s: %s\n", k, strings.Join(v[:], "")))
				resHeaders = append(resHeaders, element)
			}

			// response time in second
			resTime := time.Since(timeStart).Seconds()
			resHeaders = append(resHeaders,
				map[string]string{"Total Length": strconv.Itoa(resLength)},
				map[string]string{"Response Time": fmt.Sprintf("%f", resTime)},
			)

			// set some variable
			res.Headers = resHeaders
			res.StatusCode = resp.StatusCode
			res.Status = fmt.Sprintf("%v %v", resp.Status, resp.Proto)
			res.Body = bodyString
			res.ResponseTime = resTime
			res.Length = resLength
			// beautify
			res.Beautify = BeautifyResponse(res)
			return errors.New("auto redirect is disabled")
		}))

		client.AddRetryCondition(
			func(r *resty.Response, err error) bool {
				return false
			},
		)
	} else {
		client.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
			// keep the header the same
			//client.SetHeaders(headers)
			for key,headerValue := range headers{
				request.Header.Add(key,headerValue)
			}
			return nil
		}))
	}

	var requestTime time.Duration
	response := fasthttp.AcquireResponse()
	var resp *resty.Response
	// really sending things here
	method = strings.ToLower(strings.TrimSpace(method))
	switch method {
	case "get":
		request.SetBody([]byte(body))
		request.Header.SetMethod("GET")
		request.SetRequestURI(url)
		startTime := time.Now()
		client.Do(request,response)
		endTime := time.Now()
		requestTime = startTime.Sub(endTime)
		//resp, err = client.R().
		//	SetBody([]byte(body)).
		//	Get(url)
		break
	case "post":
		request.SetBody([]byte(body))
		request.Header.SetMethod("POST")
		request.SetRequestURI(url)
		startTime := time.Now()
		client.Do(request,response)
		endTime := time.Now()
		requestTime = startTime.Sub(endTime)
		//resp, err = client.R().EnableTrace().
		//	SetBody([]byte(body)).
		//	Post(url)
		break
	case "head":
		request.SetBody([]byte(body))
		request.Header.SetMethod("HEAD")
		request.SetRequestURI(url)
		startTime := time.Now()
		client.Do(request,response)
		endTime := time.Now()
		requestTime = startTime.Sub(endTime)
		//resp, err = client.R().
		//	SetBody([]byte(body)).
		//	Head(url)
		break
	case "options":
		request.SetBody([]byte(body))
		request.Header.SetMethod("OPTIONS")
		request.SetRequestURI(url)
		startTime := time.Now()
		client.Do(request,response)
		endTime := time.Now()
		requestTime = startTime.Sub(endTime)
		//resp, err = client.R().
		//	SetBody([]byte(body)).
		//	Options(url)
		break
	case "patch":
		request.SetBody([]byte(body))
		request.Header.SetMethod("PATCH")
		request.SetRequestURI(url)
		startTime := time.Now()
		client.Do(request,response)
		endTime := time.Now()
		requestTime = startTime.Sub(endTime)
		//resp, err = client.R().
		//	SetBody([]byte(body)).
		//	Patch(url)
		break
	case "put":
		request.SetBody([]byte(body))
		request.Header.SetMethod("PUT")
		request.SetRequestURI(url)
		startTime := time.Now()
		client.Do(request,response)
		endTime := time.Now()
		requestTime = startTime.Sub(endTime)
		//resp, err = client.R().
		//	SetBody([]byte(body)).
		//	Put(url)
		break
	case "delete":
		request.SetBody([]byte(body))
		request.Header.SetMethod("DELETE")
		request.SetRequestURI(url)
		startTime := time.Now()
		client.Do(request,response)
		endTime := time.Now()
		requestTime = startTime.Sub(endTime)
		//resp, err = client.R().
		//	SetBody([]byte(body)).
		//	Delete(url)
		break
	}

	// in case we want to get redirect stuff
	if res.StatusCode != 0 {
		return res, nil
	}

	if err != nil {
		utils.ErrorF("%v %v", url, err)
		if strings.Contains(err.Error(), "EOF") && resp.StatusCode() != 0 {
			return ParseResponse(*resp,response,requestTime), nil
		}
		return libs.Response{}, err
	}

	return ParseResponse(*resp,response,requestTime), nil
}

// ParseResponse field to Response
func ParseResponse(resp resty.Response,resp1 *fasthttp.Response,requestTime time.Duration) (res libs.Response) {
	// var res libs.Response
	resLength := len(string(resp1.Body()))
	// format the headers
	var resHeaders []map[string]string
	resp1.Header.VisitAll(func(key, value []byte) {
		element := make(map[string]string)
		stringsValue := strings.Split(string(value),"\n")
		element[string(key)] = strings.Join(stringsValue, "")
		resLength += len(fmt.Sprintf("%s: %s\n", string(key), strings.Join(stringsValue, "")))
		resHeaders = append(resHeaders, element)
	})
	//for k, v := range resp1.Header.String() {
	//	element := make(map[string]string)
	//	element[k] = strings.Join(v[:], "")
	//	resLength += len(fmt.Sprintf("%s: %s\n", k, strings.Join(v[:], "")))
	//	resHeaders = append(resHeaders, element)
	//}
	// response time in second
	resTime := float64(requestTime) / float64(time.Second)
	resHeaders = append(resHeaders,
		map[string]string{"Total Length": strconv.Itoa(resLength)},
		map[string]string{"Response Time": fmt.Sprintf("%f", resTime)},
	)

	// set some variable
	res.Headers = resHeaders
	res.StatusCode = resp.StatusCode()
	res.Status = fmt.Sprintf("%v %v", resp.Status(), resp.RawResponse.Proto)
	res.Body = string(resp.Body())
	res.ResponseTime = resTime
	res.Length = resLength
	// beautify
	res.Beautify = BeautifyResponse(res)
	return res
}

// GetHeaders generate headers if not provide
func GetHeaders(req libs.Request) map[string]string {
	// random user agent
	UserAgens := []string{
		"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3941.0 Safari/537.36",
		"Mozilla/5.0 (X11; U; Windows NT 6; en-US) AppleWebKit/534.12 (KHTML, like Gecko) Chrome/9.0.587.0 Safari/534.12",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36",
	}

	headers := make(map[string]string)
	if len(req.Headers) == 0 {
		rand.Seed(time.Now().Unix())
		headers["User-Agent"] = UserAgens[rand.Intn(len(UserAgens))]
		return headers
	}

	for _, header := range req.Headers {
		for key, value := range header {
			headers[key] = value
		}
	}

	rand.Seed(time.Now().Unix())
	// append user agent in case you didn't set user-agent
	if headers["User-Agent"] == "" {
		rand.Seed(time.Now().Unix())
		headers["User-Agent"] = UserAgens[rand.Intn(len(UserAgens))]
	}
	return headers
}

// BeautifyRequest beautify request
func BeautifyRequest(req libs.Request) string {
	var beautifyReq string
	// hardcoded HTTP/1.1 for now
	beautifyReq += fmt.Sprintf("%v %v HTTP/1.1\n", req.Method, req.URL)

	for _, header := range req.Headers {
		for key, value := range header {
			if key != "" && value != "" {
				beautifyReq += fmt.Sprintf("%v: %v\n", key, value)
			}
		}
	}
	if req.Body != "" {
		beautifyReq += fmt.Sprintf("\n%v\n", req.Body)
	}
	return beautifyReq
}

// BeautifyResponse beautify response
func BeautifyResponse(res libs.Response) string {
	var beautifyRes string
	beautifyRes += fmt.Sprintf("%v \n", res.Status)

	for _, header := range res.Headers {
		for key, value := range header {
			beautifyRes += fmt.Sprintf("%v: %v\n", key, value)
		}
	}

	beautifyRes += fmt.Sprintf("\n%v\n", res.Body)
	return beautifyRes
}
