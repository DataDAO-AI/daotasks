package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/github"
)

const (
	todoOwner = "DataDAO-AI"
	todoRepo  = "todo"
)

// Get the issues from the todo repository. The state parameter can be "open", "closed", or "all".
func issues(state string) ([]*github.Issue, error) {
	if g == nil {
		return nil, fmt.Errorf("GitHub client not initialized")
	}

	if state != "open" && state != "closed" && state != "all" {
		return nil, fmt.Errorf("invalid state provided: %s", state)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opt := &github.IssueListByRepoOptions{
		State: "open",
	}

	issues, _, err := g.Issues.ListByRepo(ctx, todoOwner, todoRepo, opt)
	if err != nil {
		return nil, err
	}

	return issues, nil
}

// Get a string describing a GitHub issue.
func issueDescription(issue *github.Issue) (string, error) {
	var content bytes.Buffer

	if issue == nil || issue.Title == nil || issue.HTMLURL == nil || issue.CreatedAt == nil || issue.UpdatedAt == nil || issue.Comments == nil {
		return "", fmt.Errorf("issue %d has nil fields", issue.GetNumber())
	}

	content.WriteString(fmt.Sprintf("Issue: [**%s**](<%s>) (%d comments)\n", *issue.Title, *issue.HTMLURL, *issue.Comments))

	if len(issue.Assignees) == 0 {
		content.WriteString("No assignees.")
	} else {
		content.WriteString("Assigned to:")
		for _, assignee := range issue.Assignees {
			if assignee.Login != nil {
				content.WriteString(fmt.Sprintf(" %s", *assignee.Login))
			}
		}
	}
	content.WriteString("\n")

	content.WriteString(fmt.Sprintf("Created at <t:%d> (<t:%d:R>), updated at <t:%d> (<t:%d:R>)\n",
		issue.CreatedAt.Unix(), issue.CreatedAt.Unix(), issue.UpdatedAt.Unix(), issue.UpdatedAt.Unix()))

	return content.String(), nil
}

func issueDescriptions(issues []*github.Issue) string {
	var content bytes.Buffer
	for i, issue := range issues {
		description, err := issueDescription(issue)
		if err != nil {
			log.Printf("Error getting issue description: %v\n", err)
			continue
		}

		if i > 0 {
			content.WriteString(strings.Repeat("━", 10) + "\n")
		}
		content.WriteString(description)
	}

	if content.Len() == 0 {
		return "No open issues found."
	}

	return content.String()
}
