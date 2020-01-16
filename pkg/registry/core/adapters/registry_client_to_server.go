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

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/networkservicemesh/networkservicemesh/controlplane/api/registry"
)

type registryClientToServer struct {
	client registry.NetworkServiceRegistryClient
}

// NewRegistryClientToServer - returns a registry.NetworkServiceRegistryClient wrapped around the supplied client
func NewRegistryClientToServer(client registry.NetworkServiceRegistryClient) registry.NetworkServiceRegistryServer {
	return &registryClientToServer{client: client}
}

func (r *registryClientToServer) RegisterNSE(ctx context.Context, registration *registry.NSERegistration) (*registry.NSERegistration, error) {
	return r.client.RegisterNSE(ctx, registration)
}

func (r *registryClientToServer) RemoveNSE(ctx context.Context, request *registry.RemoveNSERequest) (*empty.Empty, error) {
	return r.client.RemoveNSE(ctx, request)
}
