package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

type Converter struct {
	filteredDomains []string
	roundTripper    http.RoundTripper
	debug           bool
}
type ErrorResponse struct {
	ErrorCode    int
	ErrorMessage string
}

func NewConverter(roundTripper http.RoundTripper, filteredDomains []string, debug bool) *Converter {
	return &Converter{
		roundTripper:    roundTripper,
		filteredDomains: filteredDomains,
		debug:           debug,
	}
}

func (t Converter) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.roundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		msg, _ := json.Marshal(ErrorResponse{
			ErrorCode:    resp.StatusCode,
			ErrorMessage: string(b),
		})
		resp.Body = ioutil.NopCloser(bytes.NewReader(msg))
		return resp, nil
	}
	var gRoutes GorouterRoutes
	json.Unmarshal(b, &gRoutes)

	routes := make([]string, 0)
	for route, _ := range gRoutes {
		if t.IsFiltered(route) {
			continue
		}
		routes = append(routes, route)
	}
	sort.StringSlice(routes).Sort()
	msg, statusCode := t.routesToResponse(routes)
	resp.StatusCode = statusCode
	resp.Status = fmt.Sprintf("%d", statusCode)
	resp.Body = ioutil.NopCloser(bytes.NewReader(msg))
	return resp, nil
}
func (t Converter) IsFiltered(route string) bool {
	if strings.HasPrefix(route, "*") {
		return true
	}
	for _, filteredDomain := range t.filteredDomains {
		if strings.HasSuffix(route, filteredDomain) {
			return true
		}
	}
	return false
}
func (t Converter) routesToResponse(routes []string) ([]byte, int) {
	statusCode := 200
	resp, err := json.Marshal(routes)
	if err != nil {
		statusCode = 500
		errMsg := "Error during converting routes"
		if t.debug {
			errMsg += ": " + err.Error()
		}
		resp, _ = json.Marshal(ErrorResponse{
			ErrorCode:    500,
			ErrorMessage: errMsg,
		})
	}
	return resp, statusCode
}
