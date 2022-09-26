package models

// GithubCommit:
type GithubCommit struct {
	Message string `json:"message"`
	Sha     string `json:"sha"`
}
