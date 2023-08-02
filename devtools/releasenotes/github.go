package main

import (
	"context"
	"fmt"
)

type PRData struct {
	Author string
	Labels []string
	Title  string
	Body   string
	Branch string
}

func (g *Generator) ReadPR(ctx context.Context, id int) (*PRData, error) {
	pr, _, err := g.gh.PullRequests.Get(ctx, "target", "goalert", int(id))
	if err != nil {
		return nil, err
	}

	result := &PRData{
		Author: pr.User.GetLogin(),
		Labels: make([]string, len(pr.Labels)),
		Title:  pr.GetTitle(),
		Body:   pr.GetBody(),
		Branch: pr.Head.GetRef(),
	}
	for i, l := range pr.Labels {
		result.Labels[i] = l.GetName()
	}

	return result, nil
}

type IssueData struct {
	Labels   []string
	State    string
	Title    string
	Body     string
	Author   string
	Comments []Comment
}
type Comment struct {
	Author string
	Body   string
}

func (g *Generator) ReadIssue(ctx context.Context, id int) (*IssueData, error) {
	iss, _, err := g.gh.Issues.Get(ctx, "target", "goalert", int(id))
	if err != nil {
		return nil, fmt.Errorf("get issue #%d: %w", id, err)
	}

	comments, _, err := g.gh.Issues.ListComments(ctx, "target", "goalert", int(id), nil)
	if err != nil {
		return nil, fmt.Errorf("list comments for issue #%d: %w", id, err)
	}

	result := &IssueData{
		Labels:   make([]string, len(iss.Labels)),
		State:    iss.GetState(),
		Title:    iss.GetTitle(),
		Body:     iss.GetBody(),
		Author:   iss.User.GetLogin(),
		Comments: make([]Comment, len(comments)),
	}

	if result.Author == "dependabot[bot]" {
		result.Body = "Body truncated: This is a PR to update a dependency."
	}
	for i, l := range iss.Labels {
		result.Labels[i] = l.GetName()
	}
	for i, c := range comments {
		result.Comments[i] = Comment{
			Author: c.User.GetLogin(),
			Body:   c.GetBody(),
		}
	}

	return result, nil
}
