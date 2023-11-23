//go:build utest

package board

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kxplxn/goteam/server/api"
	"github.com/kxplxn/goteam/server/assert"
	"github.com/kxplxn/goteam/server/dbaccess"
	boardTable "github.com/kxplxn/goteam/server/dbaccess/board"
	userTable "github.com/kxplxn/goteam/server/dbaccess/user"
	pkgLog "github.com/kxplxn/goteam/server/log"
)

// TestPOSTHandler tests the Handle method of POSTHandler to assert that it
// behaves correctly in all possible scenarios.
func TestPOSTHandler(t *testing.T) {
	validator := &api.FakeStringValidator{}
	userSelector := &userTable.FakeSelector{}
	userBoardCounter := &dbaccess.FakeCounter{}
	boardInserter := &boardTable.FakeInserter{}
	log := &pkgLog.FakeErrorer{}
	sut := NewPOSTHandler(
		validator, userSelector, userBoardCounter, boardInserter, log,
	)

	t.Run(http.MethodPost, func(t *testing.T) {
		for _, c := range []struct {
			name                string
			validatorErr        error
			userRecord          userTable.Record
			userSelectorErr     error
			userBoardCount      int
			userBoardCounterErr error
			boardInserterErr    error
			wantStatusCode      int
			assertFunc          func(*testing.T, *http.Response, string)
		}{
			{
				name: "InvalidRequest",
				validatorErr: errors.New(
					"Board name cannot be empty.",
				),
				userRecord:          userTable.Record{},
				userSelectorErr:     nil,
				userBoardCount:      0,
				userBoardCounterErr: nil,
				boardInserterErr:    nil,
				wantStatusCode:      http.StatusBadRequest,
				assertFunc: assert.OnResErr(
					"Board name cannot be empty.",
				),
			},
			{
				name:                "UserNotRecognised",
				validatorErr:        nil,
				userRecord:          userTable.Record{},
				userSelectorErr:     sql.ErrNoRows,
				userBoardCount:      0,
				userBoardCounterErr: nil,
				boardInserterErr:    nil,
				wantStatusCode:      http.StatusUnauthorized,
				assertFunc: assert.OnResErr(
					"Username is not recognised.",
				),
			},
			{
				name:                "UserSelectorErr",
				validatorErr:        nil,
				userRecord:          userTable.Record{IsAdmin: false},
				userSelectorErr:     sql.ErrConnDone,
				userBoardCount:      0,
				userBoardCounterErr: nil,
				boardInserterErr:    nil,
				wantStatusCode:      http.StatusInternalServerError,
				assertFunc: assert.OnLoggedErr(
					sql.ErrConnDone.Error(),
				),
			},
			{
				name:                "UserNotAdmin",
				validatorErr:        nil,
				userRecord:          userTable.Record{IsAdmin: false},
				userSelectorErr:     nil,
				userBoardCount:      0,
				userBoardCounterErr: sql.ErrConnDone,
				boardInserterErr:    nil,
				wantStatusCode:      http.StatusForbidden,
				assertFunc: assert.OnResErr(
					"Only team admins can create boards.",
				),
			},
			{
				name:                "UserBoardCounterErr",
				validatorErr:        nil,
				userRecord:          userTable.Record{IsAdmin: true},
				userSelectorErr:     nil,
				userBoardCount:      0,
				userBoardCounterErr: sql.ErrConnDone,
				boardInserterErr:    nil,
				wantStatusCode:      http.StatusInternalServerError,
				assertFunc: assert.OnLoggedErr(
					sql.ErrConnDone.Error(),
				),
			},
			{
				name:                "MaxBoardsCreated",
				validatorErr:        nil,
				userRecord:          userTable.Record{IsAdmin: true},
				userSelectorErr:     nil,
				userBoardCount:      3,
				userBoardCounterErr: nil,
				boardInserterErr:    nil,
				wantStatusCode:      http.StatusBadRequest,
				assertFunc: assert.OnResErr(
					"You have already created the maximum amount of boards " +
						"allowed per user. Please delete one of your boards " +
						"to create a new one.",
				),
			},
			{
				name:                "BoardInserterErr",
				validatorErr:        nil,
				userRecord:          userTable.Record{IsAdmin: true},
				userSelectorErr:     nil,
				userBoardCount:      0,
				userBoardCounterErr: sql.ErrNoRows,
				boardInserterErr:    errors.New("create board error"),
				wantStatusCode:      http.StatusInternalServerError,
				assertFunc: assert.OnLoggedErr(
					"create board error",
				),
			},
			{
				name:                "Success",
				validatorErr:        nil,
				userRecord:          userTable.Record{IsAdmin: true},
				userSelectorErr:     nil,
				userBoardCount:      0,
				userBoardCounterErr: sql.ErrNoRows,
				boardInserterErr:    nil,
				wantStatusCode:      http.StatusOK,
				assertFunc: func(*testing.T, *http.Response, string) {
				},
			},
		} {
			t.Run(c.name, func(t *testing.T) {
				// Set pre-determinate return values for sut's dependencies.
				validator.Err = c.validatorErr
				userSelector.User = c.userRecord
				userSelector.Err = c.userSelectorErr
				userBoardCounter.BoardCount = c.userBoardCount
				userBoardCounter.Err = c.userBoardCounterErr
				boardInserter.Err = c.boardInserterErr

				// Prepare request and response recorder.
				reqBody, err := json.Marshal(ReqBody{})
				if err != nil {
					t.Fatal(err)
				}
				req, err := http.NewRequest(
					http.MethodPost, "", bytes.NewReader(reqBody),
				)
				if err != nil {
					t.Fatal(err)
				}
				w := httptest.NewRecorder()

				// Handle request with sut and get the result.
				sut.Handle(w, req, "")
				res := w.Result()

				// Assert on the status code.
				if err = assert.Equal(
					c.wantStatusCode, res.StatusCode,
				); err != nil {
					t.Error(err)
				}

				// Run case-specific assertions.
				c.assertFunc(t, res, log.InMessage)
			})
		}
	})
}
