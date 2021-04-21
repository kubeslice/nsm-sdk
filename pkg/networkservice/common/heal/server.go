// Copyright (c) 2020 Cisco Systems, Inc.
//
// Copyright (c) 2021 Doc.ai and/or its affiliates.
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

// Package heal provides a chain element that carries out proper nsm healing from client to endpoint
package heal

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/sdk/pkg/networkservice/common/discover"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/adapters"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
	"github.com/networkservicemesh/sdk/pkg/tools/addressof"
	"github.com/networkservicemesh/sdk/pkg/tools/log"
)

type healServer struct {
	ctx            context.Context
	onHeal         *networkservice.NetworkServiceClient
	healContextMap ctxWrapperMap
}

// NewServer - creates a new networkservice.NetworkServiceServer chain element that implements the healing algorithm
//             - ctx    - context for the lifecycle of the *Server* itself. Cancel when discarding the server.
//             - onHeal - client used 'onHeal'.
//                        If we detect we need to heal, onHeal.Request is used to heal.
//                        If we can't heal connection, onHeal.Close will be called.
//                        If onHeal is nil, then we simply set onHeal to this client chain element
//                        Since networkservice.NetworkServiceClient is an interface (and thus a pointer)
//                        *networkservice.NetworkServiceClient is a double pointer.  Meaning it
//                        points to a place that points to a place that implements networkservice.NetworkServiceClient
//                        This is done because when we use heal.NewClient as part of a chain, we may not *have*
//                        a pointer to this
func NewServer(ctx context.Context, onHeal *networkservice.NetworkServiceClient) networkservice.NetworkServiceServer {
	rv := &healServer{
		ctx:    ctx,
		onHeal: onHeal,
	}

	if rv.onHeal == nil {
		rv.onHeal = addressof.NetworkServiceClient(adapters.NewServerToClient(rv))
	}

	return rv
}

func (f *healServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	ctx = withRequestHealFunc(ctx, f.requestHeal)
	conn, err := next.Server(ctx).Request(ctx, request)
	if err != nil {
		return nil, err
	}

	f.healContextMap.applyLockedOrNew(request.GetConnection().GetId(), func(created bool, cw *ctxWrapper) {
		if !created {
			if cw.cancel != nil {
				cw.cancel()
				cw.cancel = nil
			}
		}
		cw.request = request.Clone().SetRequestConnection(conn.Clone())
		cw.ctx = f.createHealContext(ctx, cw.ctx)
	})

	return conn, nil
}

func (f *healServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	f.stopHeal(conn)

	rv, err := next.Server(ctx).Close(ctx, conn)
	if err != nil {
		return nil, err
	}
	return rv, nil
}

// requestHeal - heals requested connection. Returns immediately, heal is asynchronous.
func (f *healServer) requestHeal(conn *networkservice.Connection, restoreConnection bool) {
	var healCtx context.Context
	var request *networkservice.NetworkServiceRequest
	f.healContextMap.applyLocked(conn.GetId(), func(cw *ctxWrapper) {
		if cw.cancel != nil {
			cw.cancel()
		}
		ctx, cancel := context.WithCancel(cw.ctx)
		cw.cancel = cancel
		healCtx = ctx
		request = cw.request
	})

	if request == nil {
		return
	}

	request.SetRequestConnection(conn.Clone())

	if restoreConnection {
		go f.restoreConnection(healCtx, request)
	} else {
		go f.processHeal(healCtx, request)
	}
}

func (f *healServer) stopHeal(conn *networkservice.Connection) {
	cw, loaded := f.healContextMap.LoadAndDelete(conn.GetId())
	if !loaded {
		return
	}
	cw.mut.Lock()
	defer cw.mut.Unlock()
	if cw.cancel != nil {
		cw.cancel()
	}
}

func (f *healServer) restoreConnection(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) {
	if ctx.Err() != nil {
		return
	}

	// Make sure we have a valid expireTime to work with
	expires := request.GetConnection().GetNextPathSegment().GetExpires()
	expireTime, err := ptypes.Timestamp(expires)
	if err != nil {
		return
	}

	deadline := time.Now().Add(time.Minute)
	if deadline.After(expireTime) {
		deadline = expireTime
	}
	requestCtx, requestCancel := context.WithDeadline(ctx, deadline)
	defer requestCancel()

	for requestCtx.Err() == nil {
		if _, err = (*f.onHeal).Request(requestCtx, request.Clone(), opts...); err == nil {
			return
		}
	}

	f.processHeal(ctx, request.Clone(), opts...)
}

func (f *healServer) processHeal(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) {
	logEntry := log.FromContext(ctx).WithField("healServer", "processHeal")
	conn := request.GetConnection()

	candidates := discover.Candidates(ctx)
	if candidates != nil || conn.GetPath().GetIndex() == 0 {
		logEntry.Infof("Starting heal process for %s", conn.GetId())

		healCtx, healCancel := context.WithCancel(ctx)
		defer healCancel()

		reRequest := request.Clone()
		reRequest.GetConnection().NetworkServiceEndpointName = ""
		path := reRequest.GetConnection().Path
		reRequest.GetConnection().Path.PathSegments = path.PathSegments[0 : path.Index+1]

		for ctx.Err() == nil {
			_, err := (*f.onHeal).Request(healCtx, reRequest, opts...)
			if err != nil {
				logEntry.Errorf("Failed to heal connection %s: %v", conn.GetId(), err)
			} else {
				logEntry.Infof("Finished heal process for %s", conn.GetId())
				break
			}
		}
	} else {
		// Huge timeout is not required to close connection on a current path segment
		closeCtx, closeCancel := context.WithTimeout(ctx, time.Second)
		defer closeCancel()

		_, err := (*f.onHeal).Close(closeCtx, request.GetConnection().Clone(), opts...)
		if err != nil {
			logEntry.Errorf("Failed to close connection %s: %v", request.GetConnection().GetId(), err)
		}
	}
}

// createHealContext - create context to be used on heal.
//                     Uses f.ctx as base and inserts Candidates from requestCtx or cachedCtx into it, if there are any.
func (f *healServer) createHealContext(requestCtx, cachedCtx context.Context) context.Context {
	ctx := requestCtx
	if cachedCtx != nil {
		if candidates := discover.Candidates(ctx); candidates == nil || len(candidates.Endpoints) > 0 {
			ctx = cachedCtx
		}
	}
	healCtx := f.ctx
	if candidates := discover.Candidates(ctx); candidates != nil {
		healCtx = discover.WithCandidates(healCtx, candidates.Endpoints, candidates.NetworkService)
	}

	return healCtx
}

// applyLocked - searches the map for entry with key id and runs provided function while entry mutex is locked.
//               If map doesn't contain this key, does nothing.
func (m *ctxWrapperMap) applyLocked(id string, fun func(cw *ctxWrapper)) {
	cw, ok := m.Load(id)
	if !ok {
		return
	}
	cw.mut.Lock()
	defer cw.mut.Unlock()
	fun(cw)
}

// applyLocked - searches the map for entry with key id and runs provided function while entry mutex is locked.
//               If map doesn't contain this creates new key.
//               Function is informed whether it receives new key or already created key via first argument.
func (m *ctxWrapperMap) applyLockedOrNew(id string, fun func(created bool, cw *ctxWrapper)) {
	newCw := &ctxWrapper{}
	// we need to lock this before storing in map to prevent potential race with other threads that can read this map
	newCw.mut.Lock()
	defer newCw.mut.Unlock()
	cw, loaded := m.LoadOrStore(id, newCw)
	if loaded {
		cw.mut.Lock()
		defer cw.mut.Unlock()
	}
	fun(!loaded, cw)
}
