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

package grpcmetadata

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/networkservicemesh/api/pkg/api/registry"

	"github.com/networkservicemesh/sdk/pkg/registry/core/next"
)

const pathKey = "path"

type grpcMetadataNSServer struct {
}

// NewNetworkServiceRegistryServer - returns grpcmetadata NS server that receives metadata from client and sends it back
func NewNetworkServiceRegistryServer() registry.NetworkServiceRegistryServer {
	return &grpcMetadataNSServer{}
}

func (s *grpcMetadataNSServer) Register(ctx context.Context, ns *registry.NetworkService) (*registry.NetworkService, error) {
	md, loaded := metadata.FromIncomingContext(ctx)
	if !loaded {
		return nil, errors.New("failed to load grpc metadata from context")
	}
	path, err := loadFromMetadata(md)
	if err != nil {
		return nil, err
	}
	ctx = PathWithContext(ctx, path)

	resp, err := next.NetworkServiceRegistryServer(ctx).Register(ctx, ns)
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(path)
	if err != nil {
		return nil, err
	}

	header := metadata.Pairs("path", string(bytes))
	err = grpc.SendHeader(ctx, header)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *grpcMetadataNSServer) Find(query *registry.NetworkServiceQuery, server registry.NetworkServiceRegistry_FindServer) error {
	return next.NetworkServiceRegistryServer(server.Context()).Find(query, server)
}

func (s *grpcMetadataNSServer) Unregister(ctx context.Context, ns *registry.NetworkService) (*empty.Empty, error) {
	md, loaded := metadata.FromIncomingContext(ctx)
	if !loaded {
		return nil, errors.New("failed to load grpc metadata from context")
	}
	path, err := loadFromMetadata(md)
	if err != nil {
		return nil, err
	}
	ctx = PathWithContext(ctx, path)

	resp, err := next.NetworkServiceRegistryServer(ctx).Unregister(ctx, ns)
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(path)
	if err != nil {
		return nil, err
	}

	header := metadata.Pairs("path", string(bytes))
	err = grpc.SendHeader(ctx, header)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
