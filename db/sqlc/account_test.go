package db

import (
	"context"
	"database/sql"
	"master_class/db/util"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createRandomAccount(t *testing.T) Account {
	user := createRandomUser(t)

	arg := CreateAccountParams{
		Owner:    user.Username,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

	account, err := testQueries.CreateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	return account
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestGetAccount(t *testing.T) {
	generatedAccount := createRandomAccount(t)
	accountFromDb, err := testQueries.GetAccount(context.Background(), generatedAccount.ID)
	require.NoError(t, err)
	require.NotEmpty(t, accountFromDb)

	require.Equal(t, generatedAccount.ID, accountFromDb.ID)
	require.Equal(t, generatedAccount.Owner, accountFromDb.Owner)
	require.Equal(t, generatedAccount.Balance, accountFromDb.Balance)
	require.Equal(t, generatedAccount.Currency, accountFromDb.Currency)
	require.WithinDuration(t, generatedAccount.CreatedAt, accountFromDb.CreatedAt, time.Second)
}

func TestUpdateAccount(t *testing.T) {
	generatedAccount := createRandomAccount(t)

	arg := UpdateAccountBalanceParams{
		ID:      generatedAccount.ID,
		Balance: util.RandomMoney(),
	}

	updatedAccount, err := testQueries.UpdateAccountBalance(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedAccount)

	require.Equal(t, generatedAccount.ID, updatedAccount.ID)
	require.Equal(t, generatedAccount.Owner, updatedAccount.Owner)
	require.Equal(t, arg.Balance, updatedAccount.Balance)
	require.Equal(t, generatedAccount.Currency, updatedAccount.Currency)
	require.WithinDuration(t, generatedAccount.CreatedAt, updatedAccount.CreatedAt, time.Second)
}

func TestDeleteAccount(t *testing.T) {
	generatedAccount := createRandomAccount(t)

	err := testQueries.DeleteAccount(context.Background(), generatedAccount.ID)
	require.NoError(t, err)

	accountFromDb, err := testQueries.GetAccount(context.Background(), generatedAccount.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, accountFromDb)
}

func TestListAccounts(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomAccount(t)
	}

	arg := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}

	accounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, accounts, 5)

	for _, account := range accounts {
		require.NotEmpty(t, account)
	}
}
