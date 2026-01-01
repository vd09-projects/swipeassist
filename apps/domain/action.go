package domain

type ActionType string

const (
	ActionPass       ActionType = "PASS"
	ActionLike       ActionType = "LIKE"
	ActionSuperSwipe ActionType = "SUPERSWIPE"
)

type Action struct {
	AType   ActionType
	Message string
}
