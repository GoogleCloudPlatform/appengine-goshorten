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

// goshorten is a demo application for App Engine Service Accounts. It
// shows how to use `appengine.AccessToken` method in combination with
// the URL shortener API.
package goshorten

import (
	"fmt"
	"html/template"
	"net/http"

	"appengine"

	"code.google.com/p/goauth2/appengine/serviceaccount"
	"code.google.com/p/google-api-go-client/urlshortener/v1"
)

// appengineHandler wraps http.Handler to pass it a new `appengine.Context` and handle errors.
type appengineHandler func(c appengine.Context, w http.ResponseWriter, r *http.Request) error

func (h appengineHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if err := h(c, w, r); err != nil {
		c.Errorf("%v", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func init() {
	http.Handle("/", appengineHandler(handle))
	http.Handle("/shorten", appengineHandler(shorten))
}

var mainTemplate = template.Must(template.New("main").Parse(`<html>
<body>
  <h1>Go/App Engine Service account demo</h1>
  <form action="/shorten" method="POST">
    <label for="url">Enter URL:</url>
    <input type="text" name="url">
    <input type="submit" value="shorten!">
  </form>
  <h2>URLs recently shortened:</h2>
  <ul>
    {{range .Items}}
    <li>
      <a href="{{.Id}}" title="{{.Id}}">{{.LongUrl}}</a>
    </li>
    {{end}}
  </ul>
</body></html>`))

// handle renders the main page template with a submission form and the history of shortened urls.
func handle(c appengine.Context, w http.ResponseWriter, r *http.Request) error {
	client, err := serviceaccount.NewClient(c, "https://www.googleapis.com/auth/urlshortener")
	if err != nil {
		return fmt.Errorf("error creating authorized client: %v", err)
	}
	service, err := urlshortener.New(client)
	if err != nil {
		return fmt.Errorf("error creating urlshortener service: %v", err)
	}
	result, err := service.Url.List().Do()
	if err != nil {
		return fmt.Errorf("error getting history: %v", err)
	}
	c.Infof("urlshortener API response: %v", result)
	if err := mainTemplate.Execute(w, result); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}
	return nil
}

// shorten shortens a new url and redirects to the main page
func shorten(c appengine.Context, w http.ResponseWriter, r *http.Request) error {
	client, err := serviceaccount.NewClient(c, "https://www.googleapis.com/auth/urlshortener")
	if err != nil {
		return fmt.Errorf("error creating authorized client: %v", err)
	}
	service, err := urlshortener.New(client)
	if err != nil {
		return fmt.Errorf("error creating urlshortener service: %v", err)
	}
	result, err := service.Url.Insert(&urlshortener.Url{
		LongUrl: r.FormValue("url"),
	}).Do()
	if err != nil {
		return fmt.Errorf("error posting new url: %v", err)
	}
	c.Infof("urlshortener API response: %v", result)
	http.Redirect(w, r, "/", 303)
	return nil
}
