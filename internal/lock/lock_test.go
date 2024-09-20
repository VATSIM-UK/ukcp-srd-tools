package lock

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLock(t *testing.T) {
	lock, err := NewLock()
	require.NoError(t, err, "expected no error")
	require.NotNil(t, lock, "expected lock to be non-nil")

	// Clean up by unlocking
	err = lock.Unlock()
	require.NoError(t, err, "expected no error on unlock")
}

func TestUnlock(t *testing.T) {
	lock, err := NewLock()
	require.NoError(t, err, "expected no error")
	require.NotNil(t, lock, "expected lock to be non-nil")

	err = lock.Unlock()
	require.NoError(t, err, "expected no error on unlock")

	// Try unlocking again and expect an error
	err = lock.Unlock()
	require.Error(t, err, "expected error on second unlock")
}

func TestLockFailOnAlreadyLocked(t *testing.T) {
	lock1, err := NewLock()
	require.NoError(t, err, "expected no error")
	require.NotNil(t, lock1, "expected lock to be non-nil")

	lock2, err := NewLock()
	require.Error(t, err, "expected error on second lock")
	require.Nil(t, lock2, "expected lock to be nil")

	// Clean up by unlocking
	err = lock1.Unlock()
	require.NoError(t, err, "expected no error on unlock")

	// If we lock again, we should be able to
	lock3, err := NewLock()
	require.NoError(t, err, "expected no error")

	// Clean up by unlocking
	err = lock3.Unlock()
	require.NoError(t, err, "expected no error on unlock")
}
