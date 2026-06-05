package session_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/session"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

func mustSessionToken(t *testing.T) session.Token {
	t.Helper()
	tok, err := session.NewToken(validSessionToken[:43])
	require.NoError(t, err)
	return tok
}

func TestNewSession(t *testing.T) {
	tok := mustSessionToken(t)
	uid := mustUserID(t)
	createdAt := testTime
	expiresAt := testTime.Add(24 * time.Hour)

	sess, err := session.NewSession(tok, uid, "Mozilla/5.0", "203.0.113.42", createdAt, expiresAt)
	require.NoError(t, err)

	expectedHash := session.HashToken(tok)
	assert.True(t, expectedHash.Equal(sess.TokenHash()), "session.Session.TokenHash() must equal SHA-256 of the plaintext token")
	assert.True(t, uid.Equal(sess.UserID()))
	assert.Equal(t, createdAt, sess.CreatedAt())
	assert.Equal(t, expiresAt, sess.ExpiresAt())
	assert.Equal(t, createdAt, sess.LastSeenAt())
	assert.Equal(t, "Mozilla/5.0", sess.UserAgent())
	assert.Equal(t, "203.0.113.42", sess.IPAddress())
	assert.False(t, sess.IsRevoked())
}

func TestNewSessionRejectsInvalidInputs(t *testing.T) {
	tok := mustSessionToken(t)
	uid := mustUserID(t)
	createdAt := testTime
	expiresAt := testTime.Add(24 * time.Hour)

	cases := []struct {
		name      string
		tok       session.Token
		uid       user.UserID
		createdAt time.Time
		expiresAt time.Time
	}{
		{name: "zero createdAt", tok: tok, uid: uid, createdAt: time.Time{}, expiresAt: expiresAt},
		{name: "zero expiresAt", tok: tok, uid: uid, createdAt: createdAt, expiresAt: time.Time{}},
		{name: "expiresAt before createdAt", tok: tok, uid: uid, createdAt: createdAt, expiresAt: createdAt.Add(-time.Hour)},
		{name: "expiresAt equal createdAt", tok: tok, uid: uid, createdAt: createdAt, expiresAt: createdAt},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := session.NewSession(tc.tok, tc.uid, "ua", "ip", tc.createdAt, tc.expiresAt)
			require.Error(t, err)
			assert.ErrorIs(t, err, session.ErrInvalid)
		})
	}
}

func TestNewSessionRecordsSessionCreatedEvent(t *testing.T) {
	tok := mustSessionToken(t)
	uid := mustUserID(t)
	createdAt := testTime
	expiresAt := testTime.Add(24 * time.Hour)

	sess, _ := session.NewSession(tok, uid, "ua", "ip", createdAt, expiresAt)
	events := sess.Events()
	require.Len(t, events, 1)
	created, ok := events[0].(session.Created)
	require.True(t, ok)
	assert.True(t, created.TokenHash().Equal(sess.TokenHash()))
	assert.True(t, created.UserID().Equal(uid))
	assert.Equal(t, expiresAt, created.ExpiresAt())
	assert.Equal(t, createdAt, created.OccurredAt())
}

func TestReconstituteSession(t *testing.T) {
	tokHash, _ := session.NewTokenHash(validHash)
	uid := mustUserID(t)
	sessionID := session.NewSessionID(uuid.New())
	createdAt := testTime
	expiresAt := testTime.Add(24 * time.Hour)
	lastSeenAt := testTime.Add(time.Hour)

	sess, err := session.ReconstituteSession(sessionID, tokHash, uid, "ua", "ip", createdAt, expiresAt, lastSeenAt, false)
	require.NoError(t, err)
	assert.True(t, sessionID.Equal(sess.ID()))
	assert.True(t, tokHash.Equal(sess.TokenHash()))
	assert.Empty(t, sess.Events())
}

func TestSessionIsExpired(t *testing.T) {
	tok := mustSessionToken(t)
	uid := mustUserID(t)
	createdAt := testTime
	expiresAt := testTime.Add(time.Hour)

	sess, _ := session.NewSession(tok, uid, "ua", "ip", createdAt, expiresAt)
	assert.False(t, sess.IsExpired(createdAt))
	assert.False(t, sess.IsExpired(expiresAt.Add(-time.Minute)))
	assert.True(t, sess.IsExpired(expiresAt), "session is expired at exactly expiresAt")
	assert.True(t, sess.IsExpired(expiresAt.Add(time.Minute)))
}

func TestSessionTouchUpdatesLastSeen(t *testing.T) {
	tok := mustSessionToken(t)
	uid := mustUserID(t)
	createdAt := testTime
	expiresAt := testTime.Add(time.Hour)

	sess, _ := session.NewSession(tok, uid, "ua", "ip", createdAt, expiresAt)
	sess.ClearEvents()

	seenAt := createdAt.Add(10 * time.Minute)
	require.NoError(t, sess.Touch(seenAt))
	assert.Equal(t, seenAt, sess.LastSeenAt())
	assert.Empty(t, sess.Events(), "Touch must not record events")

	require.Error(t, sess.Touch(time.Time{}))
}

func TestSessionRevoke(t *testing.T) {
	tok := mustSessionToken(t)
	uid := mustUserID(t)
	createdAt := testTime
	expiresAt := testTime.Add(time.Hour)

	sess, _ := session.NewSession(tok, uid, "ua", "ip", createdAt, expiresAt)
	sess.ClearEvents()

	seenAt := createdAt.Add(10 * time.Minute)
	require.NoError(t, sess.Revoke(seenAt))
	assert.True(t, sess.IsRevoked())
	assert.True(t, sess.IsExpired(seenAt), "revoked session is expired regardless of expiresAt")

	events := sess.Events()
	require.Len(t, events, 1)
	revoked, ok := events[0].(session.Revoked)
	require.True(t, ok)
	assert.True(t, revoked.TokenHash().Equal(sess.TokenHash()))
	assert.Equal(t, seenAt, revoked.OccurredAt())

	require.Error(t, sess.Revoke(seenAt.Add(time.Hour)), "Revoke must reject already-revoked sessions")
}

func TestReconstituteSessionRejectsInvalidInputs(t *testing.T) {
	tokHash, _ := session.NewTokenHash(validHash)
	uid := mustUserID(t)
	sessionID := session.NewSessionID(uuid.New())
	createdAt := testTime
	expiresAt := testTime.Add(24 * time.Hour)
	lastSeenAt := testTime.Add(time.Hour)

	cases := []struct {
		name       string
		createdAt  time.Time
		expiresAt  time.Time
		lastSeenAt time.Time
	}{
		{name: "zero createdAt", createdAt: time.Time{}, expiresAt: expiresAt, lastSeenAt: lastSeenAt},
		{name: "zero expiresAt", createdAt: createdAt, expiresAt: time.Time{}, lastSeenAt: lastSeenAt},
		{name: "zero lastSeenAt", createdAt: createdAt, expiresAt: expiresAt, lastSeenAt: time.Time{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := session.ReconstituteSession(sessionID, tokHash, uid, "ua", "ip", tc.createdAt, tc.expiresAt, tc.lastSeenAt, false)
			require.Error(t, err)
			assert.ErrorIs(t, err, session.ErrInvalid)
		})
	}
}

func TestSessionRevokeRejectsZeroTimestamp(t *testing.T) {
	tok := mustSessionToken(t)
	uid := mustUserID(t)
	sess, _ := session.NewSession(tok, uid, "ua", "ip", testTime, testTime.Add(time.Hour))

	err := sess.Revoke(time.Time{})
	require.Error(t, err)
	assert.ErrorIs(t, err, session.ErrInvalid)
}

func TestSessionCreatedEventSessionID(t *testing.T) {
	tok := mustSessionToken(t)
	uid := mustUserID(t)
	sess, _ := session.NewSession(tok, uid, "ua", "ip", testTime, testTime.Add(time.Hour))

	events := sess.Events()
	require.Len(t, events, 1)
	created := events[0].(session.Created)
	assert.True(t, sess.ID().Equal(created.SessionID()))
}

func TestSessionRevokedEventSessionID(t *testing.T) {
	tok := mustSessionToken(t)
	uid := mustUserID(t)
	sess, _ := session.NewSession(tok, uid, "ua", "ip", testTime, testTime.Add(time.Hour))
	sess.ClearEvents()

	seenAt := testTime.Add(10 * time.Minute)
	require.NoError(t, sess.Revoke(seenAt))

	events := sess.Events()
	require.Len(t, events, 1)
	revoked := events[0].(session.Revoked)
	assert.True(t, sess.ID().Equal(revoked.SessionID()))
}
