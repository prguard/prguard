package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client
type Client struct {
	client *github.Client
	ctx    context.Context
}

// NewClient creates a new GitHub API client
func NewClient(token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		client: github.NewClient(tc),
		ctx:    ctx,
	}
}

// PullRequest represents a GitHub pull request with relevant metadata
type PullRequest struct {
	Number     int
	Title      string
	Body       string
	Author     string
	CreatedAt  time.Time
	FilesCount int
	Additions  int
	Deletions  int
	Files      []string
	State      string
	HTMLURL    string
}

// User represents a GitHub user with account information
type User struct {
	Login     string
	CreatedAt time.Time
	Type      string
}

// GetPullRequests fetches all open pull requests for a repository
func (c *Client) GetPullRequests(owner, repo string) ([]*PullRequest, error) {
	opts := &github.PullRequestListOptions{
		State: "open",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allPRs []*PullRequest
	for {
		prs, resp, err := c.client.PullRequests.List(c.ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list pull requests: %w", err)
		}

		for _, pr := range prs {
			prDetails, err := c.GetPullRequest(owner, repo, pr.GetNumber())
			if err != nil {
				return nil, err
			}
			allPRs = append(allPRs, prDetails)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allPRs, nil
}

// GetPullRequest fetches detailed information about a specific PR
func (c *Client) GetPullRequest(owner, repo string, number int) (*PullRequest, error) {
	pr, _, err := c.client.PullRequests.Get(c.ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	// Fetch files changed in the PR
	files, _, err := c.client.PullRequests.ListFiles(c.ctx, owner, repo, number, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list PR files: %w", err)
	}

	var filenames []string
	for _, file := range files {
		filenames = append(filenames, file.GetFilename())
	}

	return &PullRequest{
		Number:     pr.GetNumber(),
		Title:      pr.GetTitle(),
		Body:       pr.GetBody(),
		Author:     pr.GetUser().GetLogin(),
		CreatedAt:  pr.GetCreatedAt().Time,
		FilesCount: len(files),
		Additions:  pr.GetAdditions(),
		Deletions:  pr.GetDeletions(),
		Files:      filenames,
		State:      pr.GetState(),
		HTMLURL:    pr.GetHTMLURL(),
	}, nil
}

// GetUser fetches information about a GitHub user
func (c *Client) GetUser(username string) (*User, error) {
	user, _, err := c.client.Users.Get(c.ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &User{
		Login:     user.GetLogin(),
		CreatedAt: user.GetCreatedAt().Time,
		Type:      user.GetType(),
	}, nil
}

// ClosePullRequest closes a PR with an optional comment
func (c *Client) ClosePullRequest(owner, repo string, number int, comment string) error {
	// Add comment if provided
	if comment != "" {
		issueComment := &github.IssueComment{
			Body: github.String(comment),
		}
		_, _, err := c.client.Issues.CreateComment(c.ctx, owner, repo, number, issueComment)
		if err != nil {
			return fmt.Errorf("failed to add comment: %w", err)
		}
	}

	// Close the PR
	pr := &github.PullRequest{
		State: github.String("closed"),
	}
	_, _, err := c.client.PullRequests.Edit(c.ctx, owner, repo, number, pr)
	if err != nil {
		return fmt.Errorf("failed to close pull request: %w", err)
	}

	return nil
}

// AddLabel adds a label to a pull request
func (c *Client) AddLabel(owner, repo string, number int, label string) error {
	_, _, err := c.client.Issues.AddLabelsToIssue(c.ctx, owner, repo, number, []string{label})
	if err != nil {
		return fmt.Errorf("failed to add label: %w", err)
	}
	return nil
}

// BlockUserOrg blocks a user at the organization level
// This blocks them from ALL repositories in the organization
func (c *Client) BlockUserOrg(org, username string) error {
	_, err := c.client.Organizations.BlockUser(c.ctx, org, username)
	if err != nil {
		return fmt.Errorf("failed to block user at org level: %w", err)
	}
	return nil
}

// IsUserBlockedOrg checks if a user is blocked at the organization level
func (c *Client) IsUserBlockedOrg(org, username string) (bool, error) {
	blocked, _, err := c.client.Organizations.IsBlocked(c.ctx, org, username)
	if err != nil {
		return false, fmt.Errorf("failed to check if user is blocked at org level: %w", err)
	}
	return blocked, nil
}

// BlockUserPersonal blocks a user at the personal account level
// This blocks them from ALL repositories owned by your personal account
func (c *Client) BlockUserPersonal(username string) error {
	_, err := c.client.Users.BlockUser(c.ctx, username)
	if err != nil {
		return fmt.Errorf("failed to block user at account level: %w", err)
	}
	return nil
}

// IsUserBlockedPersonal checks if a user is blocked at the personal account level
func (c *Client) IsUserBlockedPersonal(username string) (bool, error) {
	blocked, _, err := c.client.Users.IsBlocked(c.ctx, username)
	if err != nil {
		return false, fmt.Errorf("failed to check if user is blocked at account level: %w", err)
	}
	return blocked, nil
}
