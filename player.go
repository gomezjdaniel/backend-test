package main

import (
	"fmt"
	"strconv"
)

type player struct {
	PlayerID    int64    `json:"player_id,omitempty" db:"player_id,omitempty"`
	DisplayName string   `json:"display_name,omitempty" db:"display_name,omitempty"`
	Number      int      `json:"number,omitempty" db:"number,omitempty"`
	Position    position `json:"position,omitempty" db:"position,omitempty"`
}

type position int

func (p position) String() string {
	s, ok := position_name[int(p)]
	if ok {
		return s
	}
	return strconv.Itoa(int(p))
}

func (p position) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *position) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := position_value[s]; ok {
		*p = position(i)
		return nil
	}
	return fmt.Errorf("Could not parse %s", b)
}

const (
	POSITION_INVALID position = iota
	POSITION_GOALKEEPER
	POSITION_DEFENDER
	POSITION_LEFT_WING
	POSITION_RIGHT_WING
	POSITION_MIDDLEFIELD
	POSITION_STRIKER
)

var position_name = map[int]string{
	0: "POSITION_INVALID",
	1: "POSITION_GOALKEEPER",
	2: "POSITION_DEFENDER",
	3: "POSITION_LEFT_WING",
	4: "POSITION_RIGHT_WING",
	5: "POSITION_MIDDLEFIELD",
	6: "POSITION_STRIKER",
}

var position_value = map[string]int{
	"POSITION_INVALID":     0,
	"POSITION_GOALKEEPER":  1,
	"POSITION_DEFENDER":    2,
	"POSITION_LEFT_WING":   3,
	"POSITION_RIGHT_WING":  4,
	"POSITION_MIDDLEFIELD": 5,
	"POSITION_STRIKER":     6,
}
