package models

// GithubWebhookPayload:
type GithubWebhookPayload struct {
	After       string                  `json:"after"`
	Head_commit GithubCommit            `json:"head_commit"`
	Pusher      GithubWebhookPusher     `json:"pusher"`
	Ref         string                  `json:"ref"`
	Repository  GithubWebhookRepository `json:"repository"`
	Sender      GithubWebhookSender     `json:"sender"`
}
