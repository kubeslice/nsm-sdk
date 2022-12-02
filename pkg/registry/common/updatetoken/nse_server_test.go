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

package updatetoken_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/credentials"

	"github.com/networkservicemesh/sdk/pkg/registry/common/grpcmetadata"
	"github.com/networkservicemesh/sdk/pkg/registry/common/updatepath"
	"github.com/networkservicemesh/sdk/pkg/registry/core/next"
	"github.com/networkservicemesh/sdk/pkg/registry/utils/inject/injectpeertoken"
	"github.com/networkservicemesh/sdk/pkg/tools/token"

	"github.com/networkservicemesh/api/pkg/api/registry"
)

const (
	key      = "supersecret"
	peerID   = "spiffe://test.com/peer"
	proxyID  = "spiffe://test.com/proxy"
	serverID = "spiffe://test.com/server"
)

func tokenGeneratorFunc(spiffeID string) token.GeneratorFunc {
	return func(peerAuthInfo credentials.AuthInfo) (string, time.Time, error) {
		tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": spiffeID}).SignedString([]byte(key))
		return tok, time.Date(3000, 1, 1, 1, 1, 1, 1, time.UTC), err
	}
}

func equalJSON(t require.TestingT, expected, actual interface{}) {
	json1, err1 := json.MarshalIndent(expected, "", "\t")
	require.NoError(t, err1)

	json2, err2 := json.MarshalIndent(actual, "", "\t")
	require.NoError(t, err2)
	require.Equal(t, string(json1), string(json2))
}

func Test_NSEServerLongChain(t *testing.T) {
	peerToken, _, _ := tokenGeneratorFunc(peerID)(nil)
	serverToken, _, _ := tokenGeneratorFunc(serverID)(nil)

	want := &grpcmetadata.Path{
		Index: 0,
		PathSegments: []*grpcmetadata.PathSegment{
			{Token: peerToken},
			{Token: serverToken},
		},
	}

	server := next.NewNetworkServiceEndpointRegistryServer(
		injectpeertoken.NewNetworkServiceEndpointRegistryServer(peerToken),
		updatepath.NewNetworkServiceEndpointRegistryServer(tokenGeneratorFunc(serverID)),
		injectpeertoken.NewNetworkServiceEndpointRegistryServer(serverToken),
		updatepath.NewNetworkServiceEndpointRegistryServer(tokenGeneratorFunc(serverID)),
		injectpeertoken.NewNetworkServiceEndpointRegistryServer(serverToken),
		updatepath.NewNetworkServiceEndpointRegistryServer(tokenGeneratorFunc(serverID)),
	)

	path := &grpcmetadata.Path{}
	ctx := grpcmetadata.PathWithContext(context.Background(), path)
	nse, err := server.Register(ctx, &registry.NetworkServiceEndpoint{})
	require.NoError(t, err)
	require.NotNil(t, nse)

	equalJSON(t, want, path)
}
