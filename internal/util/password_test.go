package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPassword(t *testing.T) {
	password := "secret123"

	hashedPassword, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	err = CheckPassword(password, hashedPassword)
	require.NoError(t, err)
}

func TestWrongPassword(t *testing.T) {
	password := "secret123"
	wrongPassword := "wrong123"

	hashedPassword, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	err = CheckPassword(wrongPassword, hashedPassword)
	require.Error(t, err)
}

func TestHashPasswordTwice(t *testing.T) {
	password := "secret123"

	hashedPassword1, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword1)

	hashedPassword2, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword2)

	// Bcrypt generates different hashes each time
	require.NotEqual(t, hashedPassword1, hashedPassword2)

	// But both should validate correctly
	err = CheckPassword(password, hashedPassword1)
	require.NoError(t, err)

	err = CheckPassword(password, hashedPassword2)
	require.NoError(t, err)
}
