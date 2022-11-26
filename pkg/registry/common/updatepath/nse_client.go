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

package updatepath

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/api/pkg/api/registry"

	"github.com/networkservicemesh/sdk/pkg/registry/common/grpcmetadata"
	"github.com/networkservicemesh/sdk/pkg/registry/core/next"
	"github.com/networkservicemesh/sdk/pkg/tools/spire"
)

type updatePathNSEClient struct {
	name string
}

// NewNetworkServiceEndpointRegistryClient - creates a new updatePath client to update NetworkServiceEndoint path.
func NewNetworkServiceEndpointRegistryClient(name string) registry.NetworkServiceEndpointRegistryClient {
	return &updatePathNSEClient{
		name: name,
	}
}

func (s *updatePathNSEClient) Register(ctx context.Context, nse *registry.NetworkServiceEndpoint, opts ...grpc.CallOption) (*registry.NetworkServiceEndpoint, error) {
	path, err := grpcmetadata.PathFromContext(ctx)
	if err != nil {
		path = &grpcmetadata.Path{}
		ctx = grpcmetadata.PathWithContext(ctx, path)
	}

	name := s.name
	if spiffeID, idErr := spire.SpiffeIDFromContext(ctx); idErr == nil {
		name = spiffeID.Path()
	}
	path, index, err := updatePath(path, name)
	if err != nil {
		return nil, err
	}

	ctx = grpcmetadata.PathWithContext(ctx, path)
	nse, err = next.NetworkServiceEndpointRegistryClient(ctx).Register(ctx, nse, opts...)
	if err != nil {
		return nil, err
	}
	path.Index = index

	return nse, err
}

func (s *updatePathNSEClient) Find(ctx context.Context, query *registry.NetworkServiceEndpointQuery, opts ...grpc.CallOption) (registry.NetworkServiceEndpointRegistry_FindClient, error) {
	return next.NetworkServiceEndpointRegistryClient(ctx).Find(ctx, query, opts...)
}

func (s *updatePathNSEClient) Unregister(ctx context.Context, nse *registry.NetworkServiceEndpoint, opts ...grpc.CallOption) (*empty.Empty, error) {
	path, err := grpcmetadata.PathFromContext(ctx)
	if err != nil {
		path = &grpcmetadata.Path{}
		ctx = grpcmetadata.PathWithContext(ctx, path)
	}

	_, _, err = updatePath(path, s.name)
	if err != nil {
		return nil, err
	}

	return next.NetworkServiceEndpointRegistryClient(ctx).Unregister(ctx, nse, opts...)
}
