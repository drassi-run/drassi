/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package listener

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"

	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ListenerTestSuite struct {
	suite.Suite
	key []byte
	ss  *Session
}

func TestListenerSuite(t *testing.T) {
	suite.Run(t, new(ListenerTestSuite))
}

func (s *ListenerTestSuite) SetupSuite() {
	s.key = randBytes(32)
	s.ss = &Session{
		EncryptionKey: &SessionKey{
			Value: s.key,
		},
	}
}

func (s *ListenerTestSuite) TestSetSession() {
	t := s.T()
	l := new(baseListener)

	err := l.SetSession(s.ss)
	assert.NoError(t, err)
	assert.Equal(t, s.ss, l.Session())
	assert.NotNil(t, l.encKey)
}

func (s *ListenerTestSuite) TestDecryptMessage() {
	t := s.T()

	l := new(baseListener)
	err := l.SetSession(s.ss)
	require.NoError(t, err)

	plainText := []byte("hello world")
	m := mockMessage(encryptMessage(plainText, l.encKey))

	decrypted, err := l.DecryptMessage(m)
	assert.NoError(t, err)
	assert.Equal(t, m.Id, decrypted.Id)
	assert.Equal(t, m.Type, decrypted.Type)
	assert.Equal(t, plainText, decrypted.Body)
}

type BrokerListenerTestSuite struct {
	suite.Suite
	server *httptest.Server
	mux    *http.ServeMux
	key    *rsa.PrivateKey
	l      *brokerListener
}

func TestBrokerListenerSuite(t *testing.T) {
	suite.Run(t, new(BrokerListenerTestSuite))
}

func (s *BrokerListenerTestSuite) SetupTest() {
	var err error
	s.key, err = rsa.GenerateKey(rand.Reader, 2048)
	s.Require().NoError(err)

	s.mux = http.NewServeMux()
	s.server = httptest.NewServer(s.mux)

	l, err := NewBrokerListener(s.server.URL, s.server.Client(), s.key)
	s.Require().NoError(err)
	s.l = l.(*brokerListener)
}

func (s *BrokerListenerTestSuite) TearDownTest() {
	s.server.Close()
}

func (s *BrokerListenerTestSuite) TestConnect_Success() {
	t := s.T()

	wantSessionId := "broker-session-uuid"
	runnerId, groupId := 123, 456
	s.mux.HandleFunc("POST /session", func(w http.ResponseWriter, r *http.Request) {
		var ss Session
		readJsonRequest(t, r, &ss)
		assert.Equal(t, runnerId, ss.Runner.Id)
		assert.Equal(t, groupId, ss.Runner.GroupId)

		ss.Id = wantSessionId
		writeJsonResponse(t, w, ss)
	})

	var done atomic.Bool
	s.mux.HandleFunc("DELETE /session", func(w http.ResponseWriter, r *http.Request) {
		done.Store(true)
		w.WriteHeader(http.StatusOK)
	})

	cancel, err := s.l.Connect(t.Context(), runnerId, groupId)
	require.NoError(t, err)
	assert.NotNil(t, cancel)
	assert.Equal(t, wantSessionId, s.l.Session().Id)
	assert.NoError(t, cancel())
	assert.True(t, done.Load())
}

func (s *BrokerListenerTestSuite) TestConnect_MigrationError() {
	t := s.T()
	s.mux.HandleFunc("POST /session", func(w http.ResponseWriter, r *http.Request) {
		ss := Session{
			BrokerMigrationMessage: &messages.BrokerMigration{
				BaseUrl: "http://new-broker",
			},
		}
		writeJsonResponse(t, w, ss)
	})

	_, err := s.l.Connect(t.Context(), 123, 456)
	assert.ErrorIs(t, err, notSupportMigrationErr)
}
func (s *BrokerListenerTestSuite) TestGetMessage_Success_Plain() {
	ss := mockSession(nil)
	s.testGetMessage_Success(ss)
}

func (s *BrokerListenerTestSuite) TestGetMessage_Success_Encrypted() {
	ss := mockSession(mockSessionKey(nil))
	s.testGetMessage_Success(ss)
}

//goland:noinspection GoSnakeCaseUsage
func (s *BrokerListenerTestSuite) testGetMessage_Success(ss *Session) {
	t := s.T()

	// Mock session
	err := s.l.SetSession(ss)
	require.NoError(t, err)

	plainText := []byte("hello broker")
	var msg *messages.Message
	if s.l.encKey == nil {
		msg = mockMessage(nil, plainText)
	} else {
		msg = mockMessage(encryptMessage(plainText, s.l.encKey))
	}

	s.mux.HandleFunc("GET /message", func(w http.ResponseWriter, r *http.Request) {
		writeJsonResponse(t, w, msg)
	})

	m, err := s.l.GetMessage(t.Context(), "linux", "amd64")
	require.NoError(t, err)
	require.NotNil(t, m)
	assert.Equal(t, msg.Id, m.Id)
	assert.Equal(t, plainText, m.Body)
}

func (s *BrokerListenerTestSuite) TestGetMessage_Empty() {
	t := s.T()

	// Mock session
	err := s.l.SetSession(mockSession(nil))
	require.NoError(t, err)

	s.mux.HandleFunc("GET /message", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	m, err := s.l.GetMessage(t.Context(), "linux", "amd64")
	require.NoError(t, err)
	assert.Nil(t, m)
}

func (s *BrokerListenerTestSuite) TestGetMessage_MigrationError() {
	t := s.T()
	// Mock session
	err := s.l.SetSession(mockSession(nil))
	require.NoError(t, err)

	msg := &messages.Message{
		Type: messages.TypeBrokerMigration,
	}

	s.mux.HandleFunc("GET /message", func(w http.ResponseWriter, r *http.Request) {
		writeJsonResponse(t, w, msg)
	})

	_, err = s.l.GetMessage(t.Context(), "linux", "amd64")
	assert.ErrorIs(t, err, notSupportMigrationErr)
}

func (s *BrokerListenerTestSuite) TestDeleteMessage_Success() {
	t := s.T()
	messageId := int64(123)
	err := s.l.DeleteMessage(t.Context(), &Message{Id: messageId})
	assert.NoError(t, err)
}

type MigratableListenerTestSuite struct {
	suite.Suite

	runnerMux *http.ServeMux
	runner    *httptest.Server

	brokerMux *http.ServeMux
	broker    *httptest.Server

	key *rsa.PrivateKey
	l   *migratableListener
}

func TestMigratableListenerSuite(t *testing.T) {
	suite.Run(t, new(MigratableListenerTestSuite))
}

func (s *MigratableListenerTestSuite) SetupTest() {
	var err error
	s.key, err = rsa.GenerateKey(rand.Reader, 2048)
	s.Require().NoError(err)

	s.runnerMux = http.NewServeMux()
	s.runner = httptest.NewServer(s.runnerMux)

	s.brokerMux = http.NewServeMux()
	s.broker = httptest.NewServer(s.brokerMux)

	l, err := NewMigratableListener(s.runner.URL, s.runner.Client(), s.key)
	s.Require().NoError(err)
	s.l = l.(*migratableListener)
}

func (s *MigratableListenerTestSuite) TearDownTest() {
	s.runner.Close()
	s.broker.Close()
}

func (s *MigratableListenerTestSuite) TestConnect_NoMigration_RunnerService() {
	s.testConnect_NoMigration(func(sessionId string, runnerId, groupId int, done *atomic.Bool) {
		t := s.T()

		s.runnerMux.HandleFunc("POST /_apis/distributedtask/pools/{groupId}/sessions", func(w http.ResponseWriter, r *http.Request) {
			s.Equal(strconv.Itoa(groupId), r.PathValue("groupId"))
			var ss Session
			readJsonRequest(t, r, &ss)
			assert.Equal(t, runnerId, ss.Runner.Id)
			assert.Equal(t, groupId, ss.Runner.GroupId)

			writeJsonResponse(t, w, Session{Id: sessionId})
		})

		s.runnerMux.HandleFunc("DELETE /_apis/distributedtask/pools/{groupId}/sessions/{sessionId}", func(w http.ResponseWriter, r *http.Request) {
			s.Equal(strconv.Itoa(groupId), r.PathValue("groupId"))
			s.Equal(sessionId, r.PathValue("sessionId"))
			done.Store(true)
			w.WriteHeader(http.StatusOK)
		})
	})
}

func (s *MigratableListenerTestSuite) TestConnect_NoMigration_BrokerService() {
	s.testConnect_NoMigration(func(sessionId string, runnerId, groupId int, done *atomic.Bool) {
		t := s.T()
		err := s.l.migrate(s.broker.URL) // Set current Service is Broker
		require.NoError(t, err)

		s.brokerMux.HandleFunc("POST /session", func(w http.ResponseWriter, r *http.Request) {
			var ss Session
			readJsonRequest(t, r, &ss)
			assert.Equal(t, runnerId, ss.Runner.Id)
			assert.Equal(t, groupId, ss.Runner.GroupId)

			ss.Id = sessionId
			writeJsonResponse(t, w, ss)
		})

		s.brokerMux.HandleFunc("DELETE /session", func(w http.ResponseWriter, r *http.Request) {
			done.Store(true)
			w.WriteHeader(http.StatusOK)
		})
	})
}

//goland:noinspection GoSnakeCaseUsage
func (s *MigratableListenerTestSuite) testConnect_NoMigration(setup func(sessionId string, runnerId, groupId int, done *atomic.Bool)) {
	wantSessionId := "broker-session-uuid"
	runnerId, groupId := 123, 456
	done := new(atomic.Bool)

	setup(wantSessionId, runnerId, groupId, done)

	cancel, err := s.l.Connect(s.T().Context(), runnerId, groupId)
	s.NoError(err)
	s.NotNil(cancel)
	s.Equal(wantSessionId, s.l.Session().Id)
	s.NoError(cancel())
	s.True(done.Load())
}

func (s *MigratableListenerTestSuite) TestConnect_WithMigration() {
	t := s.T()
	wantSessionId := "broker-session-uuid"
	runnerId, groupId := 123, 456

	s.runnerMux.HandleFunc("POST /_apis/distributedtask/pools/{groupId}/sessions", func(w http.ResponseWriter, r *http.Request) {
		s.Equal(strconv.Itoa(groupId), r.PathValue("groupId"))
		ss := Session{
			BrokerMigrationMessage: &messages.BrokerMigration{
				BaseUrl: s.broker.URL,
			},
		}
		writeJsonResponse(t, w, ss)
	})

	var brokerConnected atomic.Bool
	s.brokerMux.HandleFunc("POST /session", func(w http.ResponseWriter, r *http.Request) {
		brokerConnected.Store(true)

		var ss Session
		readJsonRequest(t, r, &ss)
		assert.Equal(t, runnerId, ss.Runner.Id)
		assert.Equal(t, groupId, ss.Runner.GroupId)

		ss.Id = wantSessionId
		writeJsonResponse(t, w, ss)
	})

	var done atomic.Bool
	s.brokerMux.HandleFunc("DELETE /session", func(w http.ResponseWriter, r *http.Request) {
		done.Store(true)
		w.WriteHeader(http.StatusOK)
	})

	cancel, err := s.l.Connect(t.Context(), runnerId, groupId)
	s.NoError(err)
	s.NotNil(cancel)
	s.True(brokerConnected.Load())
	s.Equal((service)(s.l.bs), s.l.cur)
	s.Equal(wantSessionId, s.l.Session().Id)
	s.NoError(cancel())
	s.True(done.Load())
}

func (s *MigratableListenerTestSuite) TestGetMessage_NoMigration_RunnerService() {
	s.testGetMessage_NoMigration(func(msg *messages.Message) {
		s.runnerMux.HandleFunc("GET /_apis/distributedtask/pools/{groupId}/messages", func(w http.ResponseWriter, r *http.Request) {
			s.Equal(strconv.Itoa(s.l.Session().Runner.GroupId), r.PathValue("groupId"))
			writeJsonResponse(s.T(), w, msg)
		})
	})
}

func (s *MigratableListenerTestSuite) TestGetMessage_NoMigration_BrokerService() {
	s.testGetMessage_NoMigration(func(msg *messages.Message) {
		t := s.T()
		err := s.l.migrate(s.broker.URL) // Set current Service is Broker
		require.NoError(t, err)

		s.brokerMux.HandleFunc("GET /message", func(w http.ResponseWriter, r *http.Request) {
			s.Equal(s.l.Session().Id, r.URL.Query().Get("sessionId"))
			writeJsonResponse(t, w, msg)
		})
	})
}

//goland:noinspection GoSnakeCaseUsage
func (s *MigratableListenerTestSuite) testGetMessage_NoMigration(setup func(msg *messages.Message)) {
	t := s.T()
	s.setSession()

	plainText := []byte("hello runner")
	msg := mockMessage(nil, plainText)
	setup(msg)

	m, err := s.l.GetMessage(t.Context(), "linux", "amd64")
	s.NoError(err)
	s.NotNil(m)
	s.Equal(msg.Id, m.Id)
	s.Equal(plainText, m.Body)
}

func (s *MigratableListenerTestSuite) TestGetMessage_WithMigration() {
	t := s.T()
	s.setSession()

	// 1. Runner returns migration message
	migrationBody, _ := json.Marshal(map[string]any{"brokerBaseUrl": s.broker.URL})
	m1 := mockMessage(nil, migrationBody)
	m1.Id, m1.Type = 1, messages.TypeBrokerMigration

	s.runnerMux.HandleFunc("GET /_apis/distributedtask/pools/{groupId}/messages", func(w http.ResponseWriter, r *http.Request) {
		s.Equal(strconv.Itoa(s.l.Session().Runner.GroupId), r.PathValue("groupId"))
		writeJsonResponse(t, w, m1)
	})

	// 2. Broker returns real message
	plainText := []byte("hello broker")
	m2 := mockMessage(nil, plainText)
	m2.Id = 2

	s.brokerMux.HandleFunc("GET /message", func(w http.ResponseWriter, r *http.Request) {
		s.Equal(s.l.Session().Id, r.URL.Query().Get("sessionId"))
		writeJsonResponse(t, w, m2)
	})

	m, err := s.l.GetMessage(t.Context(), "linux", "amd64")
	s.NoError(err)
	s.NotNil(m)
	s.Equal(int64(2), m.Id)
	s.Equal(plainText, m.Body)
	s.Equal((service)(s.l.bs), s.l.cur)
}

func (s *MigratableListenerTestSuite) TestDeleteMessage() {
	t := s.T()
	s.setSession()
	msg := &Message{Id: 999}

	s.runnerMux.HandleFunc("DELETE /_apis/distributedtask/pools/{groupId}/messages/{messageId}", func(w http.ResponseWriter, r *http.Request) {
		s.Equal(strconv.Itoa(s.l.Session().Runner.GroupId), r.PathValue("groupId"))
		s.Equal(strconv.FormatInt(msg.Id, 10), r.PathValue("messageId"))
		w.WriteHeader(http.StatusOK)
	})

	err := s.l.DeleteMessage(t.Context(), msg)
	s.NoError(err)
}

// Setup session manually to skip Connect
func (s *MigratableListenerTestSuite) setSession() {
	err := s.l.SetSession(mockSession(nil))
	s.NoError(err)
}

func encryptMessage(plainText []byte, block cipher.Block) (iv, cipherText []byte) {
	// PKCS7 padding
	padLen := aes.BlockSize - (len(plainText) % aes.BlockSize)
	padded := append(plainText, make([]byte, padLen)...)
	for i := len(plainText); i < len(padded); i++ {
		padded[i] = byte(padLen)
	}

	iv = randBytes(16)
	cipherText = make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText, padded)
	return
}

func mockMessage(iv, body []byte) *messages.Message {
	var b string
	if len(iv) > 0 {
		b = base64.StdEncoding.EncodeToString(body)
	} else {
		b = string(body)
	}

	m := &messages.Message{
		Id:   123,
		Type: "test",
		IV:   iv,
		Body: b,
	}
	return m
}

func mockSession(key *SessionKey) *Session {
	ss := &Session{
		Id:            "session-id",
		EncryptionKey: key,
		Runner:        new(types.RunnerReference),
	}
	return ss
}

func mockSessionKey(key *rsa.PrivateKey) *SessionKey {
	k := randBytes(32)

	// TODO: set encrypted=true when key available
	return &SessionKey{
		Encrypted: false,
		Value:     k,
	}
}
