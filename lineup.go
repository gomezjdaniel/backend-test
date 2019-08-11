package main

import (
	"fmt"
	"strconv"
)

type formation uint16

func (f formation) String() string {
	s, ok := formation_name[int(f)]
	if ok {
		return s
	}
	return strconv.Itoa(int(f))
}

func (f formation) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

func (f *formation) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := formation_value[s]; ok {
		*f = formation(i)
		return nil
	}
	return fmt.Errorf("Could not parse %s", b)
}

const (
	FORMATION_INVALID formation = iota
	FORMATION_FOUR_FOUR_TWO
	FORMATION_FOUR_THREE_THREE
	FORMATION_THREE_FOUR_THREE
)

var formation_name = map[int]string{
	0: "FORMATION_INVALID",
	1: "FORMATION_FOUR_FOUR_TWO",
	2: "FORMATION_FOUR_THREE_THREE",
	3: "FORMATION_THREE_FOUR_THREE",
}

var formation_value = map[string]int{
	"FORMATION_INVALID":          0,
	"FORMATION_FOUR_FOUR_TWO":    1,
	"FORMATION_FOUR_THREE_THREE": 2,
	"FORMATION_THREE_FOUR_THREE": 3,
}

type lineup struct {
	LineupID  int64     `json:"lineup_id,omitempty" db:"lineup_id,omitempty"`
	Formation formation `json:"formation,omitempty" db:"formation,omitempty"`
	IsLocal   *bool     `json:"is_local,omitempty" db:"is_local,omitempty"`
}
