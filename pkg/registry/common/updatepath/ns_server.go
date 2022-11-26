// Copyright (c) 2022 Cisco and/or its affiliates.
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

// Package updatepath provides a chain element that sets the id of an incoming or outgoing request
package updatepath

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/networkservicemesh/api/pkg/api/registry"

	"github.com/networkservicemesh/sdk/pkg/registry/common/grpcmetadata"
	"github.com/networkservicemesh/sdk/pkg/registry/core/next"
	"github.com/networkservicemesh/sdk/pkg/tools/spire"
)

type updatePathNSServer struct {
	name string
}

// NewNetworkServiceRegistryServer - creates a new updatePath server to update NetworkService path.
func NewNetworkServiceRegistryServer(name string) registry.NetworkServiceRegistryServer {
	return &updatePathNSServer{
		name: name,
	}
}

func (s *updatePathNSServer) Register(ctx context.Context, ns *registry.NetworkService) (*registry.NetworkService, error) {
	path, err := grpcmetadata.PathFromContext(ctx)
	if err != nil {
		return nil, err
	}

	name := s.name
	if spiffeID, idErr := spire.SpiffeIDFromContext(ctx); idErr == nil {
		name = spiffeID.Path()
	}
	path, index, err := updatePath(path, name)
	if err != nil {
		return nil, err
	}

	ns, err = next.NetworkServiceRegistryServer(ctx).Register(ctx, ns)
	if err != nil {
		return nil, err
	}
	path.Index = index

	return ns, err
}

func (s *updatePathNSServer) Find(query *registry.NetworkServiceQuery, server registry.NetworkServiceRegistry_FindServer) error {
	return next.NetworkServiceRegistryServer(server.Context()).Find(query, server)
}

func (s *updatePathNSServer) Unregister(ctx context.Context, ns *registry.NetworkService) (*empty.Empty, error) {
	path, err := grpcmetadata.PathFromContext(ctx)
	if err != nil {
		return nil, err
	}

	path, index, err := updatePath(path, s.name)
	if err != nil {
		return nil, err
	}

	resp, err := next.NetworkServiceRegistryServer(ctx).Unregister(ctx, ns)
	path.Index = index
	return resp, err
}
