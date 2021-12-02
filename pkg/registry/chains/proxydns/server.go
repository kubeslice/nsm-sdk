// Copyright (c) 2020-2021 Doc.ai and/or its affiliates.
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

// Package proxydns provides default chain for stateless proxy registries based on DNS
package proxydns

import (
	"context"

	"google.golang.org/grpc"

	"github.com/networkservicemesh/sdk/pkg/registry"
	"github.com/networkservicemesh/sdk/pkg/registry/common/connect"
	"github.com/networkservicemesh/sdk/pkg/registry/common/dnsresolve"
	"github.com/networkservicemesh/sdk/pkg/registry/core/chain"
)

// NewServer creates new stateless registry server that proxies queries to the second registries by DNS domains
func NewServer(ctx context.Context, dnsResolver dnsresolve.Resolver, dialOptions ...grpc.DialOption) registry.Registry {
	nseChain := chain.NewNetworkServiceEndpointRegistryServer(
		dnsresolve.NewNetworkServiceEndpointRegistryServer(dnsresolve.WithResolver(dnsResolver)),
		connect.NewNetworkServiceEndpointRegistryServer(ctx, connect.WithDialOptions(dialOptions...)))
	nsChain := chain.NewNetworkServiceRegistryServer(
		dnsresolve.NewNetworkServiceRegistryServer(dnsresolve.WithResolver(dnsResolver)),
		connect.NewNetworkServiceRegistryServer(ctx, connect.WithDialOptions(dialOptions...)))
	return registry.NewServer(nsChain, nseChain)
}
