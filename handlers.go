package main

import (
	"html/template"
	"log"
	"net/http"
)

var token string

// gin handlers
// GET /github/authorize : used to allow the user to give access to their repos by logging in using OAuth
func githubAuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	authUrl := getAuthorizationUrl()
	http.Redirect(w, r, authUrl, http.StatusTemporaryRedirect)
}

// GET /github/callback : after a successful authorization a request is sent from github (with a code) to out callback url which we defined when creating the OAuth App
func githubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	token = getGithubAccessToken(code)
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	// get the username of the authenticated user
	user := getAuthenticatedUser(token)

	// go to /github/commits to see the commits
	http.Redirect(w, r, searchApi+"?user="+user, http.StatusTemporaryRedirect)
}

// GET /github/search : return all the funny commits for the authenticated user
func githubSearchHandler(w http.ResponseWriter, r *http.Request) {
	owner := r.URL.Query().Get("user")
	commits := searchCommits(token, owner)

	// use the template to render the commits in the html page
	renderTemplate(w, "index.html", commits)
}

func renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	tmpl, err := template.New(templateName).Funcs(funcMap).ParseFiles("templates/index.html", "templates/new_card.html")
	if err != nil {
		log.Fatalf("Error parsing template: %s", err)
	}

	w.Header().Set("Content-Type", "text/html")

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Fatalf("Error executing template: %s", err)
	}
}
