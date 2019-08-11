package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/require"
)

func TestLineupCRUD(t *testing.T) {
	s := testServer()
	defer s.db.Close()

	for _, tc := range []struct {
		Name               string
		Method             string
		Target             string
		RequestSetup       func(*http.Request)
		Body               interface{}
		ExpectedStatusCode int
		ExpectedBody       string
	}{
		{
			Name:   "`lineup_id` explictly set on create",
			Method: "POST",
			Target: "/lineups",
			RequestSetup: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
			},
			Body: lineup{
				LineupID:  int64(1),
				Formation: FORMATION_FOUR_FOUR_TWO,
				IsLocal:   boolPtr(true),
			},
			ExpectedStatusCode: http.StatusUnprocessableEntity,
		},
		{
			Name:   "Create first lineup",
			Method: "POST",
			Target: "/lineups",
			RequestSetup: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
			},
			Body: lineup{
				Formation: FORMATION_FOUR_FOUR_TWO,
				IsLocal:   boolPtr(true),
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedBody:       `{"lineup_id":1}`,
		},
		{
			Name:   "Create second lineup",
			Method: "POST",
			Target: "/lineups",
			RequestSetup: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
			},
			Body: lineup{
				Formation: FORMATION_FOUR_THREE_THREE,
				IsLocal:   boolPtr(false),
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedBody:       `{"lineup_id":2}`,
		},
		{
			Name:               "Get first lineup",
			Method:             "GET",
			Target:             "/lineups/1",
			ExpectedStatusCode: http.StatusOK,
			ExpectedBody:       `{"lineup_id":1,"formation":"FORMATION_FOUR_FOUR_TWO","is_local":true}`,
		},
		{
			Name:   "Update second lineup",
			Method: "PUT",
			Target: "/lineups/2",
			RequestSetup: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
			},
			Body: lineup{
				IsLocal: boolPtr(true),
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:               "Get second lineup",
			Method:             "GET",
			Target:             "/lineups/2",
			ExpectedStatusCode: http.StatusOK,
			ExpectedBody:       `{"lineup_id":2,"formation":"FORMATION_FOUR_THREE_THREE","is_local":true}`,
		},
		{
			Name:               "Delete second lineup",
			Method:             "DELETE",
			Target:             "/lineups/2",
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:               "Attempt to get second lineup",
			Method:             "GET",
			Target:             "/lineups/2",
			ExpectedStatusCode: http.StatusNotFound,
			ExpectedBody:       `{"message":"lineup not found"}`,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			r := require.New(t)

			var req *http.Request

			if tc.Body == nil {
				req = httptest.NewRequest(tc.Method, tc.Target, nil)
			} else {
				body, err := json.Marshal(tc.Body)
				r.Nil(err)
				req = httptest.NewRequest(tc.Method, tc.Target, bytes.NewBuffer(body))
			}

			if tc.RequestSetup != nil {
				tc.RequestSetup(req)
			}

			rec := httptest.NewRecorder()

			s.web.ServeHTTP(rec, req)
			resp := rec.Result()
			r.Equal(tc.ExpectedStatusCode, resp.StatusCode)

			data, err := ioutil.ReadAll(resp.Body)
			r.Nil(err)
			r.Equal(tc.ExpectedBody, strings.TrimRight(string(data), "\n"))
		})
	}
}

func TestLineup11Players(t *testing.T) {
	s := testServer()
	defer s.db.Close()

	r := require.New(t)

	player1 := player{
		PlayerID:    int64(1),
		DisplayName: "Foo",
		Number:      1,
		Position:    POSITION_RIGHT_WING,
	}
	_, err := s.db.Collection(playersTable).Insert(&player1)
	r.Nil(err)

	player2 := player{
		PlayerID:    int64(2),
		DisplayName: "Bar",
		Number:      4,
		Position:    POSITION_RIGHT_WING,
	}
	_, err = s.db.Collection(playersTable).Insert(&player2)
	r.Nil(err)

	lineup := lineup{
		LineupID:  int64(1),
		Formation: FORMATION_FOUR_FOUR_TWO,
		IsLocal:   boolPtr(false),
	}

	_, err = s.db.Collection(lineupsTable).Insert(&lineup)
	r.Nil(err)

	for _, tc := range []struct {
		Name               string
		Method             string
		Target             string
		RequestSetup       func(*http.Request)
		Body               interface{}
		ExpectedStatusCode int
		ExpectedBody       string
	}{
		{
			Name:               "Get lineup",
			Method:             "GET",
			Target:             "/lineups/1",
			ExpectedStatusCode: http.StatusOK,
			ExpectedBody:       `{"lineup_id":1,"formation":"FORMATION_FOUR_FOUR_TWO","is_local":false}`,
		},
		{
			Name:   "Add player 1 to lineup",
			Method: "POST",
			Target: "/lineups/1/players",
			RequestSetup: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
			},
			Body: player{
				PlayerID: player1.PlayerID,
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:   "Add player 2 to lineup",
			Method: "POST",
			Target: "/lineups/1/players",
			RequestSetup: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
			},
			Body: player{
				PlayerID: player2.PlayerID,
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:               "Get lineup with players",
			Method:             "GET",
			Target:             "/lineups/1?with-players=true",
			ExpectedStatusCode: http.StatusOK,
			ExpectedBody:       `{"lineup_id":1,"formation":"FORMATION_FOUR_FOUR_TWO","is_local":false,"players":[{"player_id":1,"display_name":"Foo","number":1,"position":"POSITION_RIGHT_WING"},{"player_id":2,"display_name":"Bar","number":4,"position":"POSITION_RIGHT_WING"}]}`,
		},
		{
			Name:   "Delete player 2 from lineup",
			Method: "DELETE",
			Target: "/lineups/1/players",
			RequestSetup: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
			},
			Body: player{
				PlayerID: player1.PlayerID,
			},
			ExpectedStatusCode: http.StatusOK,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			r := require.New(t)

			var req *http.Request

			if tc.Body == nil {
				req = httptest.NewRequest(tc.Method, tc.Target, nil)
			} else {
				body, err := json.Marshal(tc.Body)
				r.Nil(err)
				req = httptest.NewRequest(tc.Method, tc.Target, bytes.NewBuffer(body))
			}

			if tc.RequestSetup != nil {
				tc.RequestSetup(req)
			}

			rec := httptest.NewRecorder()

			s.web.ServeHTTP(rec, req)
			resp := rec.Result()
			r.Equal(tc.ExpectedStatusCode, resp.StatusCode)

			data, err := ioutil.ReadAll(resp.Body)
			r.Nil(err)
			r.Equal(tc.ExpectedBody, strings.TrimRight(string(data), "\n"))
		})
	}
}

func TestLineupPlayersLimit(t *testing.T) {
	s := testServer()
	defer s.db.Close()

	r := require.New(t)

	lineup := lineup{
		LineupID:  int64(1),
		Formation: FORMATION_FOUR_FOUR_TWO,
		IsLocal:   boolPtr(false),
	}

	_, err := s.db.Collection(lineupsTable).Insert(&lineup)
	r.Nil(err)

	var players [11]player

	for i := range players {
		r.Nil(faker.FakeData(&players[i]))
		players[i].PlayerID = int64(0)

		_, err := s.db.Collection(playersTable).Insert(&players[i])
		r.Nil(err)

		_, err = s.db.Collection(lineupPlayersTable).Insert(&struct {
			LineupID int64 `db:"lineup_id"`
			PlayerID int64 `db:"player_id"`
		}{
			LineupID: lineup.LineupID,
			PlayerID: int64(i + 1),
		})
	}

	_, err = s.db.Collection(playersTable).Insert(&player{
		PlayerID:    int64(12),
		DisplayName: "John Doe",
		Position:    POSITION_RIGHT_WING,
		Number:      28,
	})
	r.Nil(err)

	for _, tc := range []struct {
		Name               string
		Method             string
		Target             string
		RequestSetup       func(*http.Request)
		Body               interface{}
		ExpectedStatusCode int
		ExpectedBody       string
	}{
		{
			Name:   "Add 12th player",
			Method: "POST",
			Target: "/lineups/1/players",
			RequestSetup: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
			},
			Body: player{
				PlayerID: int64(12),
			},
			ExpectedStatusCode: http.StatusForbidden,
			ExpectedBody:       `{"message":"lineup has reached maximum players"}`,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			r := require.New(t)

			var req *http.Request

			if tc.Body == nil {
				req = httptest.NewRequest(tc.Method, tc.Target, nil)
			} else {
				body, err := json.Marshal(tc.Body)
				r.Nil(err)
				req = httptest.NewRequest(tc.Method, tc.Target, bytes.NewBuffer(body))
			}

			if tc.RequestSetup != nil {
				tc.RequestSetup(req)
			}

			rec := httptest.NewRecorder()

			s.web.ServeHTTP(rec, req)
			resp := rec.Result()
			r.Equal(tc.ExpectedStatusCode, resp.StatusCode)

			data, err := ioutil.ReadAll(resp.Body)
			r.Nil(err)
			r.Equal(tc.ExpectedBody, strings.TrimRight(string(data), "\n"))
		})
	}
}
