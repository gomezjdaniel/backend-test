package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlayerCRUD(t *testing.T) {
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
			Name:   "`player_id` explictly set on create",
			Method: "POST",
			Target: "/players",
			RequestSetup: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
			},
			Body: player{
				PlayerID:    int64(1),
				DisplayName: "Foo",
				Number:      23,
				Position:    POSITION_RIGHT_WING,
			},
			ExpectedStatusCode: http.StatusUnprocessableEntity,
		},
		{
			Name:   "Create first player",
			Method: "POST",
			Target: "/players",
			RequestSetup: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
			},
			Body: player{
				DisplayName: "Foo",
				Number:      1,
				Position:    POSITION_GOALKEEPER,
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedBody:       `{"player_id":1}`,
		},
		{
			Name:   "Create second player",
			Method: "POST",
			Target: "/players",
			RequestSetup: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
			},
			Body: player{
				DisplayName: "Bar",
				Number:      7,
				Position:    POSITION_STRIKER,
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedBody:       `{"player_id":2}`,
		},
		{
			Name:               "List both players",
			Method:             "GET",
			Target:             "/players",
			ExpectedStatusCode: http.StatusOK,
			ExpectedBody:       `[{"player_id":1,"display_name":"Foo","number":1,"position":"POSITION_GOALKEEPER"},{"player_id":2,"display_name":"Bar","number":7,"position":"POSITION_STRIKER"}]`,
		},
		{
			Name:               "List only strikers",
			Method:             "GET",
			Target:             "/players?position=POSITION_STRIKER",
			ExpectedStatusCode: http.StatusOK,
			ExpectedBody:       `[{"player_id":2,"display_name":"Bar","number":7,"position":"POSITION_STRIKER"}]`,
		},
		{
			Name:               "Pagination `limit` param exceeds",
			Method:             "GET",
			Target:             "/players?limit=101",
			ExpectedStatusCode: http.StatusUnprocessableEntity,
			ExpectedBody:       "{\"message\":\"`limit` cannot be greater than 100\"}",
		},
		{
			Name:               "Test pagination: page 1",
			Method:             "GET",
			Target:             "/players?limit=2&page=1",
			ExpectedStatusCode: http.StatusOK,
			ExpectedBody:       `[{"player_id":1,"display_name":"Foo","number":1,"position":"POSITION_GOALKEEPER"},{"player_id":2,"display_name":"Bar","number":7,"position":"POSITION_STRIKER"}]`,
		},
		{
			Name:               "Test pagination: page 2",
			Method:             "GET",
			Target:             "/players?limit=2&page=2",
			ExpectedStatusCode: http.StatusOK,
			ExpectedBody:       `[]`,
		},
		{
			Name:   "Make first player a striker too",
			Method: "PUT",
			Target: "/players/1",
			RequestSetup: func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
			},
			Body: player{
				Position: POSITION_STRIKER,
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:               "List only strikers",
			Method:             "GET",
			Target:             "/players",
			ExpectedStatusCode: http.StatusOK,
			ExpectedBody:       `[{"player_id":1,"display_name":"Foo","number":1,"position":"POSITION_STRIKER"},{"player_id":2,"display_name":"Bar","number":7,"position":"POSITION_STRIKER"}]`,
		},
		{
			Name:               "Delete player 1",
			Method:             "DELETE",
			Target:             "/players/1",
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:               "Delete player 2",
			Method:             "DELETE",
			Target:             "/players/2",
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:               "Check there are no players",
			Method:             "GET",
			Target:             "/players",
			ExpectedStatusCode: http.StatusOK,
			ExpectedBody:       `[]`,
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
