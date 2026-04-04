package whatsapp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mau.fi/whatsmeow/types"
)

func TestRemoveSession(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	path, err := sessionPath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	history := filepath.Join(filepath.Dir(path), "whatsapp.history.json")
	require.NoError(t, os.WriteFile(path, []byte("data"), 0o644))
	require.NoError(t, os.WriteFile(path+".tmp", []byte("tmp"), 0o644))
	require.NoError(t, os.WriteFile(history, []byte("history"), 0o644))
	require.NoError(t, RemoveSession())
	for _, file := range []string{path, path + ".tmp", history} {
		_, err := os.Stat(file)
		require.True(t, os.IsNotExist(err))
	}
}

func TestRemoveSessionMissingVariants(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	path, err := sessionPath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("data"), 0o644))
	require.NoError(t, RemoveSession())
	_, err = os.Stat(path)
	require.True(t, os.IsNotExist(err))
}

func TestGetDeviceMissingSession(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	_, err := getDevice(context.Background())
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestGetDeviceExistingSession(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	path, err := sessionPath()
	require.NoError(t, err)
	container, err := newSessionContainer(path)
	require.NoError(t, err)
	device, err := container.GetFirstDevice(context.Background())
	require.NoError(t, err)
	device.ID = &types.JID{User: "12345", Server: types.DefaultUserServer}
	require.NoError(t, device.Save(context.Background()))

	got, err := getDevice(context.Background())
	require.NoError(t, err)
	require.Equal(t, device.GetJID().String(), got.GetJID().String())
}
