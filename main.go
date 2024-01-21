package main

import (
	"html/template"
	"net/http"
	"strings"
)

var (
	templates = template.Must(template.New("").
			Funcs(funcMap).
			ParseFiles("templates/index.html", "templates/new_card.html"))

	funcMap = template.FuncMap{
		"split": strings.Split,
	}
	authApi     = "/github/authorize"
	callbackApi = "/github/callback"
	searchApi   = "/github/search"
	baseUrl     = "http://localhost:8080"
)

func main() {
	// GET / : return the index.html page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "index.html", nil)
	})

	// GET /github/authorize : used to allow the user to give access to their repos by logging in using OAuth
	http.HandleFunc(authApi, githubAuthorizeHandler)

	// GET /github/callback : after a successful authorization a request is sent from github (with a code) to out callback url which we defined when creating the OAuth App
	http.HandleFunc(callbackApi, githubCallbackHandler)

	// GET /github/commits : return all the funny commits for the authenticated user
	http.HandleFunc(searchApi, githubSearchHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// start the server
	http.ListenAndServe(":8080", nil)

}
