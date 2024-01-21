// all the configuration is done in the main.go file.

// use go-github to authenticate with github through OAuth
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

type GithubAccessToken struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type GithubCommitsResponse struct {
	Items []struct {
		Sha    string `json:"sha"`
		Url    string `json:"html_url"`
		Commit struct {
			Committer struct {
				Date time.Time `json:"date"`
			} `json:"committer"`
			Message string `json:"message"`
		} `json:"commit"`
	} `json:"items"`
}

type FunnyCommit struct {
	Url     string
	Date    time.Time
	Message string
	Sha     string
	Repo    string
}

func getGithubAccessToken(code string) string {
	// use the code to get the access token
	// return the access token
	urlAccessToken := "https://github.com/login/oauth/access_token"
	clientID := getEnv("CLIENT_ID")
	clientSecret := getEnv("CLIENT_SECRET")

	reqMap := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
	}

	reqJSON, err := json.Marshal(reqMap)

	if err != nil {
		log.Fatalf("Error while marshaling the request map %s", err)
	}

	req, err := http.NewRequest("POST", urlAccessToken, bytes.NewBuffer(reqJSON))

	if err != nil {
		log.Fatalf("Error while creating the POST request %s", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatalf("Error while sending the POST request %s", err)
	}

	defer resp.Body.Close()
	var token GithubAccessToken
	json.NewDecoder(resp.Body).Decode(&token)

	return token.AccessToken
}

// authorize the user for certain scopes (repo:private, user)
func getAuthorizationUrl() string {
	authUrl := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&scope=%s&state=%s&redirect_uri=%s",
		getEnv("CLIENT_ID"),
		"repo,user",
		"random",
		baseUrl+callbackApi,
	)

	return authUrl
}

func buildSearchUrl(owner string, query string) string {
	baseUrl := "https://api.github.com/search/commits"

	queryParams := url.Values{}
	queryParams.Add("q", fmt.Sprintf("committer:%s %s", owner, query))
	queryParams.Add("type", "commits")

	return baseUrl + "?" + queryParams.Encode()
}

func searchCommits(token string, owner string) []FunnyCommit {
	keywords := []string{
		"damn",
		"shit",
		"fuck",
		"lol",
		"ass",
		"kill",
	}

	var commits []FunnyCommit
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, keyword := range keywords {
		wg.Add(1)

		go func(k string) {

			defer wg.Done()

			var commit GithubCommitsResponse
			searchUrl := buildSearchUrl(owner, k)

			req, err := http.NewRequest("GET", searchUrl, nil)

			if err != nil {
				log.Fatalf("Error while creating the GET request %s", err)
			}

			if token != "" {
				req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
			}

			// use go routines to fire all the requests
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Fatalf("Error while sending the GET request %s", err)
			}

			respBody, err := io.ReadAll(resp.Body)
			if err := json.Unmarshal(respBody, &commit); err != nil {
				log.Printf("Error while unmarshalling JSON response: %s", err)
				return
			}

			// check if the commit isn't empty
			if len(commit.Items) > 0 {
				for _, item := range commit.Items {
					funnyCommit := FunnyCommit{
						Url:     item.Url,
						Sha:     item.Sha,
						Message: item.Commit.Message,
						Date:    item.Commit.Committer.Date,
						Repo:    strings.Split(item.Url, "/")[4],
					}
					mu.Lock()
					commits = append(commits, funnyCommit)
					mu.Unlock()
				}
			}

		}(keyword)
	}

	wg.Wait()

	// sort the commits by date
	comparator := func(i, j int) bool {
		return commits[i].Date.Before(commits[j].Date)
	}

	// sort the commits by date
	sort.Slice(commits, comparator)

	return commits
}

func getAuthenticatedUser(token string) string {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)

	if err != nil {
		log.Fatalf("Error while creating the GET request %s", err)
	}

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatalf("Error while sending the GET request %s", err)
	}

	defer resp.Body.Close()

	var user map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&user)

	return user["login"].(string)
}
