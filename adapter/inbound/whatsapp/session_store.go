package whatsapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.mau.fi/util/random"
	"go.mau.fi/whatsmeow/proto/waAdv"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/util/keys"
	waLog "go.mau.fi/whatsmeow/util/log"
)

var errDeviceIDMustBeSet = errors.New("device JID must be known before saving whatsapp session")

type sessionContainer struct {
	path   string
	mu     sync.Mutex
	device *store.Device
	store  *memoryStore
}

func newSessionContainer(path string) (*sessionContainer, error) {
	container := &sessionContainer{path: path}
	container.store = newMemoryStore(container)
	if err := container.load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return container, nil
}

func (c *sessionContainer) GetFirstDevice(ctx context.Context) (*store.Device, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.device != nil {
		return c.device, nil
	}
	if err := c.loadLocked(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if c.device != nil {
		return c.device, nil
	}
	return c.newDeviceLocked(), nil
}

func (c *sessionContainer) PutDevice(ctx context.Context, device *store.Device) error {
	if device.ID == nil {
		return errDeviceIDMustBeSet
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.initializeDevice(device)
	c.device = device
	return c.persistLocked()
}

func (c *sessionContainer) DeleteDevice(ctx context.Context, device *store.Device) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := os.Remove(c.path); err != nil && !os.IsNotExist(err) {
		return err
	}
	c.device = nil
	c.store.reset()
	return nil
}

func (c *sessionContainer) load() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.loadLocked()
}

func (c *sessionContainer) loadLocked() error {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return err
	}
	var persisted persistedSession
	if err := json.Unmarshal(data, &persisted); err != nil {
		return fmt.Errorf("parse whatsapp session: %w", err)
	}
	device, err := buildDeviceFromPersisted(&persisted.Device)
	if err != nil {
		return err
	}
	c.initializeDevice(device)
	c.device = device
	c.store.loadFromPersisted(&persisted, device)
	return nil
}

func (c *sessionContainer) newDeviceLocked() *store.Device {
	device := &store.Device{
		Log:            waLog.Noop,
		Container:      c,
		NoiseKey:       keys.NewKeyPair(),
		IdentityKey:    keys.NewKeyPair(),
		RegistrationID: rand.Uint32(),
		AdvSecretKey:   random.Bytes(32),
	}
	device.SignedPreKey = device.IdentityKey.CreateSignedPreKey(1)
	c.initializeDevice(device)
	return device
}

func (c *sessionContainer) initializeDevice(device *store.Device) {
	if device == nil {
		return
	}
	device.SetAllStores(c.store)
	device.LIDs = c.store
	device.Container = c
	device.Initialized = true
	c.store.setJID(device.ID)
}

func (c *sessionContainer) persist() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.persistLocked()
}

func (c *sessionContainer) persistLocked() error {
	if c.device == nil {
		return nil
	}
	session := persistedSession{
		Device:       persistedDeviceFromStore(c.device),
		PreKeys:      c.store.preKeySnapshot(),
		NextPreKeyID: c.store.nextPreKeyID(),
	}
	tmpPath := c.path + ".tmp"
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return fmt.Errorf("create whatsapp dir: %w", err)
	}
	payload, err := json.MarshalIndent(&session, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal whatsapp session: %w", err)
	}
	if err := os.WriteFile(tmpPath, payload, 0o600); err != nil {
		return fmt.Errorf("write whatsapp session: %w", err)
	}
	return os.Rename(tmpPath, c.path)
}

type persistedSession struct {
	Device       persistedDevice   `json:"device"`
	PreKeys      []persistedPreKey `json:"preKeys"`
	NextPreKeyID uint32            `json:"nextPreKeyId"`
}

type persistedDevice struct {
	ID                    string           `json:"id"`
	LID                   string           `json:"lid"`
	RegistrationID        uint32           `json:"registrationId"`
	NoiseKey              []byte           `json:"noiseKey"`
	IdentityKey           []byte           `json:"identityKey"`
	SignedPreKey          persistedPreKey  `json:"signedPreKey"`
	AdvSecretKey          []byte           `json:"advSecretKey"`
	Account               persistedAccount `json:"account"`
	Platform              string           `json:"platform"`
	BusinessName          string           `json:"businessName"`
	PushName              string           `json:"pushName"`
	FacebookUUID          string           `json:"facebookUuid"`
	LIDMigrationTimestamp int64            `json:"lidMigrationTimestamp"`
}

type persistedAccount struct {
	Details             []byte `json:"details"`
	AccountSignatureKey []byte `json:"accountSignatureKey"`
	AccountSignature    []byte `json:"accountSignature"`
	DeviceSignature     []byte `json:"deviceSignature"`
}

type persistedPreKey struct {
	KeyID     uint32 `json:"keyId"`
	Priv      []byte `json:"priv"`
	Uploaded  bool   `json:"uploaded"`
	Signature []byte `json:"signature,omitempty"`
}

func persistedDeviceFromStore(device *store.Device) persistedDevice {
	var lid string
	if !device.LID.IsEmpty() {
		lid = device.LID.String()
	}
	return persistedDevice{
		ID:                    device.GetJID().String(),
		LID:                   lid,
		RegistrationID:        device.RegistrationID,
		NoiseKey:              append([]byte(nil), device.NoiseKey.Priv[:]...),
		IdentityKey:           append([]byte(nil), device.IdentityKey.Priv[:]...),
		SignedPreKey:          persistedPreKeyFromSigned(device.SignedPreKey),
		AdvSecretKey:          append([]byte(nil), device.AdvSecretKey...),
		Account:               persistedAccountFromDevice(device.Account),
		Platform:              device.Platform,
		BusinessName:          device.BusinessName,
		PushName:              device.PushName,
		FacebookUUID:          device.FacebookUUID.String(),
		LIDMigrationTimestamp: device.LIDMigrationTimestamp,
	}
}

func buildDeviceFromPersisted(persisted *persistedDevice) (*store.Device, error) {
	if persisted == nil || persisted.ID == "" {
		return nil, fmt.Errorf("persisted session missing device id")
	}
	jid, err := types.ParseJID(persisted.ID)
	if err != nil {
		return nil, fmt.Errorf("parse persisted jid: %w", err)
	}
	device := &store.Device{
		Log:            waLog.Noop,
		Container:      nil,
		NoiseKey:       keys.NewKeyPairFromPrivateKey(bytesTo32(persisted.NoiseKey)),
		IdentityKey:    keys.NewKeyPairFromPrivateKey(bytesTo32(persisted.IdentityKey)),
		RegistrationID: persisted.RegistrationID,
		AdvSecretKey:   append([]byte(nil), persisted.AdvSecretKey...),
	}
	device.SignedPreKey = keysNewPreKeyFromPersisted(persisted.SignedPreKey)
	if persisted.LID != "" {
		lid, err := types.ParseJID(persisted.LID)
		if err != nil {
			return nil, fmt.Errorf("parse persisted lid: %w", err)
		}
		device.LID = lid
	}
	if persisted.Account.Details != nil {
		device.Account = &waAdv.ADVSignedDeviceIdentity{
			Details:             append([]byte(nil), persisted.Account.Details...),
			AccountSignatureKey: append([]byte(nil), persisted.Account.AccountSignatureKey...),
			AccountSignature:    append([]byte(nil), persisted.Account.AccountSignature...),
			DeviceSignature:     append([]byte(nil), persisted.Account.DeviceSignature...),
		}
	}
	if persisted.FacebookUUID != "" {
		fbID, err := uuid.Parse(persisted.FacebookUUID)
		if err != nil {
			return nil, fmt.Errorf("parse facebook uuid: %w", err)
		}
		device.FacebookUUID = fbID
	}
	device.Platform = persisted.Platform
	device.BusinessName = persisted.BusinessName
	device.PushName = persisted.PushName
	device.LIDMigrationTimestamp = persisted.LIDMigrationTimestamp
	device.ID = &jid
	return device, nil
}

func persistedPreKeyFromSigned(key *keys.PreKey) persistedPreKey {
	if key == nil {
		return persistedPreKey{}
	}
	signature := make([]byte, 0, 64)
	if key.Signature != nil {
		signature = append(signature, key.Signature[:]...)
	}
	return persistedPreKey{
		KeyID:     key.KeyID,
		Priv:      append([]byte(nil), key.Priv[:]...),
		Uploaded:  true,
		Signature: signature,
	}
}

func keysNewPreKeyFromPersisted(p persistedPreKey) *keys.PreKey {
	key := &keys.PreKey{
		KeyPair: *keys.NewKeyPairFromPrivateKey(bytesTo32(p.Priv)),
		KeyID:   p.KeyID,
	}
	if len(p.Signature) == 64 {
		var sig [64]byte
		copy(sig[:], p.Signature)
		key.Signature = &sig
	}
	return key
}

func persistedAccountFromDevice(account *waAdv.ADVSignedDeviceIdentity) persistedAccount {
	if account == nil {
		return persistedAccount{}
	}
	return persistedAccount{
		Details:             append([]byte(nil), account.Details...),
		AccountSignatureKey: append([]byte(nil), account.AccountSignatureKey...),
		AccountSignature:    append([]byte(nil), account.AccountSignature...),
		DeviceSignature:     append([]byte(nil), account.DeviceSignature...),
	}
}

func bytesTo32(b []byte) [32]byte {
	var arr [32]byte
	copy(arr[:], b)
	return arr
}

type memoryStore struct {
	container *sessionContainer

	mu         sync.RWMutex
	identities map[string][32]byte
	sessions   map[string][]byte

	preKeyLock sync.Mutex
	preKeys    map[uint32]*preKeyEntry
	nextPreKey uint32

	senderKeys map[senderKey][]byte

	appStateSyncKeys    map[string]store.AppStateSyncKey
	appStateVersions    map[string]appStateVersion
	appStateMutationMAC map[string]map[string]mutationRecord

	contacts     map[types.JID]types.ContactInfo
	chatSettings map[types.JID]types.LocalChatSettings

	msgSecrets    map[msgSecretKey][]byte
	privacyTokens map[types.JID]store.PrivacyToken

	bufferedEvents map[string]store.BufferedEvent
	outgoingEvents map[outgoingKey]storedOutgoingEvent

	lidByPN map[types.JID]types.JID
	pnByLID map[types.JID]types.JID

	jid    types.JID
	hasJID bool
}

type preKeyEntry struct {
	key      *keys.PreKey
	uploaded bool
}

type senderKey struct {
	group string
	user  string
}

type appStateVersion struct {
	version uint64
	hash    [128]byte
}

type mutationRecord struct {
	version uint64
	value   []byte
}

type msgSecretKey struct {
	chat    types.JID
	sender  types.JID
	message types.MessageID
}

type outgoingKey struct {
	chat types.JID
	alt  types.JID
	id   types.MessageID
}

type storedOutgoingEvent struct {
	format    string
	payload   []byte
	timestamp time.Time
}

func newMemoryStore(container *sessionContainer) *memoryStore {
	return &memoryStore{
		container:           container,
		identities:          make(map[string][32]byte),
		sessions:            make(map[string][]byte),
		preKeys:             make(map[uint32]*preKeyEntry),
		nextPreKey:          1,
		senderKeys:          make(map[senderKey][]byte),
		appStateSyncKeys:    make(map[string]store.AppStateSyncKey),
		appStateVersions:    make(map[string]appStateVersion),
		appStateMutationMAC: make(map[string]map[string]mutationRecord),
		contacts:            make(map[types.JID]types.ContactInfo),
		chatSettings:        make(map[types.JID]types.LocalChatSettings),
		msgSecrets:          make(map[msgSecretKey][]byte),
		privacyTokens:       make(map[types.JID]store.PrivacyToken),
		bufferedEvents:      make(map[string]store.BufferedEvent),
		outgoingEvents:      make(map[outgoingKey]storedOutgoingEvent),
		lidByPN:             make(map[types.JID]types.JID),
		pnByLID:             make(map[types.JID]types.JID),
	}
}

func (s *memoryStore) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.identities = make(map[string][32]byte)
	s.sessions = make(map[string][]byte)
	s.senderKeys = make(map[senderKey][]byte)
	s.appStateSyncKeys = make(map[string]store.AppStateSyncKey)
	s.appStateVersions = make(map[string]appStateVersion)
	s.appStateMutationMAC = make(map[string]map[string]mutationRecord)
	s.contacts = make(map[types.JID]types.ContactInfo)
	s.chatSettings = make(map[types.JID]types.LocalChatSettings)
	s.msgSecrets = make(map[msgSecretKey][]byte)
	s.privacyTokens = make(map[types.JID]store.PrivacyToken)
	s.bufferedEvents = make(map[string]store.BufferedEvent)
	s.outgoingEvents = make(map[outgoingKey]storedOutgoingEvent)
	s.lidByPN = make(map[types.JID]types.JID)
	s.pnByLID = make(map[types.JID]types.JID)
	if s.preKeys == nil {
		s.preKeys = make(map[uint32]*preKeyEntry)
	}
	s.nextPreKey = 1
}

func (s *memoryStore) setJID(jid *types.JID) {
	if jid == nil {
		s.hasJID = false
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jid = *jid
	s.hasJID = true
}

func (s *memoryStore) preKeySnapshot() []persistedPreKey {
	s.preKeyLock.Lock()
	defer s.preKeyLock.Unlock()
	keys := make([]persistedPreKey, 0, len(s.preKeys))
	ids := make([]uint32, 0, len(s.preKeys))
	for id := range s.preKeys {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		entry := s.preKeys[id]
		if entry == nil {
			continue
		}
		keys = append(keys, persistedPreKey{
			KeyID:    id,
			Priv:     append([]byte(nil), entry.key.Priv[:]...),
			Uploaded: entry.uploaded,
		})
	}
	return keys
}

func (s *memoryStore) nextPreKeyID() uint32 {
	s.preKeyLock.Lock()
	defer s.preKeyLock.Unlock()
	return s.nextPreKey
}

// PutIdentity stores the trusted identity key for the given address.
func (s *memoryStore) PutIdentity(ctx context.Context, address string, key [32]byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.identities[address] = key
	return nil
}

// DeleteAllIdentities removes all stored identities under the provided phone prefix.
func (s *memoryStore) DeleteAllIdentities(ctx context.Context, phone string) error {
	prefix := phone + ":"
	s.mu.Lock()
	defer s.mu.Unlock()
	for addr := range s.identities {
		if strings.HasPrefix(addr, prefix) {
			delete(s.identities, addr)
		}
	}
	return nil
}

// DeleteIdentity removes the stored identity for an individual address.
func (s *memoryStore) DeleteIdentity(ctx context.Context, address string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.identities, address)
	return nil
}

// IsTrustedIdentity determines whether the given identity matches the stored one.
func (s *memoryStore) IsTrustedIdentity(ctx context.Context, address string, key [32]byte) (bool, error) {
	s.mu.RLock()
	stored, ok := s.identities[address]
	s.mu.RUnlock()
	if !ok {
		return true, nil
	}
	return stored == key, nil
}

// GetSession reads a stored session for the given address.
func (s *memoryStore) GetSession(ctx context.Context, address string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.sessions[address]
	if !ok {
		return nil, nil
	}
	return append([]byte(nil), value...), nil
}

// HasSession reports whether a session exists for the given address.
func (s *memoryStore) HasSession(ctx context.Context, address string) (bool, error) {
	s.mu.RLock()
	_, ok := s.sessions[address]
	defer s.mu.RUnlock()
	return ok, nil
}

// GetManySessions returns the sessions for the provided addresses.
func (s *memoryStore) GetManySessions(ctx context.Context, addresses []string) (map[string][]byte, error) {
	result := make(map[string][]byte, len(addresses))
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, addr := range addresses {
		if value, ok := s.sessions[addr]; ok {
			result[addr] = append([]byte(nil), value...)
		}
	}
	return result, nil
}

// PutSession stores the session blob for the given address.
func (s *memoryStore) PutSession(ctx context.Context, address string, session []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[address] = append([]byte(nil), session...)
	return nil
}

// PutManySessions stores session blobs for multiple addresses.
func (s *memoryStore) PutManySessions(ctx context.Context, sessions map[string][]byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for addr, payload := range sessions {
		s.sessions[addr] = append([]byte(nil), payload...)
	}
	return nil
}

// DeleteAllSessions removes all sessions matching the provided phone prefix.
func (s *memoryStore) DeleteAllSessions(ctx context.Context, phone string) error {
	prefix := phone + ":"
	s.mu.Lock()
	defer s.mu.Unlock()
	for addr := range s.sessions {
		if strings.HasPrefix(addr, prefix) {
			delete(s.sessions, addr)
		}
	}
	return nil
}

// DeleteSession removes a single session entry.
func (s *memoryStore) DeleteSession(ctx context.Context, address string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, address)
	return nil
}

// MigratePNToLID renames session keys from a PN to its LID counterpart.
func (s *memoryStore) MigratePNToLID(ctx context.Context, pn, lid types.JID) error {
	pnSignal := pn.SignalAddress().String()
	lidSignal := lid.SignalAddress().String()
	s.mu.Lock()
	defer s.mu.Unlock()
	for addr, data := range s.sessions {
		if strings.HasPrefix(addr, pnSignal) {
			newKey := lidSignal + addr[len(pnSignal):]
			s.sessions[newKey] = data
			delete(s.sessions, addr)
		}
	}
	return nil
}

// GetOrGenPreKeys returns existing unuploaded prekeys and generates more when needed.
func (s *memoryStore) GetOrGenPreKeys(ctx context.Context, count uint32) ([]*keys.PreKey, error) {
	needPersist := false
	result := make([]*keys.PreKey, 0, count)
	s.preKeyLock.Lock()
	ids := make([]uint32, 0, len(s.preKeys))
	for id := range s.preKeys {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		entry := s.preKeys[id]
		if entry == nil || entry.uploaded {
			continue
		}
		result = append(result, entry.key)
		if uint32(len(result)) == count {
			break
		}
	}
	for uint32(len(result)) < count {
		key := keys.NewPreKey(s.nextPreKey)
		s.preKeys[s.nextPreKey] = &preKeyEntry{key: key}
		result = append(result, key)
		s.nextPreKey++
		needPersist = true
	}
	s.preKeyLock.Unlock()
	if needPersist {
		if err := s.container.persist(); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// GenOnePreKey generates exactly one new prekey.
func (s *memoryStore) GenOnePreKey(ctx context.Context) (*keys.PreKey, error) {
	s.preKeyLock.Lock()
	key := keys.NewPreKey(s.nextPreKey)
	s.preKeys[s.nextPreKey] = &preKeyEntry{key: key}
	s.nextPreKey++
	s.preKeyLock.Unlock()
	if err := s.container.persist(); err != nil {
		return nil, err
	}
	return key, nil
}

// GetPreKey retrieves a stored prekey by id.
func (s *memoryStore) GetPreKey(ctx context.Context, id uint32) (*keys.PreKey, error) {
	s.preKeyLock.Lock()
	defer s.preKeyLock.Unlock()
	entry := s.preKeys[id]
	if entry == nil {
		return nil, nil
	}
	return entry.key, nil
}

// RemovePreKey deletes a stored prekey.
func (s *memoryStore) RemovePreKey(ctx context.Context, id uint32) error {
	s.preKeyLock.Lock()
	delete(s.preKeys, id)
	s.preKeyLock.Unlock()
	return s.container.persist()
}

// MarkPreKeysAsUploaded marks all prekeys with id up to upToID as uploaded.
func (s *memoryStore) MarkPreKeysAsUploaded(ctx context.Context, upToID uint32) error {
	changed := false
	s.preKeyLock.Lock()
	for id, entry := range s.preKeys {
		if entry != nil && id <= upToID && !entry.uploaded {
			entry.uploaded = true
			changed = true
		}
	}
	s.preKeyLock.Unlock()
	if changed {
		return s.container.persist()
	}
	return nil
}

// UploadedPreKeyCount returns the number of uploaded prekeys.
func (s *memoryStore) UploadedPreKeyCount(ctx context.Context) (int, error) {
	count := 0
	s.preKeyLock.Lock()
	for _, entry := range s.preKeys {
		if entry != nil && entry.uploaded {
			count++
		}
	}
	s.preKeyLock.Unlock()
	return count, nil
}

// PutSenderKey stores a sender key for the given group and user.
func (s *memoryStore) PutSenderKey(ctx context.Context, group, user string, session []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.senderKeys[senderKey{group: group, user: user}] = append([]byte(nil), session...)
	return nil
}

// GetSenderKey retrieves a previously stored sender key.
func (s *memoryStore) GetSenderKey(ctx context.Context, group, user string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.senderKeys[senderKey{group: group, user: user}]
	if !ok {
		return nil, nil
	}
	return append([]byte(nil), value...), nil
}

// PutAppStateSyncKey stores a sync key indexed by id.
func (s *memoryStore) PutAppStateSyncKey(ctx context.Context, id []byte, key store.AppStateSyncKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	copyID := append([]byte(nil), id...)
	keyCopy := store.AppStateSyncKey{
		Data:        append([]byte(nil), key.Data...),
		Fingerprint: append([]byte(nil), key.Fingerprint...),
		Timestamp:   key.Timestamp,
	}
	s.appStateSyncKeys[string(copyID)] = keyCopy
	return nil
}

// GetAppStateSyncKey retrieves a sync key for the given id.
func (s *memoryStore) GetAppStateSyncKey(ctx context.Context, id []byte) (*store.AppStateSyncKey, error) {
	s.mu.RLock()
	value, ok := s.appStateSyncKeys[string(id)]
	s.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	copy := value
	return &copy, nil
}

// GetLatestAppStateSyncKeyID returns the id of the latest stored sync key.
func (s *memoryStore) GetLatestAppStateSyncKeyID(ctx context.Context) ([]byte, error) {
	var latest []byte
	var latestTS int64
	s.mu.RLock()
	for id, key := range s.appStateSyncKeys {
		if key.Timestamp > latestTS {
			latestTS = key.Timestamp
			latest = append([]byte(nil), []byte(id)...)
		}
	}
	s.mu.RUnlock()
	return latest, nil
}

// GetAllAppStateSyncKeys returns all stored sync keys.
func (s *memoryStore) GetAllAppStateSyncKeys(ctx context.Context) ([]*store.AppStateSyncKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]*store.AppStateSyncKey, 0, len(s.appStateSyncKeys))
	for _, key := range s.appStateSyncKeys {
		copy := key
		keys = append(keys, &copy)
	}
	return keys, nil
}

// PutAppStateVersion stores the version/hash for the given name.
func (s *memoryStore) PutAppStateVersion(ctx context.Context, name string, version uint64, hash [128]byte) error {
	s.mu.Lock()
	s.appStateVersions[name] = appStateVersion{version: version, hash: hash}
	s.mu.Unlock()
	return nil
}

// GetAppStateVersion retrieves the stored version/hash for the given name.
func (s *memoryStore) GetAppStateVersion(ctx context.Context, name string) (uint64, [128]byte, error) {
	s.mu.RLock()
	value, ok := s.appStateVersions[name]
	s.mu.RUnlock()
	if !ok {
		return 0, [128]byte{}, nil
	}
	return value.version, value.hash, nil
}

// DeleteAppStateVersion removes the stored version for the given name.
func (s *memoryStore) DeleteAppStateVersion(ctx context.Context, name string) error {
	s.mu.Lock()
	delete(s.appStateVersions, name)
	s.mu.Unlock()
	return nil
}

// PutAppStateMutationMACs stores mutation MACs for the given name/version.
func (s *memoryStore) PutAppStateMutationMACs(ctx context.Context, name string, version uint64, mutations []store.AppStateMutationMAC) error {
	s.mu.Lock()
	inner, ok := s.appStateMutationMAC[name]
	if !ok {
		inner = make(map[string]mutationRecord)
		s.appStateMutationMAC[name] = inner
	}
	for _, mac := range mutations {
		key := string(mac.IndexMAC)
		inner[key] = mutationRecord{version: version, value: append([]byte(nil), mac.ValueMAC...)}
	}
	s.mu.Unlock()
	return nil
}

// DeleteAppStateMutationMACs removes stored MACs for the provided indexes.
func (s *memoryStore) DeleteAppStateMutationMACs(ctx context.Context, name string, indexMACs [][]byte) error {
	s.mu.Lock()
	inner, ok := s.appStateMutationMAC[name]
	if ok {
		for _, mac := range indexMACs {
			delete(inner, string(mac))
		}
	}
	s.mu.Unlock()
	return nil
}

// GetAppStateMutationMAC returns the most recent MAC value for the given name/index.
func (s *memoryStore) GetAppStateMutationMAC(ctx context.Context, name string, indexMAC []byte) (valueMAC []byte, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	inner, ok := s.appStateMutationMAC[name]
	if !ok {
		return nil, nil
	}
	record, ok := inner[string(indexMAC)]
	if !ok {
		return nil, nil
	}
	return append([]byte(nil), record.value...), nil
}

// PutPushName stores the push name for a user and indicates whether it changed.
func (s *memoryStore) PutPushName(ctx context.Context, user types.JID, pushName string) (bool, string, error) {
	s.mu.Lock()
	entry := s.contacts[user]
	entry.PushName = pushName
	s.contacts[user] = entry
	s.mu.Unlock()
	return true, pushName, nil
}

// PutBusinessName stores the business name for a user and reports whether it changed.
func (s *memoryStore) PutBusinessName(ctx context.Context, user types.JID, businessName string) (bool, string, error) {
	s.mu.Lock()
	entry := s.contacts[user]
	entry.BusinessName = businessName
	s.contacts[user] = entry
	s.mu.Unlock()
	return true, businessName, nil
}

// PutContactName stores the full and first name for a user.
func (s *memoryStore) PutContactName(ctx context.Context, user types.JID, fullName, firstName string) error {
	s.mu.Lock()
	entry := s.contacts[user]
	entry.FullName = fullName
	entry.FirstName = firstName
	s.contacts[user] = entry
	s.mu.Unlock()
	return nil
}

// PutAllContactNames stores multiple contact entries at once.
func (s *memoryStore) PutAllContactNames(ctx context.Context, contacts []store.ContactEntry) error {
	s.mu.Lock()
	for _, contact := range contacts {
		jid := contact.JID
		entry := s.contacts[jid]
		entry.FullName = contact.FullName
		entry.FirstName = contact.FirstName
		s.contacts[jid] = entry
	}
	s.mu.Unlock()
	return nil
}

// PutManyRedactedPhones stores redacted phones for multiple entries.
func (s *memoryStore) PutManyRedactedPhones(ctx context.Context, entries []store.RedactedPhoneEntry) error {
	s.mu.Lock()
	for _, entry := range entries {
		contact := s.contacts[entry.JID]
		contact.RedactedPhone = entry.RedactedPhone
		s.contacts[entry.JID] = contact
	}
	s.mu.Unlock()
	return nil
}

// GetContact returns the stored contact info for a user.
func (s *memoryStore) GetContact(ctx context.Context, user types.JID) (types.ContactInfo, error) {
	s.mu.RLock()
	info := s.contacts[user]
	s.mu.RUnlock()
	return info, nil
}

// GetAllContacts returns a copy of all stored contacts.
func (s *memoryStore) GetAllContacts(ctx context.Context) (map[types.JID]types.ContactInfo, error) {
	copyMap := make(map[types.JID]types.ContactInfo)
	s.mu.RLock()
	for k, v := range s.contacts {
		copyMap[k] = v
	}
	s.mu.RUnlock()
	return copyMap, nil
}

// PutMutedUntil updates the muted-until timestamp for a chat.
func (s *memoryStore) PutMutedUntil(ctx context.Context, chat types.JID, mutedUntil time.Time) error {
	s.mu.Lock()
	entry := s.chatSettings[chat]
	entry.MutedUntil = mutedUntil
	s.chatSettings[chat] = entry
	s.mu.Unlock()
	return nil
}

// PutPinned marks a chat as pinned/unpinned.
func (s *memoryStore) PutPinned(ctx context.Context, chat types.JID, pinned bool) error {
	s.mu.Lock()
	entry := s.chatSettings[chat]
	entry.Pinned = pinned
	s.chatSettings[chat] = entry
	s.mu.Unlock()
	return nil
}

// PutArchived marks a chat as archived/unarchived.
func (s *memoryStore) PutArchived(ctx context.Context, chat types.JID, archived bool) error {
	s.mu.Lock()
	entry := s.chatSettings[chat]
	entry.Archived = archived
	s.chatSettings[chat] = entry
	s.mu.Unlock()
	return nil
}

// GetChatSettings retrieves stored settings for a chat.
func (s *memoryStore) GetChatSettings(ctx context.Context, chat types.JID) (types.LocalChatSettings, error) {
	s.mu.RLock()
	settings := s.chatSettings[chat]
	s.mu.RUnlock()
	return settings, nil
}

// PutMessageSecrets stores multiple message secrets.
func (s *memoryStore) PutMessageSecrets(ctx context.Context, inserts []store.MessageSecretInsert) error {
	s.mu.Lock()
	for _, entry := range inserts {
		key := msgSecretKey{chat: entry.Chat, sender: entry.Sender, message: entry.ID}
		s.msgSecrets[key] = append([]byte(nil), entry.Secret...)
	}
	s.mu.Unlock()
	return nil
}

// PutMessageSecret stores a single message secret.
func (s *memoryStore) PutMessageSecret(ctx context.Context, chat, sender types.JID, id types.MessageID, secret []byte) error {
	s.mu.Lock()
	key := msgSecretKey{chat: chat, sender: sender, message: id}
	s.msgSecrets[key] = append([]byte(nil), secret...)
	s.mu.Unlock()
	return nil
}

// GetMessageSecret retrieves a stored message secret and the original sender JID.
func (s *memoryStore) GetMessageSecret(ctx context.Context, chat, sender types.JID, id types.MessageID) ([]byte, types.JID, error) {
	s.mu.RLock()
	secret, ok := s.msgSecrets[msgSecretKey{chat: chat, sender: sender, message: id}]
	s.mu.RUnlock()
	if !ok {
		return nil, types.EmptyJID, nil
	}
	return append([]byte(nil), secret...), sender, nil
}

// PutPrivacyTokens stores one or more privacy tokens.
func (s *memoryStore) PutPrivacyTokens(ctx context.Context, tokens ...store.PrivacyToken) error {
	s.mu.Lock()
	for _, token := range tokens {
		key := token.User
		s.privacyTokens[key] = token
	}
	s.mu.Unlock()
	return nil
}

// GetPrivacyToken retrieves a stored privacy token for a user.
func (s *memoryStore) GetPrivacyToken(ctx context.Context, user types.JID) (*store.PrivacyToken, error) {
	s.mu.RLock()
	value, ok := s.privacyTokens[user]
	s.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	copy := value
	return &copy, nil
}

// GetBufferedEvent returns a buffered event for the given hash.
func (s *memoryStore) GetBufferedEvent(ctx context.Context, ciphertextHash [32]byte) (*store.BufferedEvent, error) {
	key := string(ciphertextHash[:])
	s.mu.RLock()
	value, ok := s.bufferedEvents[key]
	s.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	copy := value
	return &copy, nil
}

// PutBufferedEvent stores a buffered event payload.
func (s *memoryStore) PutBufferedEvent(ctx context.Context, ciphertextHash [32]byte, plaintext []byte, serverTimestamp time.Time) error {
	key := string(ciphertextHash[:])
	s.mu.Lock()
	s.bufferedEvents[key] = store.BufferedEvent{
		Plaintext:  append([]byte(nil), plaintext...),
		InsertTime: time.Now(),
		ServerTime: serverTimestamp,
	}
	s.mu.Unlock()
	return nil
}

// DoDecryptionTxn executes the provided function while holding the lock.
func (s *memoryStore) DoDecryptionTxn(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// ClearBufferedEventPlaintext removes the plaintext while preserving the record.
func (s *memoryStore) ClearBufferedEventPlaintext(ctx context.Context, ciphertextHash [32]byte) error {
	key := string(ciphertextHash[:])
	s.mu.Lock()
	if entry, ok := s.bufferedEvents[key]; ok {
		entry.Plaintext = nil
		s.bufferedEvents[key] = entry
	}
	s.mu.Unlock()
	return nil
}

// DeleteOldBufferedHashes removes buffered events older than the provided timestamp.
func (s *memoryStore) DeleteOldBufferedHashes(ctx context.Context) error {
	threshold := time.Now().Add(-24 * time.Hour)
	s.mu.Lock()
	for key, entry := range s.bufferedEvents {
		if entry.InsertTime.Before(threshold) {
			delete(s.bufferedEvents, key)
		}
	}
	s.mu.Unlock()
	return nil
}

// GetOutgoingEvent retrieves a stored outgoing event.
func (s *memoryStore) GetOutgoingEvent(ctx context.Context, chatJID, altChatJID types.JID, id types.MessageID) (format string, result []byte, err error) {
	key := outgoingKey{chat: chatJID, id: id}
	s.mu.RLock()
	entry, ok := s.outgoingEvents[key]
	if !ok && !altChatJID.IsEmpty() {
		entry, ok = s.outgoingEvents[outgoingKey{chat: altChatJID, id: id}]
	}
	s.mu.RUnlock()
	if !ok {
		return "", nil, nil
	}
	return entry.format, append([]byte(nil), entry.payload...), nil
}

// AddOutgoingEvent stores an outgoing event payload for retry.
func (s *memoryStore) AddOutgoingEvent(ctx context.Context, chatJID types.JID, id types.MessageID, format string, plaintext []byte) error {
	key := outgoingKey{chat: chatJID, id: id}
	s.mu.Lock()
	s.outgoingEvents[key] = storedOutgoingEvent{
		format:    format,
		payload:   append([]byte(nil), plaintext...),
		timestamp: time.Now(),
	}
	s.mu.Unlock()
	return nil
}

// DeleteOldOutgoingEvents removes retry entries older than the given time.
func (s *memoryStore) DeleteOldOutgoingEvents(ctx context.Context) error {
	threshold := time.Now().Add(-24 * time.Hour)
	s.mu.Lock()
	for key, entry := range s.outgoingEvents {
		if entry.timestamp.Before(threshold) {
			delete(s.outgoingEvents, key)
		}
	}
	s.mu.Unlock()
	return nil
}

// PutManyLIDMappings stores multiple LID/PN relations.
func (s *memoryStore) PutManyLIDMappings(ctx context.Context, mappings []store.LIDMapping) error {
	s.mu.Lock()
	for _, mapping := range mappings {
		s.lidByPN[mapping.PN] = mapping.LID
		s.pnByLID[mapping.LID] = mapping.PN
	}
	s.mu.Unlock()
	return nil
}

// PutLIDMapping stores a single PN/LID mapping.
func (s *memoryStore) PutLIDMapping(ctx context.Context, lid, jid types.JID) error {
	s.mu.Lock()
	s.lidByPN[jid] = lid
	s.pnByLID[lid] = jid
	s.mu.Unlock()
	return nil
}

// GetPNForLID retrieves the PN associated with the provided LID.
func (s *memoryStore) GetPNForLID(ctx context.Context, lid types.JID) (types.JID, error) {
	s.mu.RLock()
	value := s.pnByLID[lid]
	s.mu.RUnlock()
	return value, nil
}

// GetLIDForPN retrieves the LID associated with the provided PN.
func (s *memoryStore) GetLIDForPN(ctx context.Context, pn types.JID) (types.JID, error) {
	s.mu.RLock()
	value := s.lidByPN[pn]
	s.mu.RUnlock()
	return value, nil
}

// GetManyLIDsForPNs returns a map of PN to LID for the provided PNs.
func (s *memoryStore) GetManyLIDsForPNs(ctx context.Context, pns []types.JID) (map[types.JID]types.JID, error) {
	result := make(map[types.JID]types.JID, len(pns))
	s.mu.RLock()
	for _, pn := range pns {
		result[pn] = s.lidByPN[pn]
	}
	s.mu.RUnlock()
	return result, nil
}

func (s *memoryStore) loadFromPersisted(session *persistedSession, device *store.Device) {
	s.setJID(device.ID)
	s.preKeyLock.Lock()
	s.preKeys = make(map[uint32]*preKeyEntry, len(session.PreKeys))
	for _, entry := range session.PreKeys {
		key := keysNewPreKeyFromPersisted(entry)
		s.preKeys[entry.KeyID] = &preKeyEntry{key: key, uploaded: entry.Uploaded}
	}
	if session.NextPreKeyID > 0 {
		s.nextPreKey = session.NextPreKeyID
	}
	s.preKeyLock.Unlock()
}
