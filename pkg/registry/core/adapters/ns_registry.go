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

package adapters

import (
	"context"
	"errors"
	"github.com/networkservicemesh/sdk/pkg/registry/core/next"
	"io"

	"google.golang.org/grpc"

	streamchannel "github.com/networkservicemesh/sdk/pkg/registry/core/streamchannel"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/registry"
)

type networkServiceRegistryServer struct {
	client registry.NetworkServiceRegistryClient
}

func (n *networkServiceRegistryServer) Register(ctx context.Context, request *registry.NetworkService) (*registry.NetworkService, error) {
	doneCtx := withDone(ctx)
	service, err := n.client.Register(doneCtx, request)
	if err != nil {
		return nil, err
	}
	if request == nil {
		request = &registry.NetworkService{}
	}
	if !isDone(doneCtx) {
		return service, nil
	}
	return next.NetworkServiceRegistryServer(ctx).Register(ctx, request)
}

func (n *networkServiceRegistryServer) Find(query *registry.NetworkServiceQuery, s registry.NetworkServiceRegistry_FindServer) error {
	doneCtx := withDone(s.Context())
	client, err := n.client.Find(doneCtx, query)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}
	for {
		if err := client.Context().Err(); err != nil {
			return err
		}
		msg, err := client.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		err = s.Send(msg)
		if err != nil {
			return err
		}
	}
	return next.NetworkServiceRegistryServer(s.Context()).Find(query, s)
}

func (n *networkServiceRegistryServer) Unregister(ctx context.Context, request *registry.NetworkService) (*empty.Empty, error) {
	doneCtx := withDone(ctx)
	service, err := n.client.Unregister(doneCtx, request)
	if err != nil {
		return nil, err
	}
	if request == nil {
		request = &registry.NetworkService{}
	}
	if !isDone(doneCtx) {
		return service, nil
	}
	return next.NetworkServiceRegistryServer(ctx).Unregister(ctx, request)
}

// NetworkServiceClientToServer - returns a registry.NetworkServiceRegistryClient wrapped around the supplied client
func NetworkServiceClientToServer(client registry.NetworkServiceRegistryClient) registry.NetworkServiceRegistryServer {
	return &networkServiceRegistryServer{client: next.NewNetworkServiceRegistryClient(client, &nsDoneClient{})}
}

var _ registry.NetworkServiceRegistryServer = &networkServiceRegistryServer{}

type networkServiceRegistryClient struct {
	server registry.NetworkServiceRegistryServer
}

func (n *networkServiceRegistryClient) Register(ctx context.Context, in *registry.NetworkService, _ ...grpc.CallOption) (*registry.NetworkService, error) {
	doneCtx := withDone(ctx)
	service, err := n.server.Register(doneCtx, in)
	if err != nil {
		return nil, err
	}
	if in == nil {
		in = &registry.NetworkService{}
	}
	if !isDone(doneCtx) {
		return service, nil
	}
	return next.NetworkServiceRegistryClient(ctx).Register(ctx, in)
}

func (n *networkServiceRegistryClient) Find(ctx context.Context, in *registry.NetworkServiceQuery, opts ...grpc.CallOption) (registry.NetworkServiceRegistry_FindClient, error) {
	ch := make(chan *registry.NetworkService, channelSize)
	doneCtx := withDone(ctx)
	s := streamchannel.NewNetworkServiceFindServer(doneCtx, ch)
	if in != nil && in.Watch {
		go func() {
			defer close(ch)
			_ = n.server.Find(in, s)
		}()
	} else {
		defer close(ch)
		if err := n.server.Find(in, s); err != nil {
			return nil, err
		}
	}
	if !isDone(doneCtx) {
		return streamchannel.NewNetworkServiceFindClient(s.Context(), ch), nil
	}
	return next.NetworkServiceRegistryClient(ctx).Find(ctx, in, opts...)
}

func (n *networkServiceRegistryClient) Unregister(ctx context.Context, in *registry.NetworkService, _ ...grpc.CallOption) (*empty.Empty, error) {
	doneCtx := withDone(ctx)
	service, err := n.server.Unregister(doneCtx, in)
	if err != nil {
		return nil, err
	}
	if !isDone(doneCtx) {
		return service, nil
	}
	return next.NetworkServiceRegistryClient(ctx).Unregister(ctx, in)
}

var _ registry.NetworkServiceRegistryClient = &networkServiceRegistryClient{}

// NetworkServiceServerToClient - returns a registry.NetworkServiceRegistryServer wrapped around the supplied server
func NetworkServiceServerToClient(server registry.NetworkServiceRegistryServer) registry.NetworkServiceRegistryClient {
	return &networkServiceRegistryClient{server: next.NewNetworkServiceRegistryServer(server, &nsDoneServer{})}
}
