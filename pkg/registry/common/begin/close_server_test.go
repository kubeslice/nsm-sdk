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

package begin_test

import (
	"context"
	"sync"
	"testing"

	"github.com/networkservicemesh/api/pkg/api/registry"

	"github.com/networkservicemesh/sdk/pkg/registry/common/begin"
	"github.com/networkservicemesh/sdk/pkg/registry/common/null"
	"github.com/networkservicemesh/sdk/pkg/registry/core/adapters"
	"github.com/networkservicemesh/sdk/pkg/registry/core/chain"
	"github.com/networkservicemesh/sdk/pkg/registry/core/next"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestCloseServer(t *testing.T) {
	t.Cleanup(func() { goleak.VerifyNone(t) })
	server := chain.NewNetworkServiceEndpointRegistryServer(
		begin.NewNetworkServiceEndpointRegistryServer(),
		adapters.NetworkServiceEndpointClientToServer(&markClient{t: t}),
	)
	id := "1"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conn, err := server.Register(ctx, &registry.NetworkServiceEndpoint{
		Name: id,
	})
	assert.NotNil(t, t, conn)
	assert.NoError(t, err)
	assert.Equal(t, conn.GetNetworkServiceLabels()[mark].Labels[mark], mark)
	conn = conn.Clone()
	delete(conn.GetNetworkServiceLabels()[mark].Labels, mark)
	assert.Zero(t, conn.GetNetworkServiceLabels()[mark].Labels[mark])
	_, err = server.Unregister(ctx, conn)
	assert.NoError(t, err)
}

func TestDoubleCloseServer(t *testing.T) {
	t.Cleanup(func() { goleak.VerifyNone(t) })
	server := chain.NewNetworkServiceEndpointRegistryServer(
		begin.NewNetworkServiceEndpointRegistryServer(),
		&doubleCloseServer{t: t, NetworkServiceEndpointRegistryServer: null.NewNetworkServiceEndpointRegistryServer()},
	)
	id := "1"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conn, err := server.Register(ctx, &registry.NetworkServiceEndpoint{
		Name: id,
	})
	assert.NotNil(t, t, conn)
	assert.NoError(t, err)
	conn = conn.Clone()
	_, err = server.Unregister(ctx, conn)
	assert.NoError(t, err)
	_, err = server.Unregister(ctx, conn)
	assert.NoError(t, err)
}

type doubleCloseServer struct {
	t *testing.T
	sync.Once
	registry.NetworkServiceEndpointRegistryServer
}

func (s *doubleCloseServer) Unregister(ctx context.Context, in *registry.NetworkServiceEndpoint) (*emptypb.Empty, error) {
	count := 1
	s.Do(func() {
		count++
	})
	assert.Equal(s.t, 2, count, "Close has been called more than once")
	return next.NetworkServiceEndpointRegistryServer(ctx).Unregister(ctx, in)
}
