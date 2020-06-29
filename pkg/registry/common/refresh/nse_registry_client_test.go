// Copyright (c) 2020 Doc.ai and/or its affiliates.
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

package refresh_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/registry"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/sdk/pkg/registry/common/refresh"
)

const testExpirationDuration = time.Millisecond * 100

type testNSEClient struct {
	sync.Mutex
	requestCount int
}

func (t *testNSEClient) Register(ctx context.Context, in *registry.NetworkServiceEndpoint, opts ...grpc.CallOption) (*registry.NetworkServiceEndpoint, error) {
	t.Lock()
	defer t.Unlock()
	t.requestCount++
	return in, nil
}

func (t *testNSEClient) Find(ctx context.Context, in *registry.NetworkServiceEndpointQuery, opts ...grpc.CallOption) (registry.NetworkServiceEndpointRegistry_FindClient, error) {
	panic("implement me")
}

func (t *testNSEClient) Unregister(ctx context.Context, in *registry.NetworkServiceEndpoint, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, nil
}

func TestNewNetworkServiceEndpointRegistryClient(t *testing.T) {
	defer goleak.VerifyNone(t)
	testClient := testNSEClient{}
	refreshClient := refresh.NewNetworkServiceEndpointRegistryClient(&testClient,
		refresh.WithRetryPeriod(time.Millisecond*100),
		refresh.WithDefaultExpiration(testExpirationDuration),
	)
	_, err := refreshClient.Register(context.Background(), &registry.NetworkServiceEndpoint{
		Name: "nse-1",
	})
	require.Nil(t, err)
	require.Eventually(t, func() bool {
		testClient.Lock()
		defer testClient.Unlock()
		return testClient.requestCount == 1
	}, testExpirationDuration*2, testExpirationDuration/4)
	_, err = refreshClient.Unregister(context.Background(), &registry.NetworkServiceEndpoint{Name: "nse-1"})
	require.Nil(t, err)
}

func TestNewNetworkServiceEndpointRegistryClient_CalledRegisterTwice(t *testing.T) {
	defer goleak.VerifyNone(t)
	testClient := testNSEClient{}
	refreshClient := refresh.NewNetworkServiceEndpointRegistryClient(&testClient,
		refresh.WithRetryPeriod(time.Millisecond*100),
		refresh.WithDefaultExpiration(time.Millisecond*100),
	)
	_, err := refreshClient.Register(context.Background(), &registry.NetworkServiceEndpoint{
		Name: "nse-1",
	})
	require.Nil(t, err)
	_, err = refreshClient.Register(context.Background(), &registry.NetworkServiceEndpoint{
		Name: "nse-1",
	})
	require.Nil(t, err)
	require.Eventually(t, func() bool {
		testClient.Lock()
		defer testClient.Unlock()
		return testClient.requestCount == 1
	}, testExpirationDuration*2, testExpirationDuration/4)
	_, err = refreshClient.Unregister(context.Background(), &registry.NetworkServiceEndpoint{Name: "nse-1"})
	require.Nil(t, err)
}
