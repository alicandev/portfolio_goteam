package board

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"server/assert"
	"server/db"
)

func TestPOSTHandler(t *testing.T) {
	userBoardCounter := &db.FakeCounter{}
	boardInserter := &db.FakeBoardInserter{}
	sut := NewPOSTHandler(userBoardCounter, boardInserter)
	sub := "bob123"

	t.Run(http.MethodPost, func(t *testing.T) {
		for _, c := range []struct {
			name                   string
			reqBody                ReqBody
			userBoardCounterOutRes int
			boardInserterOutErr    error
			wantStatusCode         int
			wantErr                string
		}{
			{
				name:                   "BoardNameNil",
				reqBody:                ReqBody{},
				userBoardCounterOutRes: 0,
				boardInserterOutErr:    nil,
				wantStatusCode:         http.StatusBadRequest,
				wantErr:                errNameEmpty,
			},
			{
				name:                   "BoardNameEmpty",
				reqBody:                ReqBody{Name: ""},
				userBoardCounterOutRes: 0,
				boardInserterOutErr:    nil,
				wantStatusCode:         http.StatusBadRequest,
				wantErr:                errNameEmpty,
			},
			{
				name:                   "BoardNameTooLong",
				reqBody:                ReqBody{Name: "boardyboardsyboardkyboardishboardxyz"},
				userBoardCounterOutRes: 0,
				boardInserterOutErr:    nil,
				wantStatusCode:         http.StatusBadRequest,
				wantErr:                errNameTooLong,
			},
			{
				name:                   "MaxBoardsCreated",
				reqBody:                ReqBody{Name: "someboard"},
				userBoardCounterOutRes: 3,
				boardInserterOutErr:    nil,
				wantStatusCode:         http.StatusBadRequest,
				wantErr:                errMaxBoards,
			},
			{
				name:                   "BoardCreatorError",
				reqBody:                ReqBody{Name: "someboard"},
				userBoardCounterOutRes: 0,
				boardInserterOutErr:    errors.New("board inserter error"),
				wantStatusCode:         http.StatusInternalServerError,
				wantErr:                "",
			},
		} {
			t.Run(c.name, func(t *testing.T) {
				userBoardCounter.OutRes = c.userBoardCounterOutRes
				boardInserter.OutErr = c.boardInserterOutErr

				reqBodyJSON, err := json.Marshal(c.reqBody)
				if err != nil {
					t.Fatal(err)
				}
				req, err := http.NewRequest(
					http.MethodPost, "/board", bytes.NewReader(reqBodyJSON),
				)
				if err != nil {
					t.Fatal(err)
				}

				w := httptest.NewRecorder()

				sut.Handle(w, req, sub)

				if err = assert.Equal(
					c.wantStatusCode, w.Result().StatusCode,
				); err != nil {
					t.Error(err)
				}

				// if 400 is expected - there must be a validation error in response body
				if c.wantStatusCode == http.StatusBadRequest {
					resBody := ResBody{}
					if err := json.NewDecoder(w.Result().Body).Decode(
						&resBody,
					); err != nil {
						t.Error(err)
					}
					if err := assert.Equal(c.wantErr, resBody.Error); err != nil {
						t.Error(err)
					}
				}

				// DEPENDENCY-INPUT-BASED ASSERTIONS

				// if max boards is not reached, board creator must be called
				if c.userBoardCounterOutRes >= maxBoards ||
					c.reqBody.Name == "" ||
					len(c.reqBody.Name) > maxNameLength {
					return
				}
				if err := assert.Equal(
					db.NewBoard(c.reqBody.Name, sub),
					boardInserter.InBoard,
				); err != nil {
					t.Error(err)
				}
			})
		}
	})
}
