package db

import (
	"context"
	"database/sql"
	"master_class/db/util"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createRandomTransfer(t *testing.T, from, to Account) Transfer {
	args := CreateTransferParams{
		FromAccountID: sql.NullInt64{Int64: from.ID, Valid: true},
		ToAccountID:   sql.NullInt64{Int64: to.ID, Valid: true},
		Amount:        util.RandomMoney(),
	}
	transfer, err := testQueries.CreateTransfer(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, args.FromAccountID, transfer.FromAccountID)
	require.Equal(t, args.ToAccountID, transfer.ToAccountID)
	require.Equal(t, args.Amount, transfer.Amount)
	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)

	return transfer
}

func TestCreateTransfer(t *testing.T) {
	from := createRandomAccount(t)
	to := createRandomAccount(t)
	createRandomTransfer(t, from, to)
}

func TestGetTransfer(t *testing.T) {
	transfer1 := createRandomTransfer(t, createRandomAccount(t), createRandomAccount(t))
	transfer2, err := testQueries.GetTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transfer2)

	require.Equal(t, transfer1.ID, transfer2.ID)
	require.Equal(t, transfer1.FromAccountID, transfer2.FromAccountID)
	require.Equal(t, transfer1.ToAccountID, transfer2.ToAccountID)
	require.Equal(t, transfer1.Amount, transfer2.Amount)
	require.WithinDuration(t, transfer1.CreatedAt, transfer2.CreatedAt, time.Second)
}

func TestListTransfers(t *testing.T) {
	from := createRandomAccount(t)
	to := createRandomAccount(t)

	for i := 0; i < 10; i++ {
		createRandomTransfer(t, from, to)
	}

	args := ListTransfersParams{
		Limit:  5,
		Offset: 5,
	}

	transfers, err := testQueries.ListTransfers(context.Background(), args)
	require.NoError(t, err)
	require.Len(t, transfers, 5)
}

func TestUpdateTransfer(t *testing.T) {
	transfer := createRandomTransfer(t, createRandomAccount(t), createRandomAccount(t))

	args := UpdateTransferParams{
		ID:     transfer.ID,
		Amount: util.RandomMoney(),
	}

	updatedTransfer, err := testQueries.UpdateTransfer(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, updatedTransfer)

	require.Equal(t, transfer.ID, updatedTransfer.ID)
	require.Equal(t, args.Amount, updatedTransfer.Amount)
	require.WithinDuration(t, transfer.CreatedAt, updatedTransfer.CreatedAt, time.Second)
}

func TestDeleteTransfer(t *testing.T) {
	transfer := createRandomTransfer(t, createRandomAccount(t), createRandomAccount(t))

	err := testQueries.DeleteTransfer(context.Background(), transfer.ID)
	require.NoError(t, err)

	transferFromDb, err := testQueries.GetTransfer(context.Background(), transfer.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, transferFromDb)
}
