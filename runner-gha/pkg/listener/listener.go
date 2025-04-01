package listener

import (
	"context"
	"crypto/cipher"
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"

	"drassi.run/core/util/http"
	"drassi.run/gha-runner/pkg/types"
)

const (
	groupEndpoint    = "_apis/distributedtask/pools"
	sessionEndpoint  = groupEndpoint + "/%d/sessions"
	messagesEndpoint = groupEndpoint + "/%d/messages"
)

type Listener interface {
	CreateSession(ctx context.Context, runnerId, groupId int, key *rsa.PrivateKey) error
	DeleteSession(ctx context.Context) error

	GetMessage(ctx context.Context, os, arch string) (*Message, error)
	DeleteMessage(ctx context.Context, msg *Message) error

	RefreshToken(ctx context.Context) error
}

func NewListener(url string, hc *http.Client) (Listener, error) {
	client, err := xhttp.NewClient(url)
	if err != nil {
		return nil, err
	}

	client = client.WithDefaultErrorHandler(types.ParseActionsError).
		WithDefaultHeader("User-Agent", "gha-runner") // TODO

	if hc != nil {
		client = client.WithHttpClient(hc)
	}

	l := listener{client: client}
	return &l, nil
}

type listener struct {
	client *xhttp.Client

	session       *Session
	eKey          cipher.Block
	lastMessageId int64
}

func (l *listener) CreateSession(ctx context.Context, runnerId, groupId int, key *rsa.PrivateKey) error {
	ss := &Session{
		Runner: &types.RunnerReference{
			Id:                runnerId,
			Version:           "3.0.0",
			GroupId:           groupId,
			Enabled:           true,
			Ephemeral:         false,
			Status:            types.RunnerStatusOnline,
			DisableUpdate:     true,
			ProvisioningState: "Provisioned",
		},
	}
	if hn, err := os.Hostname(); err != nil {
		ss.OwnerName = "RUNNER"
	} else {
		ss.OwnerName = hn
	}

	r := l.client.Post(fmt.Sprintf(sessionEndpoint, groupId)).
		SetQuery("api-version", "5.1-preview").
		WithBodyProvider(xhttp.JsonEncode(ss)).
		OnSuccess(xhttp.JsonDecode(ss))
	if err := r.Do(ctx); err != nil {
		return err
	}

	if block, err := ss.GetKey(key); err != nil {
		return err
	} else {
		l.session, l.eKey = ss, block
	}
	return nil
}

func (l *listener) DeleteSession(ctx context.Context) error {
	if l.session == nil {
		return nil
	}

	endpoint := path.Join(fmt.Sprintf(sessionEndpoint, l.session.Runner.GroupId), l.session.Id)
	return l.client.Delete(endpoint).
		SetQuery("api-version", "5.1-preview").
		Do(ctx)
}

func (l *listener) GetMessage(ctx context.Context, os string, arch string) (*Message, error) {
	runner := l.session.Runner
	r := l.client.Get(fmt.Sprintf(messagesEndpoint, runner.GroupId)).
		SetQuery("api-version", "6.0-preview").
		SetQuery("sessionId", l.session.Id).
		SetQuery("disableUpdate", strconv.FormatBool(runner.DisableUpdate))
	if os != "" {
		r.SetQuery("os", os)
	}
	if arch != "" {
		r.SetQuery("architecture", arch)
	}
	if status := runner.Status; status != "" {
		r.SetQuery("status", string(status))
	}
	if version := runner.Version; version != "" {
		r.SetQuery("runnerVersion", version)
	}

	m := new(Message)
	r.OnSuccess(xhttp.JsonDecode(m))
	if err := r.Do(ctx); err != nil {
		return nil, err
	} else if m.Type == "" {
		return nil, nil
	}

	if body, err := m.DecryptBody(l.eKey); err != nil {
		return nil, err
	} else {
		msg := &Message{Id: m.Id, Type: m.Type, Body: body}
		return msg, nil
	}
}

func (l *listener) DeleteMessage(ctx context.Context, msg *Message) error {
	runner := l.session.Runner
	endpoint := path.Join(fmt.Sprintf(messagesEndpoint, runner.GroupId), strconv.FormatInt(msg.Id, 10))

	return l.client.Delete(endpoint).
		SetQuery("api-version", "6.0-preview").
		SetQuery("sessionId", l.session.Id).
		Do(ctx)
}

func (l *listener) RefreshToken(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}
