// Copyright (c) 2021-2023 Cisco and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package heal

import (
	"context"
	"time"

	"go.uber.org/atomic"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/sdk/pkg/networkservice/common/begin"
	"github.com/networkservicemesh/sdk/pkg/tools/log"
)

type eventLoop struct {
	heal         *healClient
	eventLoopCtx context.Context
	chainCtx     context.Context
	conn         *networkservice.Connection
	eventFactory begin.EventFactory
	client       networkservice.MonitorConnection_MonitorConnectionsClient
	reselectFlag atomic.Bool
	logger       log.Logger
}

type checkReselect func() bool

func newEventLoop(ctx context.Context, cc grpc.ClientConnInterface, conn *networkservice.Connection, heal *healClient) (context.CancelFunc, checkReselect, error) {
	conn = conn.Clone()

	ev := begin.FromContext(ctx)
	// Is another chain element asking for events?  If not, no need to monitor
	if ev == nil {
		return func() {}, nil, nil
	}

	// Create new eventLoopCtx and store its eventLoopCancel
	eventLoopCtx, eventLoopCancel := context.WithCancel(ctx)

	// Create selector to only ask for events related to our Connection
	selector := &networkservice.MonitorScopeSelector{
		PathSegments: []*networkservice.PathSegment{
			{
				Id:   conn.GetCurrentPathSegment().GetId(),
				Name: conn.GetCurrentPathSegment().GetName(),
			},
		},
	}

	client, err := networkservice.NewMonitorConnectionClient(cc).MonitorConnections(eventLoopCtx, selector)
	if err != nil {
		eventLoopCancel()
		return nil, nil, errors.Wrap(err, "failed get MonitorConnections client")
	}

	// get the initial state transfer and use it to detect whether we have a real connection or not
	_, err = client.Recv()
	if err != nil {
		eventLoopCancel()
		return nil, nil, errors.Wrap(err, "failed to get the initial state transfer")
	}

	logger := log.FromContext(ctx).WithField("heal", "eventLoop")
	cev := &eventLoop{
		heal:         heal,
		eventLoopCtx: eventLoopCtx,
		chainCtx:     ctx,
		conn:         conn,
		eventFactory: ev,
		client:       newClientFilter(client, conn, logger),
		reselectFlag: atomic.Bool{},
		logger:       logger,
	}

	// Start the eventLoop
	go cev.eventLoop()
	return eventLoopCancel, func() bool { return cev.reselectFlag.Load() }, nil
}

func newDataPlaneEventLoop(ctx context.Context, conn *networkservice.Connection, heal *healClient) (context.CancelFunc, checkReselect, error) {
	conn = conn.Clone()

	ev := begin.FromContext(ctx)
	// Is another chain element asking for events?  If not, no need to monitor
	if ev == nil {
		return func() {}, nil, nil
	}

	if heal.livenessCheck == nil {
		return nil, nil, errors.Errorf("liveness check is missing")
	}

	// Create new eventLoopCtx and store its eventLoopCancel
	eventLoopCtx, eventLoopCancel := context.WithCancel(ctx)

	logger := log.FromContext(ctx).WithField("heal", "eventLoop")
	cev := &eventLoop{
		heal:         heal,
		eventLoopCtx: eventLoopCtx,
		chainCtx:     ctx,
		conn:         conn,
		eventFactory: ev,
		client:       nil,
		reselectFlag: atomic.Bool{},
		logger:       logger,
	}

	// Start the eventLoop
	go cev.eventLoop()
	return eventLoopCancel, func() bool { return cev.reselectFlag.Load() }, nil
}

func (cev *eventLoop) monitorCtrlPlane() <-chan struct{} {
	if cev.client == nil {
		return nil
	}
	res := make(chan struct{}, 1)

	go func() {
		defer close(res)

		for {
			eventIn, err := cev.client.Recv()

			if cev.chainCtx.Err() != nil || cev.eventLoopCtx.Err() != nil {
				res <- struct{}{}
				return
			}

			// Handle error
			if err != nil {
				s, _ := status.FromError(err)
				// This condition means, that the client closed the connection. Stop healing
				if s.Code() == codes.Canceled {
					res <- struct{}{}
				}
				// Otherwise - Start healing
				return
			}

			// Handle event. Start healing
			if eventIn.GetConnections()[cev.conn.GetId()].GetState() == networkservice.State_DOWN {
				return
			}
		}
	}()

	return res
}

func (cev *eventLoop) eventLoop() {
	reselect := false

	ctrlPlaneCh := cev.monitorCtrlPlane()
	dataPlaneCh := cev.monitorDataPlane()

	if dataPlaneCh == nil {
		// If we don't know data plane status, use reselect for safety
		reselect = true
	}

	select {
	case _, ok := <-ctrlPlaneCh:
		if ok {
			// Connection closed
			return
		}
		// Start healing
		cev.logger.Warnf("Control plane is down")
	case _, ok := <-dataPlaneCh:
		if ok {
			// Connection closed
			return
		}
		// Start healing
		cev.logger.Warnf("Data plane is down")
		reselect = true
	case <-cev.chainCtx.Done():
		return
	case <-cev.eventLoopCtx.Done():
		return
	}
	if cev.chainCtx.Err() != nil {
		return
	}

	if reselect {
		cev.reselectFlag.Store(true)
	}

	// We need to force check the DataPlane if a down event was received from the ControlPlane
	if !reselect {
		deadlineCtx, deadlineCancel := context.WithDeadline(cev.chainCtx, time.Now().Add(cev.heal.livenessCheckTimeout))
		if !cev.heal.livenessCheck(deadlineCtx, cev.conn) {
			cev.logger.Warnf("Data plane is down")
			reselect = true
			cev.reselectFlag.Store(true)
		}
		deadlineCancel()
	}

	var options []begin.Option
	if reselect {
		cev.logger.Debugf("Reconnect with reselect")
		options = append(options, begin.WithReselect())
	}
	cev.eventFactory.Request(options...)
}

func (cev *eventLoop) monitorDataPlane() <-chan struct{} {
	if cev.heal.livenessCheck == nil {
		return nil
	}

	res := make(chan struct{}, 1)
	go func() {
		defer close(res)
		deadlineCtx, deadlineCancel := context.WithDeadline(cev.chainCtx, time.Now().Add(cev.heal.livenessCheckTimeout))
		alive := cev.heal.livenessCheck(deadlineCtx, cev.conn)
		deadlineCancel()
		if !alive {
			// Start healing
			return
		}
		ticker := time.NewTicker(cev.heal.livenessCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				deadlineCtx, deadlineCancel := context.WithDeadline(cev.chainCtx, time.Now().Add(cev.heal.livenessCheckTimeout))
				alive := cev.heal.livenessCheck(deadlineCtx, cev.conn)
				deadlineCancel()
				if !alive {
					// Start healing
					return
				}
			case <-cev.eventLoopCtx.Done():
				// EventLoop was canceled. Stop monitoring
				res <- struct{}{}
				return
			}
		}
	}()

	return res
}
