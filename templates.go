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
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	ipc "github.com/jdudmesh/gomon-ipc"
)

const SoftRestartMessage = "__soft_reload"
const HardRestartMessage = "__hard_restart"

type CloseFunc func()

type ReloadManager interface {
	ListenAndServe() error
	Close()
}

type ReloaderFunc func(hint string) error

type reloadManager struct {
	ipcClient  ipc.Connection
	reloaderFn ReloaderFunc
}

func New(reloaderFn ReloaderFunc) (ReloadManager, error) {
	var err error

	serverHost := ipc.DefaultServerHost
	if h, ok := os.LookupEnv("GOMON_IPC_HOST"); ok {
		serverHost = h
	}

	serverPort := ipc.DefaultServerPort
	if p, ok := os.LookupEnv("GOMON_IPC_POST"); ok {
		serverPort, err = strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("unable to parse GOMON_IPC_PORT: %w", err)
		}
	}

	t := &reloadManager{
		reloaderFn: reloaderFn,
	}

	ipcClient, err := ipc.NewConnection(ipc.ClientConnection,
		ipc.WithServerHost(serverHost),
		ipc.WithServerPort(serverPort),
		ipc.WithReadHandler(t.handleInboundMessage))
	if err != nil {
		return nil, fmt.Errorf("unable to start IPC client: %w", err)
	}

	t.ipcClient = ipcClient

	return t, nil
}

func (t *reloadManager) ListenAndServe() error {
	if t == nil {
		return errors.New("reload manager not initialized")
	}

	ctx := context.Background()

	err := t.ipcClient.ListenAndServe(ctx, func(state ipc.ConnectionState) error {
		if state == ipc.Connected {
			return t.sendStartupMessage()
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("unable to start IPC client: %w", err)
	}

	return nil
}

func (t *reloadManager) sendStartupMessage() error {
	ctx, cancelFn := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelFn()

	// send an acknowledgement message
	err := t.ipcClient.Write(ctx, []byte(HardRestartMessage))
	if err != nil {
		return fmt.Errorf("unable to send startup message: %w", err)
	}

	return nil
}

func (t *reloadManager) handleInboundMessage(data []byte) error {
	t.reloaderFn(string(data))

	ctx, cancelFn := context.WithTimeout(context.Background(), time.Second)
	defer cancelFn()

	// send an acknowledgement message
	err := t.ipcClient.Write(ctx, []byte(SoftRestartMessage))
	if err != nil {
		return fmt.Errorf("unable to send reload message: %w", err)
	}

	return nil
}

func (t *reloadManager) Close() {
	if t.ipcClient != nil {
		t.ipcClient.Close()
	}
}
