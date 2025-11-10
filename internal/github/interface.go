package github

// GitHubClient defines the interface for GitHub operations needed by commands
type GitHubClient interface {
	// PR operations
	GetPullRequests(owner, repo string) ([]*PullRequest, error)
	ClosePullRequest(owner, repo string, number int, comment string) error
	AddLabel(owner, repo string, number int, label string) error

	// User operations
	GetUser(username string) (*User, error)
	BlockUserOrg(org, username string) error
	BlockUserPersonal(username string) error
}
