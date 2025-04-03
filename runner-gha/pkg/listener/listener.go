package listener

import (
	"context"
	"crypto/cipher"
	"crypto/rsa"
	"errors"
	"net/http"

	"drassi.run/core/util/http"
	"drassi.run/gha-runner/pkg/message"
	"drassi.run/gha-runner/pkg/types"
)

const maxMigrations = 3

var (
	tooManyMigrationErr    = errors.New("too many migrations")
	notSupportMigrationErr = errors.New("migration is not supported")
)

func newClient(url string, hc *http.Client) (*xhttp.Client, error) {
	client, err := xhttp.NewClient(url)
	if err != nil {
		return nil, err
	}

	client = client.WithDefaultErrorHandler(types.ParseActionsError).
		WithDefaultHeader("User-Agent", "gha-runner") // TODO

	if hc != nil {
		client = client.WithHttpClient(hc)
	}
	return client, nil
}

type Listener interface {
	Connect(ctx context.Context, runnerId, groupId int) (func() error, error)

	GetMessage(ctx context.Context, os, arch string) (*message.Message, error)
	DeleteMessage(ctx context.Context, msg *message.Message) error

	RefreshToken(ctx context.Context) error
}

type baseListener struct {
	privKey *rsa.PrivateKey
	encKey  cipher.Block

	session *Session
}

func (l *baseListener) SetSession(ss *Session) error {
	if block, err := ss.GetKey(l.privKey); err != nil {
		return err
	} else {
		l.session, l.encKey = ss, block
		return nil
	}
}

func (l *baseListener) Session() *Session {
	return l.session
}

func (l *baseListener) DecryptMessage(msg *message.Message) (*message.Message, error) {
	if body, err := msg.DecryptBody(l.encKey); err != nil {
		return nil, err
	} else {
		m := &message.Message{Id: msg.Id, Type: msg.Type, Body: body}
		return m, nil
	}
}

// NewMigratableListener create [Listener] that first interact with Runner server,
// then migrate to Broker server if received a [messages.BrokerMigration] message
//
//   - https://github.com/actions/runner/blob/v2.323.0/src/Runner.Listener/MessageListener.cs#L40
func NewMigratableListener(url string, hc *http.Client, key *rsa.PrivateKey) (Listener, error) {
	client, err := newClient(url, hc)
	if err != nil {
		return nil, err
	}

	l := migratableListener{
		baseListener: baseListener{
			privKey: key,
		},

		rs: &runnerService{client: client},
		bs: &brokerService{client: client},
	}
	l.cur = l.rs
	return &l, nil
}

type migratableListener struct {
	baseListener

	rs  *runnerService
	bs  *brokerService
	cur service
}

func (l *migratableListener) migrate(url string) error {
	if err := l.bs.SetUrl(url); err != nil {
		return err
	}
	l.cur = l.bs
	return nil
}

func (l *migratableListener) Connect(ctx context.Context, runnerId, groupId int) (func() error, error) {
	for try := 0; try < maxMigrations; try++ {
		ref := &types.RunnerReference{
			Id:                runnerId,
			Version:           "3.0.0",
			GroupId:           groupId,
			Enabled:           true,
			Ephemeral:         false,
			Status:            types.RunnerStatusOnline,
			DisableUpdate:     true,
			ProvisioningState: "Provisioned",
		}

		session, cancel, err := l.cur.Connect(ctx, ref)
		if err != nil {
			return nil, err
		}

		if mm := session.BrokerMigrationMessage; mm != nil {
			if err = l.migrate(mm.BaseUrl); err != nil {
				return nil, err
			}
			continue
		}

		if err = l.SetSession(session); err != nil {
			return nil, err
		}
		return cancel, nil
	}

	return nil, tooManyMigrationErr
}

func (l *migratableListener) GetMessage(ctx context.Context, os string, arch string) (*message.Message, error) {
	for try := 0; try < maxMigrations; try++ {
		msg, err := l.cur.GetMessage(ctx, l.Session(), os, arch)
		if err != nil {
			return nil, err
		}

		msg, err = l.DecryptMessage(msg)
		if err != nil {
			return nil, err
		}

		if msg.Type != message.TypeBrokerMigration {
			return msg, nil
		}

		if msg, err := message.Decode[message.BrokerMigration](msg.Body); err != nil {
			return nil, err
		} else if err = l.migrate(msg.BaseUrl); err != nil {
			return nil, err
		}
	}
	return nil, tooManyMigrationErr
}

func (l *migratableListener) DeleteMessage(ctx context.Context, msg *message.Message) error {
	return l.cur.DeleteMessage(ctx, l.Session(), msg.Id)
}

func (l *migratableListener) RefreshToken(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

// NewBrokerListener create [Listener] that interact with Broker server for flow v2.
// BrokerListener cannot handle migration, and will return an error when receiving a messages.BrokerMigration message.
//
//   - https://github.com/actions/runner/blob/v2.323.0/src/Runner.Listener/BrokerMessageListener.cs#L21
func NewBrokerListener(url string, hc *http.Client, key *rsa.PrivateKey) (Listener, error) {
	client, err := newClient(url, hc)
	if err != nil {
		return nil, err
	}

	l := &brokerListener{
		baseListener: baseListener{
			privKey: key,
		},

		svc: &brokerService{
			client: client,
		},
	}
	return l, nil
}

type brokerListener struct {
	baseListener

	svc *brokerService
}

func (l *brokerListener) Connect(ctx context.Context, runnerId, groupId int) (func() error, error) {
	ref := &types.RunnerReference{
		Id:                runnerId,
		Version:           "3.0.0",
		GroupId:           groupId,
		Enabled:           true,
		Ephemeral:         false,
		Status:            types.RunnerStatusOnline,
		DisableUpdate:     true,
		ProvisioningState: "Provisioned",
	}

	session, cancel, err := l.svc.Connect(ctx, ref)
	if err != nil {
		return nil, err
	}

	if mm := session.BrokerMigrationMessage; mm != nil {
		return nil, notSupportMigrationErr
	}

	if err = l.SetSession(session); err != nil {
		return nil, err
	}

	return cancel, nil
}

func (l *brokerListener) GetMessage(ctx context.Context, os, arch string) (*message.Message, error) {
	msg, err := l.svc.GetMessage(ctx, l.Session(), os, arch)
	if err != nil {
		return nil, err
	}

	if msg.Type == message.TypeBrokerMigration {
		return nil, notSupportMigrationErr
	}

	return l.DecryptMessage(msg)
}

func (l *brokerListener) DeleteMessage(ctx context.Context, msg *message.Message) error {
	return l.svc.DeleteMessage(ctx, l.Session(), msg.Id)
}

func (l *brokerListener) RefreshToken(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}
