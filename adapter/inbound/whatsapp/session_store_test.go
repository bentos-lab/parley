package whatsapp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mau.fi/whatsmeow/types"
)

func TestSessionStorePersistence(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	path, err := sessionPath()
	require.NoError(t, err)
	container, err := newSessionContainer(path)
	require.NoError(t, err)
	device, err := container.GetFirstDevice(context.Background())
	require.NoError(t, err)
	device.ID = &types.JID{User: "12345", Server: types.DefaultUserServer}
	device.LID = types.NewADJID("12345", types.LIDDomain, 1)
	require.NoError(t, device.Save(context.Background()))

	preKeys, err := container.store.GetOrGenPreKeys(context.Background(), 2)
	require.NoError(t, err)
	require.Len(t, preKeys, 2)
	require.NoError(t, container.store.MarkPreKeysAsUploaded(context.Background(), preKeys[len(preKeys)-1].KeyID))
	require.NoError(t, container.persist())

	require.FileExists(t, path)

	reloaded, err := newSessionContainer(path)
	require.NoError(t, err)
	reloadedDevice, err := reloaded.GetFirstDevice(context.Background())
	require.NoError(t, err)
	require.Equal(t, device.GetJID().String(), reloadedDevice.GetJID().String())

	count, err := reloaded.store.UploadedPreKeyCount(context.Background())
	require.NoError(t, err)
	require.Equal(t, 2, count)
}
