package main

type actionType uint16

const (
	ACTION_INVALID actionType = iota

	ACTION_CARD_YELLOW
	ACTION_CARD_RED

	ACTION_GOAL
	ACTION_GOAL_OWN

	ACTION_ASSIST
)

type action struct {
	PlayerID  int64      `json:"player_id,omitempty" db:"player_id,omitempty"`
	Type      actionType `json:"action,omitempty" db:"action,omitempty"`
	Timestamp uint64
}
