package client

// gomon is a simple command line tool that watches your files and automatically restarts the application when it detects any changes in the working directory.
// Copyright (C) 2023 John Dudmesh

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

import (
	"errors"
	"os"
	"time"

	ipc "github.com/james-barrow/golang-ipc"
)

const MsgTypeReload = 1
const MsgTypeReloaded = 2
const MsgTypePing = 99

type CloseFunc func()

type TemplateManager interface {
	Run() error
	Close() error
}

type Reloader interface {
	Reload(string)
}

type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type templateManager struct {
	ipcChannel string
	ipcClient  *ipc.Client
	reloader   Reloader
	logger     Logger
}

func newManager(reloader Reloader, logger Logger) (*templateManager, error) {
	ipcChannel, ok := os.LookupEnv("GOMON_IPC_CHANNEL")
	if !ok {
		return nil, errors.New("GOMON_IPC_CHANNEL not set")
	}

	t := &templateManager{
		ipcChannel,
		nil,
		reloader,
		logger,
	}

	return t, nil
}

func (t *templateManager) Run() error {
	var err error

	t.ipcClient, err = ipc.StartClient(t.ipcChannel, nil)
	if err != nil {
		t.logger.Errorf("Unable to start IPC client: %w", err)
		return err
	}

	go func() {
		for {
			msg, err := t.ipcClient.Read()
			if err != nil {
				t.logger.Errorf("Unable to receive message: %v", err)
				break
			}

			switch msg.MsgType {
			case MsgTypeReload:
				data := string(msg.Data)
				t.logger.Infof("Reload notification: %s", data)
				t.reloader.Reload(data)
				err = t.ipcClient.Write(MsgTypeReloaded, msg.Data)
				if err != nil {
					t.logger.Errorf("Unable to send message: %v", err)
					break
				}
			case MsgTypePing:
				t.logger.Infof("Ping received")
			case -1:
				t.logger.Infof("Internal message received: %+v", msg)
			default:
				t.logger.Errorf("Unknown message: %v", msg)
			}
		}
	}()

	time.Sleep(250 * time.Millisecond)
	err = t.ipcClient.Write(MsgTypePing, nil)
	if err != nil {
		t.logger.Errorf("Unable to send startup message: %v", err)
	}

	return nil
}

func (t *templateManager) Close() error {
	t.ipcClient.Close()
	return nil
}