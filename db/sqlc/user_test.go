package db

import (
	"context"
	"master_class/db/util"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	arg := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: "secret",
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)

	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	generatedUser := createRandomUser(t)
	userFromDb, err := testQueries.GetUser(context.Background(), generatedUser.Username)
	require.NoError(t, err)
	require.NotEmpty(t, userFromDb)

	require.Equal(t, generatedUser.Username, userFromDb.Username)
	require.Equal(t, generatedUser.HashedPassword, userFromDb.HashedPassword)
	require.Equal(t, generatedUser.FullName, userFromDb.FullName)
	require.Equal(t, generatedUser.Email, userFromDb.Email)
	require.Equal(t, generatedUser.PasswordChangedAt, userFromDb.PasswordChangedAt)
	require.WithinDuration(t, generatedUser.CreatedAt, userFromDb.CreatedAt, time.Second)
}
