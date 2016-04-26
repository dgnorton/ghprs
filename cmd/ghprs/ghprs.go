package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func usage() {
	println(`usage: ghprs -repo "<owner>/<repo>" [-token <access token>] [-o <outfile>]`)
	println("")
	println("Example:")
	println(`   # Using GitHub API without authentication (much lower rate limit):`)
	println(`   ghprs -repo "influxdata/influxdb"`)
	println("")
	println(`   # Using GitHub API with access token:`)
	println(`    ghprs -repo "influxdata/influxdb" -token abc123def456abc789`)
}

func main() {
	// Parse command line options.
	accessToken := flag.String("token", "", "github access token")
	ownerAndRepo := flag.String("repo", "", "github repo (e.g., influxdata/influxdb)")
	outfile := flag.String("o", "ghprs.json", "output file")
	flag.Parse()

	// Parse owner and repo from command line.
	if *ownerAndRepo == "" {
		usage()
		os.Exit(1)
	}

	strs := strings.Split(*ownerAndRepo, "/")
	if len(strs) != 2 {
		usage()
		os.Exit(1)
	}

	owner := strs[0]
	repo := strs[1]

	// Create oauth2 client, if access token was provided on the command line.
	var tc *http.Client
	if *accessToken != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: *accessToken})
		tc = oauth2.NewClient(oauth2.NoContext, ts)
	}

	// Create GitHub API client.
	client := github.NewClient(tc)

	// Set PR list options.
	opt := &github.PullRequestListOptions{
		State: "closed",
		Base:  "master",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	// Create an array to collect all PRs into.
	prs := []github.PullRequest{}

	// Loop to request pages of PRs.
	for {
		// Send request to GitHub.
		prspage, resp, err := client.PullRequests.List(owner, repo, opt)
		fatalIfErr(err)

		// Check if we've hit our rate limit.
		if resp.Rate.Remaining == 0 {
			fatalIfErr(fmt.Errorf("rate limit reached, try again after %s", resp.Rate.Reset.String()))
		}

		// Notify user of remaing requests before hitting rate limit.
		fmt.Printf("requests remaining: %d\n", resp.Rate.Remaining)
		fmt.Printf("current page: %d, last page: %d\n", opt.ListOptions.Page, resp.LastPage)

		// Append PRs from the result to the main PR list.
		prs = append(prs, prspage...)

		// Check if we've fetched the last page of PRs.
		if resp.LastPage == 0 {
			// Check if we need to switch to open PRs.
			if opt.State == "closed" {
				opt.State = "open"
				opt.ListOptions.Page = 1
				continue
			}

			// We've fetched all closed and open PRs so exit the loop.
			break
		}

		// Update the request options to fetch the next page.
		opt.ListOptions.Page = resp.NextPage
	}

	// Write PRs JSON to file.
	b, err := json.MarshalIndent(prs, "", "\t")
	fatalIfErr(err)

	f, err := os.Create(*outfile)
	fatalIfErr(err)
	defer f.Close()
	_, err = f.Write(b)
	fatalIfErr(err)

	fmt.Printf("PRs written to %s\n", *outfile)
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
