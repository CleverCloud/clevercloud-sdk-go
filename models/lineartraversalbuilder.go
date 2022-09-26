package models

// LinearTraversalBuilder:
type LinearTraversalBuilder struct {
	Attributes        Attributes  `json:"attributes"`
	BeforeBuilder     Traversal   `json:"beforeBuilder"`
	Empty             bool        `json:"empty"`
	InOffset          int32       `json:"inOffset"`
	InPort            InPort      `json:"inPort"`
	InSlots           int32       `json:"inSlots"`
	IslandTag         IslandTag   `json:"islandTag"`
	OutPort           OutPort     `json:"outPort"`
	PendingBuilder    interface{} `json:"pendingBuilder"`
	TraversalComplete bool        `json:"traversalComplete"`
	TraversalSoFar    Traversal   `json:"traversalSoFar"`
}
