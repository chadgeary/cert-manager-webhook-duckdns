package duckdns

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
)

const (
	defaultBaseURL = "https://www.duckdns.org"
	domainStub     = "/update?domains="
	tokenStub      = "&token="
	ip4Stub        = "&ip="
	ip6Stub        = "&ipv6="
	txtStub        = "&txt="
	verboseStub    = "&verbose="
	clearStub      = "&clear="

	defaultUserAgent = "duckdns-go/1.0.3"
)

// Response structure containing the http response and the data from the body
type Response struct {
	HTTPResponse *http.Response
	Data         string
}

// Config structure containing the client configuration
type ConfigC struct {
	DomainNames []string
	Token       string
	IPv4        string
	IPv6        string
	Verbose     bool
}

// Valid function to check if the client configuration is valid
func (c *ConfigC) Valid() bool {
	if c.Token != "" && len(c.DomainNames) > 0 {
		return true
	}
	return false
}

// Client structure
type ClientC struct {
	httpClient *http.Client
	BaseURL    string
	UserAgent  string

	Config *ConfigC
}

// NewClient function to return a valid duckdns client
func NewClient(httpClient *http.Client, config *ConfigC) *ClientC {
	if !config.Valid() {
		klog.Fatal("Configuration is not valid")
	}

	c := &ClientC{httpClient: httpClient,
		BaseURL:   defaultBaseURL,
		UserAgent: defaultUserAgent,
		Config:    config}
	return c
}

// SetUserAgent function to set a custom header for the UserAgent
func (c *ClientC) SetUserAgent(ua string) {
	c.UserAgent = ua
}

// SetVerbose function to set the response of the client request to verbose=true
func (c *ConfigC) SetVerbose(verbose bool) {
	c.Verbose = verbose
}

func (c *ClientC) makeGetRequest(ctx context.Context, path, pathObf string, response *Response) (*http.Response, error) {

	req, err := c.newRequest(http.MethodGet, path, pathObf)
	if err != nil {
		return nil, err
	}

	resp, err := c.request(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *ClientC) newRequest(method, path, pathObf string) (*http.Request, error) {
	url := c.BaseURL + path
	urlObf := c.BaseURL + pathObf

	klog.Infof("Sending request to %v", urlObf)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header = make(http.Header)
	req.Header.Add("User-Agent", c.UserAgent)

	return req, err
}

func (c *ClientC) request(ctx context.Context, req *http.Request, response *Response) (*http.Response, error) {
	if ctx == nil {
		return nil, errors.New("context must be non-nil")
	}
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if response != nil {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return resp, err
		}
		response.Data = string(bytes)
	}

	return resp, err
}

// UpdateIP function to update IPv4 and/or without IP address
func (c *ClientC) UpdateIP(ctx context.Context) (*Response, error) {
	subdomains := strings.Join(c.Config.DomainNames, ",")
	url := fmt.Sprintf("%s%s%s%s%s", domainStub, subdomains, tokenStub, c.Config.Token, ip4Stub)
	urlObf := fmt.Sprintf("%s%s%s%s%s", domainStub, subdomains, tokenStub, "*********", ip4Stub)

	if c.Config.Verbose {
		url = fmt.Sprintf("%s%s%s", url, verboseStub, strconv.FormatBool(c.Config.Verbose))
		urlObf = fmt.Sprintf("%s%s%s", urlObf, verboseStub, strconv.FormatBool(c.Config.Verbose))
	}

	response := &Response{}
	resp, err := c.makeGetRequest(ctx, url, urlObf, response)

	if err != nil {
		return response, err
	}

	response.HTTPResponse = resp
	return response, err
}

// UpdateIPWithValues to update IPv4 and/or with IP address
func (c *ClientC) UpdateIPWithValues(ctx context.Context, ipv4, ipv6 string) (*Response, error) {
	subdomains := strings.Join(c.Config.DomainNames, ",")
	url := fmt.Sprintf("%s%s%s%s%s", domainStub, subdomains, tokenStub, c.Config.Token, ip4Stub)
	urlObf := fmt.Sprintf("%s%s%s%s%s", domainStub, subdomains, tokenStub, "*********", ip4Stub)

	if ipv6 == "" {
		url = fmt.Sprintf("%s%s", url, ipv4)
		urlObf = fmt.Sprintf("%s%s", urlObf, ipv4)
	} else {
		url = fmt.Sprintf("%s%s%s%s", url, ipv4, ip6Stub, ipv6)
		urlObf = fmt.Sprintf("%s%s%s%s", urlObf, ipv4, ip6Stub, ipv6)
	}

	if c.Config.Verbose {
		url = fmt.Sprintf("%s%s%s", url, verboseStub, strconv.FormatBool(c.Config.Verbose))
		urlObf = fmt.Sprintf("%s%s%s", urlObf, verboseStub, strconv.FormatBool(c.Config.Verbose))
	}

	resp := &Response{}
	_, err := c.makeGetRequest(ctx, url, urlObf, resp)

	return resp, err
}

// ClearIP function that clears the IP from duckdns system
func (c *ClientC) ClearIP(ctx context.Context) (*Response, error) {
	subdomains := strings.Join(c.Config.DomainNames, ",")
	url := fmt.Sprintf("%s%s%s%s%s%s", domainStub, subdomains, tokenStub, c.Config.Token, clearStub, "true")
	urlObf := fmt.Sprintf("%s%s%s%s%s%s", domainStub, subdomains, tokenStub, "*********", clearStub, "true")

	if c.Config.Verbose {
		url = fmt.Sprintf("%s%s%s", url, verboseStub, strconv.FormatBool(c.Config.Verbose))
		urlObf = fmt.Sprintf("%s%s%s", urlObf, verboseStub, strconv.FormatBool(c.Config.Verbose))
	}

	resp := &Response{}
	_, err := c.makeGetRequest(ctx, url, urlObf, resp)

	return resp, err
}

// UpdateRecord function to update TXT record
func (c *ClientC) UpdateRecord(ctx context.Context, record string) (*Response, error) {
	subdomains := strings.Join(c.Config.DomainNames, ",")
	url := fmt.Sprintf("%s%s%s%s%s%s", domainStub, subdomains, tokenStub, c.Config.Token, txtStub, record)
	urlObf := fmt.Sprintf("%s%s%s%s%s%s", domainStub, subdomains, tokenStub, "*********", txtStub, record)

	if c.Config.Verbose {
		url = fmt.Sprintf("%s%s%s", url, verboseStub, strconv.FormatBool(c.Config.Verbose))
		urlObf = fmt.Sprintf("%s%s%s", urlObf, verboseStub, strconv.FormatBool(c.Config.Verbose))
	}

	resp := &Response{}
	_, err := c.makeGetRequest(ctx, url, urlObf, resp)

	return resp, err
}

// ClearRecord function to clear TXT record
func (c *ClientC) ClearRecord(ctx context.Context, record string) (*Response, error) {
	subdomains := strings.Join(c.Config.DomainNames, ",")
	url := fmt.Sprintf("%s%s%s%s%s%s%s%s", domainStub, subdomains, tokenStub, c.Config.Token, txtStub, record, clearStub, "true")
	urlObf := fmt.Sprintf("%s%s%s%s%s%s%s%s", domainStub, subdomains, tokenStub, "*********", txtStub, record, clearStub, "true")

	if c.Config.Verbose {
		url = fmt.Sprintf("%s%s%s", url, verboseStub, strconv.FormatBool(c.Config.Verbose))
		urlObf = fmt.Sprintf("%s%s%s", urlObf, verboseStub, strconv.FormatBool(c.Config.Verbose))
	}

	resp := &Response{}
	_, err := c.makeGetRequest(ctx, url, urlObf, resp)

	return resp, err
}

// GetRecord function to get TXT record like dig+ <domain> TXT
func (c *ClientC) GetRecord() (string, error) {
	var subdomains string
	if strings.Contains(c.Config.DomainNames[0], "duckdns.org") {
		subdomains = c.Config.DomainNames[0]
	} else {
		subdomains = c.Config.DomainNames[0] + ".duckdns.org"
	}
	txt, err := net.LookupTXT(subdomains)
	if err != nil {
		return "", fmt.Errorf("unable to get txt record, %v", err)
	}

	if len(txt) == 0 {
		return "", nil
	}

	//duckdns should have only 1 record
	return txt[0], nil
}
