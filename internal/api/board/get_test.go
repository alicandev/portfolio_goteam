//go:build utest

package board

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kxplxn/goteam/internal/api"
	"github.com/kxplxn/goteam/pkg/assert"
	boardTable "github.com/kxplxn/goteam/pkg/dbaccess/board"
	teamTable "github.com/kxplxn/goteam/pkg/dbaccess/team"
	userTable "github.com/kxplxn/goteam/pkg/dbaccess/user"
	pkgLog "github.com/kxplxn/goteam/pkg/log"
)

// TestGETHandler tests the Handle method of GETHandler to assert that it
// behaves correctly in all possible scenarios.
func TestGETHandler(t *testing.T) {
	userSelector := &userTable.FakeSelector{}
	boardInserter := &boardTable.FakeInserter{}
	idValidator := &api.FakeStringValidator{}
	boardSelectorRecursive := &boardTable.FakeSelectorRecursive{}
	teamSelector := &teamTable.FakeSelector{}
	userSelectorByTeamID := &userTable.FakeSelectorByTeamID{}
	boardSelectorByTeamID := &boardTable.FakeSelectorByTeamID{}
	log := &pkgLog.FakeErrorer{}

	sut := NewGETHandler(
		userSelector,
		boardInserter,
		idValidator,
		boardSelectorRecursive,
		teamSelector,
		userSelectorByTeamID,
		boardSelectorByTeamID,
		log,
	)

	for _, c := range []struct {
		name                      string
		boardID                   string
		user                      userTable.Record
		userSelectorErr           error
		boardInserterErr          error
		idValidatorErr            error
		team                      teamTable.Record
		teamSelectorErr           error
		members                   []userTable.Record
		userSelectorByTeamIDErr   error
		boards                    []boardTable.Record
		boardSelectorByTeamIDErr  error
		activeBoard               boardTable.RecursiveRecord
		boardSelectorRecursiveErr error
		wantStatusCode            int
		assertFunc                func(*testing.T, *http.Response, string)
	}{
		{
			name:                      "UserIsNotRecognised",
			boardID:                   "",
			user:                      userTable.Record{},
			userSelectorErr:           sql.ErrNoRows,
			boardInserterErr:          nil,
			idValidatorErr:            nil,
			team:                      teamTable.Record{},
			teamSelectorErr:           nil,
			members:                   []userTable.Record{},
			userSelectorByTeamIDErr:   nil,
			boards:                    []boardTable.Record{},
			boardSelectorByTeamIDErr:  nil,
			activeBoard:               boardTable.RecursiveRecord{},
			boardSelectorRecursiveErr: nil,
			wantStatusCode:            http.StatusUnauthorized,
			assertFunc:                assert.OnLoggedErr(sql.ErrNoRows.Error()),
		},
		{
			name:                      "UserSelectorErr",
			boardID:                   "",
			user:                      userTable.Record{},
			userSelectorErr:           sql.ErrConnDone,
			boardInserterErr:          nil,
			idValidatorErr:            nil,
			team:                      teamTable.Record{},
			teamSelectorErr:           nil,
			members:                   []userTable.Record{},
			userSelectorByTeamIDErr:   nil,
			boards:                    []boardTable.Record{},
			boardSelectorByTeamIDErr:  nil,
			activeBoard:               boardTable.RecursiveRecord{},
			boardSelectorRecursiveErr: nil,
			wantStatusCode:            http.StatusInternalServerError,
			assertFunc:                assert.OnLoggedErr(sql.ErrConnDone.Error()),
		},
		{
			name:                      "BoardInserterErr",
			boardID:                   "",
			user:                      userTable.Record{IsAdmin: true},
			userSelectorErr:           nil,
			boardInserterErr:          errors.New("error inserting board"),
			idValidatorErr:            nil,
			team:                      teamTable.Record{},
			teamSelectorErr:           nil,
			members:                   []userTable.Record{},
			userSelectorByTeamIDErr:   nil,
			boards:                    []boardTable.Record{},
			boardSelectorByTeamIDErr:  nil,
			activeBoard:               boardTable.RecursiveRecord{},
			boardSelectorRecursiveErr: nil,
			wantStatusCode:            http.StatusInternalServerError,
			assertFunc: assert.OnLoggedErr(
				"error inserting board",
			),
		},
		{
			name:                      "InvalidID",
			boardID:                   "foo",
			user:                      userTable.Record{},
			userSelectorErr:           nil,
			boardInserterErr:          nil,
			idValidatorErr:            errors.New("error invalid id"),
			team:                      teamTable.Record{},
			teamSelectorErr:           nil,
			members:                   []userTable.Record{},
			userSelectorByTeamIDErr:   nil,
			boards:                    []boardTable.Record{},
			boardSelectorByTeamIDErr:  nil,
			activeBoard:               boardTable.RecursiveRecord{},
			boardSelectorRecursiveErr: nil,
			wantStatusCode:            http.StatusBadRequest,
			assertFunc: func(
				_ *testing.T, _ *http.Response, _ string,
			) {
			},
		},
		{
			name:                      "TeamSelectorErr",
			boardID:                   "",
			user:                      userTable.Record{IsAdmin: true},
			userSelectorErr:           nil,
			boardInserterErr:          nil,
			idValidatorErr:            nil,
			team:                      teamTable.Record{},
			teamSelectorErr:           sql.ErrNoRows,
			members:                   []userTable.Record{},
			userSelectorByTeamIDErr:   nil,
			boards:                    []boardTable.Record{},
			boardSelectorByTeamIDErr:  nil,
			activeBoard:               boardTable.RecursiveRecord{},
			boardSelectorRecursiveErr: nil,
			wantStatusCode:            http.StatusInternalServerError,
			assertFunc:                assert.OnLoggedErr(sql.ErrNoRows.Error()),
		},
		{
			name:                      "UserSelectorByTeamIDErr",
			boardID:                   "",
			user:                      userTable.Record{IsAdmin: true},
			userSelectorErr:           nil,
			boardInserterErr:          nil,
			idValidatorErr:            nil,
			team:                      teamTable.Record{},
			teamSelectorErr:           nil,
			members:                   []userTable.Record{},
			userSelectorByTeamIDErr:   sql.ErrNoRows,
			boards:                    []boardTable.Record{},
			boardSelectorByTeamIDErr:  nil,
			activeBoard:               boardTable.RecursiveRecord{},
			boardSelectorRecursiveErr: nil,
			wantStatusCode:            http.StatusInternalServerError,
			assertFunc:                assert.OnLoggedErr(sql.ErrNoRows.Error()),
		},
		{
			name:                      "BoardSelectorByTeamIDErr",
			boardID:                   "",
			user:                      userTable.Record{IsAdmin: true},
			userSelectorErr:           nil,
			boardInserterErr:          nil,
			idValidatorErr:            nil,
			team:                      teamTable.Record{},
			teamSelectorErr:           nil,
			members:                   []userTable.Record{},
			userSelectorByTeamIDErr:   nil,
			boards:                    []boardTable.Record{},
			boardSelectorByTeamIDErr:  sql.ErrNoRows,
			activeBoard:               boardTable.RecursiveRecord{},
			boardSelectorRecursiveErr: nil,
			wantStatusCode:            http.StatusInternalServerError,
			assertFunc:                assert.OnLoggedErr(sql.ErrNoRows.Error()),
		},
		{
			name:                      "NoBoards",
			boardID:                   "",
			user:                      userTable.Record{IsAdmin: true},
			userSelectorErr:           nil,
			boardInserterErr:          nil,
			idValidatorErr:            nil,
			team:                      teamTable.Record{},
			teamSelectorErr:           nil,
			members:                   []userTable.Record{},
			userSelectorByTeamIDErr:   sql.ErrNoRows,
			boards:                    []boardTable.Record{},
			boardSelectorByTeamIDErr:  nil,
			activeBoard:               boardTable.RecursiveRecord{},
			boardSelectorRecursiveErr: nil,
			wantStatusCode:            http.StatusInternalServerError,
			assertFunc:                assert.OnLoggedErr(sql.ErrNoRows.Error()),
		},
		{
			name:                      "RecursiveBoardNotFound",
			boardID:                   "1",
			user:                      userTable.Record{IsAdmin: false},
			userSelectorErr:           nil,
			boardInserterErr:          nil,
			idValidatorErr:            nil,
			team:                      teamTable.Record{},
			teamSelectorErr:           nil,
			members:                   []userTable.Record{},
			userSelectorByTeamIDErr:   nil,
			boards:                    []boardTable.Record{},
			boardSelectorByTeamIDErr:  nil,
			activeBoard:               boardTable.RecursiveRecord{},
			boardSelectorRecursiveErr: sql.ErrNoRows,
			wantStatusCode:            http.StatusNotFound,
			assertFunc:                func(_ *testing.T, _ *http.Response, _ string) {},
		},
		{
			name:                      "BoardSelectorErr",
			boardID:                   "1",
			user:                      userTable.Record{IsAdmin: false},
			userSelectorErr:           nil,
			boardInserterErr:          nil,
			idValidatorErr:            nil,
			team:                      teamTable.Record{},
			teamSelectorErr:           nil,
			members:                   []userTable.Record{},
			userSelectorByTeamIDErr:   nil,
			boards:                    []boardTable.Record{},
			boardSelectorByTeamIDErr:  nil,
			activeBoard:               boardTable.RecursiveRecord{},
			boardSelectorRecursiveErr: sql.ErrConnDone,
			wantStatusCode:            http.StatusInternalServerError,
			assertFunc:                assert.OnLoggedErr(sql.ErrConnDone.Error()),
		},
		{
			name:    "BoardWrongTeam",
			boardID: "1",
			user: userTable.Record{
				TeamID: 1, IsAdmin: false,
			},
			userSelectorErr:           nil,
			boardInserterErr:          nil,
			idValidatorErr:            nil,
			team:                      teamTable.Record{},
			teamSelectorErr:           nil,
			members:                   []userTable.Record{},
			userSelectorByTeamIDErr:   nil,
			boards:                    []boardTable.Record{},
			boardSelectorByTeamIDErr:  nil,
			activeBoard:               boardTable.RecursiveRecord{TeamID: 2},
			boardSelectorRecursiveErr: nil,
			wantStatusCode:            http.StatusForbidden,
			assertFunc:                func(_ *testing.T, _ *http.Response, _ string) {},
		},
		{
			name:             "OK",
			boardID:          "",
			user:             userTable.Record{Username: "bob", IsAdmin: true},
			userSelectorErr:  nil,
			boardInserterErr: nil,
			idValidatorErr:   nil,
			team:             teamTable.Record{ID: 1, InviteCode: "InvCode"},
			teamSelectorErr:  nil,
			members: []userTable.Record{
				{Username: "foo", IsAdmin: true},
				{Username: "bob123", IsAdmin: false},
			},
			userSelectorByTeamIDErr: nil,
			boards: []boardTable.Record{
				{ID: 1, Name: "board 1", TeamID: 1},
				{ID: 2, Name: "board 2", TeamID: 1},
				{ID: 3, Name: "board 3", TeamID: 1},
			},
			boardSelectorByTeamIDErr: nil,
			activeBoard: func() boardTable.RecursiveRecord {
				task1Desc := "task1Desc"
				task2Desc := "task2Desc"
				return boardTable.RecursiveRecord{
					ID: 2, Name: "Active", Columns: []boardTable.Column{
						{ID: 3, Order: 1, Tasks: []boardTable.Task{}},
						{ID: 4, Order: 2, Tasks: []boardTable.Task{
							{
								ID:          5,
								Title:       "task1title",
								Description: &task1Desc,
								Order:       3,
								Subtasks: []boardTable.Subtask{
									{
										ID:     5,
										Title:  "subtask1",
										Order:  4,
										IsDone: true,
									},
									{
										ID:     6,
										Title:  "subtask2",
										Order:  5,
										IsDone: false,
									},
								},
							},
							{
								ID:          7,
								Title:       "task2title",
								Description: &task2Desc,
								Order:       6,
								Subtasks:    []boardTable.Subtask{},
							},
						}},
					},
				}
			}(),
			boardSelectorRecursiveErr: nil,
			wantStatusCode:            http.StatusOK,
			assertFunc: func(t *testing.T, r *http.Response, _ string) {
				var resp GETResp
				if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
					t.Fatal(err)
				}

				assert.Equal(t.Error, resp.User.Username, "bob")
				assert.True(t.Error, resp.User.IsAdmin)
				assert.Equal(t.Error, resp.Team.ID, 1)
				assert.Equal(t.Error, "InvCode", resp.Team.InviteCode)

				for i, wantMember := range []TeamMember{
					{Username: "foo", IsAdmin: true},
					{Username: "bob123", IsAdmin: false},
				} {
					member := resp.TeamMembers[i]

					assert.Equal(t.Error, member.Username, wantMember.Username)
					assert.Equal(t.Error, member.IsAdmin, wantMember.IsAdmin)
				}

				for i, wantBoard := range []Board{
					{ID: 1, Name: "board 1"},
					{ID: 2, Name: "board 2"},
					{ID: 3, Name: "board 3"},
				} {
					board := resp.Boards[i]

					assert.Equal(t.Error, board.ID, wantBoard.ID)
					assert.Equal(t.Error, board.Name, wantBoard.Name)
				}

				assert.Equal(t.Error, resp.ActiveBoard.ID, 2)
				assert.Equal(t.Error, resp.ActiveBoard.Name, "Active")
				assert.Equal(t.Error, len(resp.ActiveBoard.Columns), 2)
				for i, wantCol := range boardSelectorRecursive.Rec.Columns {
					col := resp.ActiveBoard.Columns[i]

					assert.Equal(t.Error, col.ID, wantCol.ID)
					assert.Equal(t.Error, col.Order, wantCol.Order)
					assert.Equal(t.Error, len(col.Tasks), len(wantCol.Tasks))

					for j, wantTask := range wantCol.Tasks {
						task := col.Tasks[j]

						assert.Equal(t.Error, task.ID, wantTask.ID)
						assert.Equal(t.Error, task.Title, wantTask.Title)
						assert.Equal(t.Error,
							task.Description, *wantTask.Description,
						)
						assert.Equal(t.Error,
							len(task.Subtasks), len(wantTask.Subtasks),
						)

						for k, wantSubtask := range wantTask.Subtasks {
							subtask := task.Subtasks[k]

							assert.Equal(t.Error, subtask.ID, wantSubtask.ID)
							assert.Equal(t.Error,
								subtask.Title, wantSubtask.Title,
							)
							assert.Equal(t.Error,
								subtask.Order, wantSubtask.Order,
							)
							assert.Equal(t.Error,
								subtask.IsDone, wantSubtask.IsDone,
							)
						}
					}
				}
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			userSelector.Rec = c.user
			userSelector.Err = c.userSelectorErr
			boardInserter.Err = c.boardInserterErr
			idValidator.Err = c.idValidatorErr
			teamSelector.Rec = c.team
			teamSelector.Err = c.teamSelectorErr
			userSelectorByTeamID.Recs = c.members
			userSelectorByTeamID.Err = c.userSelectorByTeamIDErr
			boardSelectorByTeamID.Recs = c.boards
			boardSelectorByTeamID.Err = c.boardSelectorByTeamIDErr
			boardSelectorRecursive.Rec = c.activeBoard
			boardSelectorRecursive.Err = c.boardSelectorRecursiveErr

			r := httptest.NewRequest(http.MethodGet, "/?id="+c.boardID, nil)
			w := httptest.NewRecorder()

			sut.Handle(w, r, "bob123")

			res := w.Result()
			assert.Equal(t.Error, res.StatusCode, c.wantStatusCode)
			c.assertFunc(t, res, log.InMessage)
		})
	}
}
