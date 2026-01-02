package domain

type AppActionType string

const (
	AppActionPass       AppActionType = "PASS"
	AppActionLike       AppActionType = "LIKE"
	AppActionSuperSwipe AppActionType = "SUPERSWIPE"
)

type AppAction struct {
	AType   AppActionType
	Message string
}
