package monitor

import (
	"errors"
	"fmt"
	"github.com/no-src/gofs/contract"
	"github.com/no-src/gofs/retry"
	"github.com/no-src/gofs/sync"
	"github.com/no-src/gofs/tran"
	"github.com/no-src/gofs/util"
	"github.com/no-src/log"
	"net/url"
	"strings"
)

type remoteClientMonitor struct {
	syncer   sync.Sync
	retry    retry.Retry
	client   tran.Client
	closed   bool
	messages chan message
	syncOnce bool
}

type message struct {
	data []byte
}

// NewRemoteClientMonitor create an instance of remoteClientMonitor to monitor the remote file change
func NewRemoteClientMonitor(syncer sync.Sync, retry retry.Retry, syncOnce bool, host string, port int, messageQueue int) (m Monitor, err error) {
	if syncer == nil {
		err = errors.New("syncer can't be nil")
		return nil, err
	}
	m = &remoteClientMonitor{
		syncer:   syncer,
		retry:    retry,
		client:   tran.NewClient(host, port),
		messages: make(chan message, messageQueue),
		syncOnce: syncOnce,
	}
	return m, nil
}

func (m *remoteClientMonitor) Start() error {
	if m.client == nil {
		return errors.New("remote sync client is nil")
	}
	err := m.client.Connect()
	if err != nil {
		return err
	}

	// check sync once command
	if m.syncOnce {
		go func() {
			err = m.client.Write(contract.InfoCommand)
			if err != nil {
				log.Error(err, "write info command error")
			}
		}()

		var info contract.FileServerInfo
		m.retry.Do(func() error {
			infoResult, err := m.client.ReadAll()
			if err != nil {
				return err
			}
			err = util.Unmarshal(infoResult, &info)
			if err != nil {
				return err
			}

			if info.ApiType != contract.InfoApi {
				err = errors.New("info command is received other message")
				log.Error(err, "retry to get info response error")
				return err
			}

			return nil
		}, "receive info command response").Wait()

		if info.Code != contract.Success {
			return errors.New("receive info command response error => " + info.Message)
		}

		if info.ApiType != contract.InfoApi {
			return errors.New("info command is received other message")
		}
		return m.syncer.SyncOnce(info.ServerAddr + info.SrcPath)
	}

	go m.processingMessage()
	for {
		if m.closed {
			return errors.New("remote monitor is closed")
		}
		data, err := m.client.ReadAll()
		if err != nil {
			log.Error(err, "client read data error")
			if m.client.IsClosed() {
				log.Debug("try reconnect to server %s:%d", m.client.Host(), m.client.Port())
				m.retry.Do(func() error {
					return m.client.Connect()
				}, fmt.Sprintf("client reconnect to %s:%d", m.client.Host(), m.client.Port()))
			}
		} else {
			m.messages <- message{
				data: data,
			}
		}
	}
	return nil
}

func (m *remoteClientMonitor) processingMessage() {
	for {
		message := <-m.messages
		log.Info("client read request => %s", string(message.data))
		var msg sync.Message
		err := util.Unmarshal(message.data, &msg)
		if err != nil {
			log.Error(err, "client unmarshal data error")
		} else if msg.Code != contract.Success {
			log.Error(errors.New(msg.Message), "remote monitor received the error message")
		} else if msg.ApiType != contract.SyncMessageApi {
			log.Debug("received other message, ignore it => %s", string(message.data))
		} else {
			values := url.Values{}
			values.Add(contract.FsDir, util.String(msg.IsDir))
			values.Add(contract.FsSize, util.String(msg.Size))
			values.Add(contract.FsHash, util.String(msg.Hash))
			values.Add(contract.FsCtime, util.String(msg.CTime))
			values.Add(contract.FsAtime, util.String(msg.ATime))
			values.Add(contract.FsMtime, util.String(msg.MTime))

			// replace question marks with "%3F" to avoid parse the path is breaking when it contains some question marks
			path := msg.BaseUrl + strings.ReplaceAll(msg.Path, "?", "%3F") + fmt.Sprintf("?%s", values.Encode())

			switch msg.Action {
			case sync.CreateAction:
				m.syncer.Create(path)
				break
			case sync.WriteAction:
				m.syncer.Write(path)
				break
			case sync.RemoveAction:
				m.syncer.Remove(path)
				break
			case sync.RenameAction:
				m.syncer.Rename(path)
				break
			case sync.ChmodAction:
				m.syncer.Chmod(path)
				break
			}
		}
	}
}

func (m *remoteClientMonitor) Close() error {
	m.closed = true
	return m.client.Close()
}