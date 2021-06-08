// Copyright (c) 2020-2021 Doc.ai and/or its affiliates.
//
// Copyright (c) 2020-2021 Cisco Systems, Inc.
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

package authorize

import (
	"context"
	"sync/atomic"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
)

type authorizeClient struct {
	policies   policiesList
	serverPeer atomic.Value
}

// NewClient - returns a new authorization networkservicemesh.NetworkServiceClient
func NewClient(opts ...Option) networkservice.NetworkServiceClient {
	var result = &authorizeClient{
		policies: defaultPolicies(),
	}
	for _, o := range opts {
		o.apply(&result.policies)
	}
	return result
}

func (a *authorizeClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*networkservice.Connection, error) {
	var p peer.Peer
	opts = append(opts, grpc.Peer(&p))
	resp, err := next.Client(ctx).Request(ctx, request, opts...)
	if err != nil {
		return nil, err
	}
	if p != (peer.Peer{}) {
		a.serverPeer.Store(&p)
		ctx = peer.NewContext(ctx, &p)
	}
	if err = a.policies.check(ctx, resp); err != nil {
		_, _ = next.Client(ctx).Close(ctx, resp, opts...)
		return nil, err
	}
	return resp, err
}

func (a *authorizeClient) Close(ctx context.Context, conn *networkservice.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	p, ok := a.serverPeer.Load().(*peer.Peer)
	if ok && p != nil {
		ctx = peer.NewContext(ctx, p)
	}
	if err := a.policies.check(ctx, conn); err != nil {
		return nil, err
	}
	return next.Client(ctx).Close(ctx, conn, opts...)
}
