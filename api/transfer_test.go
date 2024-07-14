package api

import (
	"database/sql"
	"fmt"
	mockdb "master_class/db/mock"
	db "master_class/db/sqlc"
	"master_class/util"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type createTransferTestCases struct {
	name                string
	account_sender_id   int64
	account_receiver_id int64
	amount              int64
	currency            string
	buildStubs          func(store *mockdb.MockStore)
	checkResponse       func(t *testing.T, recorder *httptest.ResponseRecorder)
}

func TestCreateTransferApi(t *testing.T) {
	account_sender := randomAccount(nil)

	currency := account_sender.Currency
	account_receiver := randomAccount(&currency)

	testCases := getCreateTransferTestCases(account_sender, account_receiver)

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := "/transfers"
			body := fmt.Sprintf(`{"from_account_id": %d, "to_account_id": %d, "amount": %d, "currency": "%s"}`, tc.account_sender_id, tc.account_receiver_id, tc.amount, tc.currency)
			request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func getCreateTransferTestCases(account_sender db.Account, account_receiver db.Account) []createTransferTestCases {
	return []createTransferTestCases{
		{
			name:                "OK",
			account_sender_id:   account_sender.ID,
			account_receiver_id: account_receiver.ID,
			amount:              100,
			currency:            account_sender.Currency,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), account_sender.ID).
					Return(account_sender, nil).
					Times(1)
				store.EXPECT().
					GetAccount(gomock.Any(), account_receiver.ID).
					Return(account_receiver, nil).
					Times(1)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(db.TransferTxParams{
						FromAccountID: account_sender.ID,
						ToAccountID:   account_receiver.ID,
						Amount:        100,
					})).
					Return(db.TransferTxResult{}, nil).
					Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:                "FromAccountNotFound",
			account_sender_id:   account_sender.ID,
			account_receiver_id: account_receiver.ID,
			amount:              100,
			currency:            account_sender.Currency,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), account_sender.ID).
					Return(db.Account{}, sql.ErrNoRows).
					Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:                "ToAccountNotFound",
			account_sender_id:   account_sender.ID,
			account_receiver_id: account_receiver.ID,
			amount:              100,
			currency:            account_sender.Currency,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), account_sender.ID).
					Return(account_sender, nil).
					Times(1)
				store.EXPECT().
					GetAccount(gomock.Any(), account_receiver.ID).
					Return(db.Account{}, sql.ErrNoRows).
					Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:                "TransferTxFailed",
			account_sender_id:   account_sender.ID,
			account_receiver_id: account_receiver.ID,
			amount:              100,
			currency:            account_sender.Currency,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), account_sender.ID).
					Return(account_sender, nil).
					Times(1)
				store.EXPECT().
					GetAccount(gomock.Any(), account_receiver.ID).
					Return(account_receiver, nil).
					Times(1)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(db.TransferTxParams{
						FromAccountID: account_sender.ID,
						ToAccountID:   account_receiver.ID,
						Amount:        100,
					})).
					Return(db.TransferTxResult{}, sql.ErrTxDone).
					Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:                "InvalidCurrency",
			account_sender_id:   account_sender.ID,
			account_receiver_id: account_receiver.ID,
			amount:              100,
			currency:            util.PickOtherCurrency(account_sender.Currency),
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), account_sender.ID).
					Return(account_sender, nil).
					Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:                "WrongBalance",
			account_sender_id:   account_sender.ID,
			account_receiver_id: account_receiver.ID,
			amount:              0,
			currency:            account_sender.Currency,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:                "GetAccountFailed",
			account_sender_id:   account_sender.ID,
			account_receiver_id: account_receiver.ID,
			amount:              100,
			currency:            account_sender.Currency,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Return(db.Account{}, sql.ErrConnDone).
					Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}
}
