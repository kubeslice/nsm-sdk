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

// Package next provides a mechanism for chained registry.{Registry,Discovery}{Server,Client}s to call
// the next element in the chain.
package next

import (
	"context"

	"github.com/networkservicemesh/sdk/pkg/registry/core/adapters"

	"github.com/networkservicemesh/api/pkg/api/registry"
)

const (
	nextRegistryServerKey contextKeyType = "NextRegistryServer"
	nextRegistryClientKey contextKeyType = "NextRegistryClient"
)

// withNextRegistryServer -
//    Wraps 'parent' in a new Context that has the DiscoveryServer registry.NetworkServiceRegistryServer to be called in the chain
//    Should only be set in CompositeEndpoint.Request/Close
func withNextRegistryServer(parent context.Context, next registry.NetworkServiceRegistryServer) context.Context {
	if parent == nil {
		parent = context.TODO()
	}
	return context.WithValue(parent, nextRegistryServerKey, next)
}

// RegistryServer -
//   Returns the RegistryServer registry.NetworkServiceRegistryServer to be called in the chain from the context.Context
func RegistryServer(ctx context.Context) registry.NetworkServiceRegistryServer {
	rv, ok := ctx.Value(nextRegistryServerKey).(registry.NetworkServiceRegistryServer)
	if !ok {
		client, ok := ctx.Value(nextRegistryClientKey).(registry.NetworkServiceRegistryClient)
		if ok {
			rv = adapters.NewRegistryClientToServer(client)
		}
	}
	if rv != nil {
		return rv
	}
	return &tailRegistryServer{}
}

// withNextRegistryClient -
//    Wraps 'parent' in a new Context that has the RegistryClient registry.NetworkServiceRegistryClient to be called in the chain
//    Should only be set in CompositeEndpoint.Request/Close
func withNextRegistryClient(parent context.Context, next registry.NetworkServiceRegistryClient) context.Context {
	if parent == nil {
		parent = context.TODO()
	}
	return context.WithValue(parent, nextRegistryClientKey, next)
}

// RegistryClient -
//   Returns the RegistryClient registry.NetworkServiceRegistryClient to be called in the chain from the context.Context
func RegistryClient(ctx context.Context) registry.NetworkServiceRegistryClient {
	rv, ok := ctx.Value(nextRegistryClientKey).(registry.NetworkServiceRegistryClient)
	if !ok {
		server, ok := ctx.Value(nextRegistryServerKey).(registry.NetworkServiceRegistryServer)
		if ok {
			rv = adapters.NewRegistryServerToClient(server)
		}
	}
	if rv != nil {
		return rv
	}
	return &tailRegistryClient{}
}
