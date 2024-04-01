package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	mockdb "master_class/db/mock"
	db "master_class/db/sqlc"
	"master_class/db/util"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type EqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e EqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := util.CheckPasswordHash(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = arg.HashedPassword

	return reflect.DeepEqual(e.arg, arg)
}

func (e EqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return EqCreateUserParamsMatcher{arg: arg, password: password}
}

type createUserTestCases struct {
	name          string
	request       createUserRequest
	buildStubs    func(store *mockdb.MockStore)
	checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
}

type changeUserPasswordTestCases struct {
	name          string
	request       changeUserPasswordRequest
	buildStubs    func(store *mockdb.MockStore)
	checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
}

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser()

	testCases := getCreateUserTestCases(user, password)

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.request)
			require.NoError(t, err)

			url := "/users"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestChangeUserPasswordAPI(t *testing.T) {
	user, _ := randomUser()

	testCases := getChangeUserPasswordTestCases(user)

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.request)
			require.NoError(t, err)

			url := "/users/" + tc.request.Username + "/password"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func getCreateUserTestCases(user db.User, password string) []createUserTestCases {
	userRequest := createUserRequest{
		Username: user.Username,
		Password: password,
		FullName: user.FullName,
		Email:    user.Email,
	}

	return []createUserTestCases{
		{
			name:    "ValidRequest",
			request: userRequest,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(
						gomock.Any(),
						EqCreateUserParams(
							db.CreateUserParams{
								Username:       user.Username,
								FullName:       user.FullName,
								Email:          user.Email,
								HashedPassword: user.HashedPassword,
							},
							password,
						),
					).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyMatchUser(t, recorder.Body.String(), user)
			},
		},
		{
			name: "ValidationError",
			request: createUserRequest{
				Username: "invalid-user#",
				Password: "short",
				FullName: "",
				Email:    "invalid-email",
			},
			buildStubs: func(store *mockdb.MockStore) {},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:    "InternalError",
			request: userRequest,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}
}

func getChangeUserPasswordTestCases(user db.User) []changeUserPasswordTestCases {
	userRequest := changeUserPasswordRequest{
		Username: user.Username,
		Password: util.RandomString(6),
	}

	return []changeUserPasswordTestCases{
		{
			name:    "ValidRequest",
			request: userRequest,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUserPassword(gomock.Any(), gomock.Any()).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "ValidationError",
			request: changeUserPasswordRequest{
				Username: user.Username,
				Password: "short",
			},
			buildStubs: func(store *mockdb.MockStore) {},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:    "InternalError",
			request: userRequest,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUserPassword(gomock.Any(), gomock.Any()).
					Times(1).
					Return(user, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}
}

func randomUser() (db.User, string) {
	return db.User{
		Username: util.RandomOwner(),
		FullName: util.RandomOwner(),
		Email:    util.RandomEmail(),
	}, util.RandomString(6)
}

func requireBodyMatchUser(t *testing.T, body string, user db.User) {
	require.Contains(t, body, user.Username)
	require.Contains(t, body, user.FullName)
	require.Contains(t, body, user.Email)
	require.Contains(t, body, user.PasswordChangedAt.String())
	require.Contains(t, body, user.CreatedAt.String())
	require.NotContains(t, body, "\"Password\":")
}
