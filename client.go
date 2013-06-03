//
// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package goshorten

import (
	"net/http"
	"time"

	"appengine"
	"appengine/memcache"
	"appengine/urlfetch"
)

// accessToken returns an access token for the given scope and caches it.
func accessToken(c appengine.Context, scope string) (string, error) {
	if token, err := memcache.Get(c, "token"); err == nil {
		return string(token.Value), nil
	}
	tok, expiry, err := appengine.AccessToken(c, scope)
	if err != nil {
		return "", err
	}
	// Ignore memcache errors and return the access token.
	memcache.Set(c, &memcache.Item{
		Key:        "token",
		Value:      []byte(tok),
		Expiration: expiry.Sub(time.Now()),
	})
	return tok, nil
}

// authorizedClient returns an *http.Client authorized for the given scope.
func authorizedClient(c appengine.Context, scope string) (*http.Client, error) {
	tok, err := accessToken(c, scope)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: &authorizedTransport{
			transport: &urlfetch.Transport{
				Context:                       c,
				Deadline:                      0,
				AllowInvalidServerCertificate: false,
			},
			token: tok,
		},
	}, nil
}

// authorizedTransport is an implementation of http.RoundTripper adding an access token to every request.
type authorizedTransport struct {
	transport http.RoundTripper
	token     string
}

// RoundTrip issues an authorized HTTP request and returns its response.
func (t *authorizedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	newReq := *req
	newReq.Header.Set("Authorization", "OAuth "+t.token)
	return t.transport.RoundTrip(&newReq)
}
