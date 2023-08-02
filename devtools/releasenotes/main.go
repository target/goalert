package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/google/go-github/v51/github"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/oauth2"
)

//go:embed issuesummary.txt
var issuePrompt string

type Config struct {
	AI *openai.Client
	GH *github.Client
}

func main() {
	log.SetFlags(log.Lshortfile)
	openaiKey := flag.String("openai", os.Getenv("OPENAI_API_KEY"), "OpenAI API Key")
	githubToken := flag.String("github", os.Getenv("GH_TOKEN"), "GitHub Token")
	iss := flag.Int("issue", 0, "Summarize a GitHub issue.")
	pr := flag.Int("pr", 0, "Summarize a GitHub PR.")
	flag.Parse()

	// http.DefaultTransport = &loggingTransport{RoundTripper: http.DefaultTransport}

	ctx := context.Background()

	gh := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *githubToken},
	)))

	ai := openai.NewClient(*openaiKey)
	cfg := &Config{
		AI: ai,
		GH: gh,
	}

	if *iss != 0 {
		resp, err := cfg.SummarizeIssue(ctx, *iss)
		if err != nil {
			panic(err)
		}

		fmt.Println(resp)
		return
	}

	if *pr != 0 {
		resp, err := cfg.SummarizePR(ctx, *pr)
		if err != nil {
			panic(err)
		}

		fmt.Println(resp)
		return
	}
}

type loggingTransport struct {
	http.RoundTripper
}

func (s *loggingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	bytes, err := httputil.DumpRequestOut(r, true)
	if err != nil {
		return nil, err
	}
	log.Println(string(bytes))

	resp, err := s.RoundTripper.RoundTrip(r)
	if err != nil {
		return resp, err
	}

	respBytes, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return resp, err
	}
	log.Println(string(respBytes))

	return resp, err
}
