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

// Package setid provides NSE client chain element for setting nse.Name
package setid

import (
	"context"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/networkservicemesh/api/pkg/api/registry"

	"github.com/networkservicemesh/sdk/pkg/registry/core/next"
)

type setIDClient struct {
	names namesSet
}

// NewNetworkServiceEndpointRegistryClient creates a new NSE client chain element generating unique nse.Name on Register
func NewNetworkServiceEndpointRegistryClient() registry.NetworkServiceEndpointRegistryClient {
	return new(setIDClient)
}

func (c *setIDClient) Register(ctx context.Context, nse *registry.NetworkServiceEndpoint, opts ...grpc.CallOption) (reg *registry.NetworkServiceEndpoint, err error) {
	if _, ok := c.names.Load(nse.Name); ok {
		return next.NetworkServiceEndpointRegistryClient(ctx).Register(ctx, nse, opts...)
	}

	nameSuffix := nse.Name
	if nameSuffix == "" {
		nameSuffix = strings.Join(nse.NetworkServiceNames, "-")
	}
	nameSuffix = "-" + nameSuffix

	for err = status.Error(codes.AlreadyExists, ""); err != nil && isAlreadyExistsError(err); {
		name := uuid.New().String() + nameSuffix

		nse.Name = name
		if reg, err = next.NetworkServiceEndpointRegistryClient(ctx).Register(ctx, nse, opts...); err == nil {
			c.names.Store(name, struct{}{})
		}
	}

	return reg, err
}

func isAlreadyExistsError(e error) bool {
	grpcStatus, ok := status.FromError(e)
	return ok && grpcStatus.Code() == codes.AlreadyExists
}

func (c *setIDClient) Find(ctx context.Context, query *registry.NetworkServiceEndpointQuery, opts ...grpc.CallOption) (registry.NetworkServiceEndpointRegistry_FindClient, error) {
	return next.NetworkServiceEndpointRegistryClient(ctx).Find(ctx, query, opts...)
}

func (c *setIDClient) Unregister(ctx context.Context, nse *registry.NetworkServiceEndpoint, opts ...grpc.CallOption) (*empty.Empty, error) {
	if _, ok := c.names.LoadAndDelete(nse.Name); !ok {
		return new(empty.Empty), nil
	}
	return next.NetworkServiceEndpointRegistryClient(ctx).Unregister(ctx, nse, opts...)
}
