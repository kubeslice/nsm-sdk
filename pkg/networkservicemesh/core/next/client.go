// Copyright (c) 2020 Cisco Systems, Inc.
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

package next

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/networkservicemesh/controlplane/api/connection"
	"github.com/networkservicemesh/networkservicemesh/controlplane/api/networkservice"
)

type nextClient struct {
	clients []networkservice.NetworkServiceClient
	index   int
}

// ClientWrapper - a function that wraps around a networkservice.NetworkServiceClient
type ClientWrapper func(networkservice.NetworkServiceClient) networkservice.NetworkServiceClient

// ClientChainer - a function that chains together a list of networkservice.NetworkServiceClients
type ClientChainer func(...networkservice.NetworkServiceClient) networkservice.NetworkServiceClient

// NewWrappedNetworkServiceClient chains together clients with wrapper wrapped around each one
func NewWrappedNetworkServiceClient(wrapper ClientWrapper, clients ...networkservice.NetworkServiceClient) networkservice.NetworkServiceClient {
	rv := &nextClient{
		clients: clients,
	}
	for i := range rv.clients {
		rv.clients[i] = wrapper(rv.clients[i])
	}
	return rv
}

// NewNetworkServiceClient - chains together clients into a single networkservice.NetworkServiceServer
func NewNetworkServiceClient(clients ...networkservice.NetworkServiceClient) networkservice.NetworkServiceClient {
	return NewWrappedNetworkServiceClient(nil, clients...)
}

func (n *nextClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*connection.Connection, error) {
	if n.index+1 < len(n.clients) {
		return n.clients[n.index].Request(withNextClient(ctx, &nextClient{clients: n.clients, index: n.index + 1}), request, opts...)
	}
	return n.clients[n.index].Request(withNextClient(ctx, newTailClient()), request, opts...)
}

func (n *nextClient) Close(ctx context.Context, conn *connection.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	if n.index+1 < len(n.clients) {
		return n.clients[n.index].Close(withNextClient(ctx, &nextClient{clients: n.clients, index: n.index + 1}), conn, opts...)
	}
	return n.clients[n.index].Close(withNextClient(ctx, newTailClient()), conn, opts...)
}
