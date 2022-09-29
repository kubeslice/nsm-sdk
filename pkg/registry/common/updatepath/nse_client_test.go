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

package updatepath_test

import (
	"context"
	"testing"

	"github.com/networkservicemesh/api/pkg/api/registry"
	"github.com/networkservicemesh/sdk/pkg/registry/common/updatepath"
	"github.com/networkservicemesh/sdk/pkg/registry/core/next"
	"github.com/networkservicemesh/sdk/pkg/registry/utils/checks/checknse"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

type nse_client_sample struct {
	name string
	test func(t *testing.T, newUpdatePathServer func(name string) registry.NetworkServiceEndpointRegistryClient)
}

var nse_client_samples = []*nse_client_sample{
	{
		name: "NoPath",
		test: func(t *testing.T, newUpdatePathServer func(name string) registry.NetworkServiceEndpointRegistryClient) {
			t.Cleanup(func() {
				goleak.VerifyNone(t)
			})

			server := newUpdatePathServer(nse1)

			nse, err := server.Register(context.Background(), registerNSERequest(nil))
			require.NoError(t, err)
			require.NotNil(t, nse)

			path := path(0, 1)
			requirePathEqual(t, path, nse.Path, 0)
		},
	},
	{
		name: "SameName",
		test: func(t *testing.T, newUpdatePathServer func(name string) registry.NetworkServiceEndpointRegistryClient) {
			t.Cleanup(func() {
				goleak.VerifyNone(t)
			})

			server := newUpdatePathServer(nse2)

			nse, err := server.Register(context.Background(), registerNSERequest(path(1, 2)))
			require.NoError(t, err)
			require.NotNil(t, nse)

			requirePathEqual(t, path(1, 2), nse.Path)
		},
	},
	{
		name: "DifferentName",
		test: func(t *testing.T, newUpdatePathServer func(name string) registry.NetworkServiceEndpointRegistryClient) {
			t.Cleanup(func() {
				goleak.VerifyNone(t)
			})

			server := newUpdatePathServer(nse3)

			nse, err := server.Register(context.Background(), registerNSERequest(path(1, 2)))
			require.NoError(t, err)
			requirePathEqual(t, path(1, 3), nse.Path, 2)
		},
	},
	{
		name: "InvalidIndex",
		test: func(t *testing.T, newUpdatePathServer func(name string) registry.NetworkServiceEndpointRegistryClient) {
			t.Cleanup(func() {
				goleak.VerifyNone(t)
			})

			server := newUpdatePathServer(nse3)

			_, err := server.Register(context.Background(), registerNSERequest(path(3, 2)))
			require.Error(t, err)
		},
	},
	{
		name: "DifferentNextName",
		test: func(t *testing.T, newUpdatePathServer func(name string) registry.NetworkServiceEndpointRegistryClient) {
			t.Cleanup(func() {
				goleak.VerifyNone(t)
			})

			var nsePath *registry.Path
			server := next.NewNetworkServiceEndpointRegistryClient(
				newUpdatePathServer(nse3),
				checknse.NewClient(t, func(t *testing.T, nse *registry.NetworkServiceEndpoint) {
					nsePath = nse.Path
					requirePathEqual(t, path(2, 3), nsePath, 2)
				}),
			)

			requestPath := path(1, 3)
			requestPath.PathSegments[2].Name = "different"
			nse, err := server.Register(context.Background(), registerNSERequest(requestPath))
			require.NoError(t, err)
			require.NotNil(t, nse)

			nsePath.Index = 1
			requirePathEqual(t, nsePath, nse.Path, 2)
		},
	},
	{
		name: "NoNextAvailable",
		test: func(t *testing.T, newUpdatePathServer func(name string) registry.NetworkServiceEndpointRegistryClient) {
			t.Cleanup(func() {
				goleak.VerifyNone(t)
			})

			var nsePath *registry.Path
			server := next.NewNetworkServiceEndpointRegistryClient(
				newUpdatePathServer(nse3),
				checknse.NewClient(t, func(t *testing.T, nse *registry.NetworkServiceEndpoint) {
					nsePath = nse.Path
					requirePathEqual(t, path(2, 3), nsePath, 2)
				}),
			)

			nse, err := server.Register(context.Background(), registerNSERequest(path(1, 2)))
			require.NoError(t, err)
			require.NotNil(t, nse)

			nsePath.Index = 1
			requirePathEqual(t, nsePath, nse.Path, 2)
		},
	},
	{
		name: "SameNextName",
		test: func(t *testing.T, newUpdatePathServer func(name string) registry.NetworkServiceEndpointRegistryClient) {
			t.Cleanup(func() {
				goleak.VerifyNone(t)
			})

			server := next.NewNetworkServiceEndpointRegistryClient(
				newUpdatePathServer(nse3),
				checknse.NewClient(t, func(t *testing.T, nse *registry.NetworkServiceEndpoint) {
					requirePathEqual(t, path(2, 3), nse.Path)
				}),
			)

			nse, err := server.Register(context.Background(), registerNSERequest(path(1, 3)))
			require.NoError(t, err)
			require.NotNil(t, nse)

			requirePathEqual(t, path(1, 3), nse.Path)
		},
	},
}

func TestUpdatePath(t *testing.T) {
	for i := range nse_client_samples {
		sample := nse_client_samples[i]
		t.Run("TestNetworkServiceEndpointRegistryClient_"+sample.name, func(t *testing.T) {
			sample.test(t, updatepath.NewNetworkServiceEndpointRegistryClient)
		})
	}
}
