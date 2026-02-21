package github

import (
	"fmt"
	"time"
)

// WebhookPayload 是 GitHub webhook 的主要結構
type WebhookPayload struct {
	Action      string       `json:"action"` // opened, synchronize, closed, etc.
	PullRequest *PullRequest `json:"pull_request,omitempty"`
	Review            *Review      `json:"review,omitempty"`
	RequestedReviewer *User        `json:"requested_reviewer,omitempty"`
	WorkflowRun       *WorkflowRun `json:"workflow_run,omitempty"`
	Repository        Repository   `json:"repository"`
	Sender            User         `json:"sender"`
}

type PullRequest struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"` // open, closed
	HTMLURL   string    `json:"html_url"`
	DiffURL   string    `json:"diff_url"`
	User      User      `json:"user"`
	Base      Branch    `json:"base"`
	Head      Branch    `json:"head"`
	Merged    bool      `json:"merged"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Additions int       `json:"additions"`
	Deletions int       `json:"deletions"`
}

type Review struct {
	ID          int       `json:"id"`
	User        User      `json:"user"`
	Body        string    `json:"body"`
	State       string    `json:"state"` // approved, changes_requested, commented
	HTMLURL     string    `json:"html_url"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type WorkflowRun struct {
	ID           int              `json:"id"`
	Name         string           `json:"name"`
	HeadSHA      string           `json:"head_sha"`
	Status       string           `json:"status"`     // completed
	Conclusion   string           `json:"conclusion"` // success, failure, timed_out, cancelled
	HTMLURL      string           `json:"html_url"`
	PullRequests []WorkflowRunPR  `json:"pull_requests"`
}

type WorkflowRunPR struct {
	Number int `json:"number"`
}

type Repository struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"` // owner/repo
	HTMLURL  string `json:"html_url"`
}

type User struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
}

type Branch struct {
	Ref string `json:"ref"` // branch name
	SHA string `json:"sha"`
}

// GetPRIdentifier 回傳唯一識別這個 PR 的 key
// 格式: "owner/repo#123"
func (w *WebhookPayload) GetPRIdentifier() string {
	if w.PullRequest != nil {
		return fmt.Sprintf("%s#%d", w.Repository.FullName, w.PullRequest.Number)
	}
	return ""
}
