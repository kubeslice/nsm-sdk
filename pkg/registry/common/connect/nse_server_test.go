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

package connect_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/networkservicemesh/api/pkg/api/registry"

	"github.com/networkservicemesh/sdk/pkg/registry/common/connect"
	"github.com/networkservicemesh/sdk/pkg/registry/common/memory"
	"github.com/networkservicemesh/sdk/pkg/registry/common/null"
	"github.com/networkservicemesh/sdk/pkg/registry/core/streamchannel"
	"github.com/networkservicemesh/sdk/pkg/tools/clienturlctx"
	"github.com/networkservicemesh/sdk/pkg/tools/grpcutils"
)

func startNSEServer(ctx context.Context, listenOn *url.URL, server registry.NetworkServiceEndpointRegistryServer) error {
	grpcServer := grpc.NewServer()

	registry.RegisterNetworkServiceEndpointRegistryServer(grpcServer, server)
	grpcutils.RegisterHealthServices(grpcServer, server)

	errCh := grpcutils.ListenAndServe(ctx, listenOn, grpcServer)
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func waitNSEServerStarted(target *url.URL) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cc, err := grpc.DialContext(ctx, grpcutils.URLToTarget(target), grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer func() {
		_ = cc.Close()
	}()

	healthCheckRequest := &grpc_health_v1.HealthCheckRequest{
		Service: registry.ServiceNames(null.NewNetworkServiceEndpointRegistryServer())[0],
	}

	client := grpc_health_v1.NewHealthClient(cc)
	for ctx.Err() == nil {
		response, err := client.Check(ctx, healthCheckRequest)
		if err != nil {
			return err
		}
		if response.Status == grpc_health_v1.HealthCheckResponse_SERVING {
			return nil
		}
	}
	return ctx.Err()
}

func startTestNSEServers(ctx context.Context, t *testing.T) (url1, url2 *url.URL, cancel1, cancel2 context.CancelFunc) {
	var ctx1, ctx2 context.Context

	ctx1, cancel1 = context.WithCancel(ctx)

	url1 = &url.URL{Scheme: "tcp", Host: "127.0.0.1:0"}
	require.NoError(t, startNSEServer(ctx1, url1, memory.NewNetworkServiceEndpointRegistryServer()))

	ctx2, cancel2 = context.WithCancel(ctx)

	url2 = &url.URL{Scheme: "tcp", Host: "127.0.0.1:0"}
	require.NoError(t, startNSEServer(ctx2, url2, memory.NewNetworkServiceEndpointRegistryServer()))

	require.NoError(t, waitNSEServerStarted(url1))
	require.NoError(t, waitNSEServerStarted(url2))

	return url1, url2, cancel1, cancel2
}

func TestConnectNSEServer_AllUnregister(t *testing.T) {
	t.Cleanup(func() { goleak.VerifyNone(t) })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	url1, url2, cancel1, cancel2 := startTestNSEServers(ctx, t)
	defer cancel1()
	defer cancel2()

	ignoreCurrent := goleak.IgnoreCurrent()

	s := connect.NewNetworkServiceEndpointRegistryServer(ctx, connect.WithDialOptions(grpc.WithInsecure()))

	_, err := s.Register(clienturlctx.WithClientURL(context.Background(), url1), &registry.NetworkServiceEndpoint{Name: "nse-1"})
	require.NoError(t, err)

	_, err = s.Register(clienturlctx.WithClientURL(context.Background(), url2), &registry.NetworkServiceEndpoint{Name: "nse-1-1"})
	require.NoError(t, err)

	ch := make(chan *registry.NetworkServiceEndpoint, 1)
	findSrv := streamchannel.NewNetworkServiceEndpointFindServer(clienturlctx.WithClientURL(ctx, url1), ch)
	err = s.Find(&registry.NetworkServiceEndpointQuery{NetworkServiceEndpoint: &registry.NetworkServiceEndpoint{
		Name: "nse-1",
	}}, findSrv)
	require.NoError(t, err)
	require.Equal(t, (<-ch).Name, "nse-1")

	findSrv = streamchannel.NewNetworkServiceEndpointFindServer(clienturlctx.WithClientURL(ctx, url2), ch)
	err = s.Find(&registry.NetworkServiceEndpointQuery{NetworkServiceEndpoint: &registry.NetworkServiceEndpoint{
		Name: "nse-1",
	}}, findSrv)
	require.NoError(t, err)
	require.Equal(t, (<-ch).Name, "nse-1-1")

	_, err = s.Unregister(clienturlctx.WithClientURL(ctx, url1), &registry.NetworkServiceEndpoint{Name: "nse-1"})
	require.NoError(t, err)

	_, err = s.Unregister(clienturlctx.WithClientURL(ctx, url2), &registry.NetworkServiceEndpoint{Name: "nse-1-1"})
	require.NoError(t, err)

	goleak.VerifyNone(t, ignoreCurrent)
}

func TestConnectNSEServer_AllDead_Register(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	url1, url2, cancel1, cancel2 := startTestNSEServers(ctx, t)

	s := connect.NewNetworkServiceEndpointRegistryServer(ctx, connect.WithDialOptions(grpc.WithInsecure()))

	_, err := s.Register(clienturlctx.WithClientURL(ctx, url1), &registry.NetworkServiceEndpoint{Name: "nse-1"})
	require.NoError(t, err)

	_, err = s.Register(clienturlctx.WithClientURL(ctx, url2), &registry.NetworkServiceEndpoint{Name: "nse-1-1"})
	require.NoError(t, err)

	cancel1()
	cancel2()

	var i int
	for err, i = goleak.Find(), 0; err != nil && i < 3; err, i = goleak.Find(), i+1 {
	}
}

func TestConnectNSEServer_AllDead_WatchingFind(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	url1, url2, cancel1, cancel2 := startTestNSEServers(ctx, t)

	s := connect.NewNetworkServiceEndpointRegistryServer(ctx, connect.WithDialOptions(grpc.WithInsecure()))

	go func() {
		ch := make(chan *registry.NetworkServiceEndpoint, 1)
		findSrv := streamchannel.NewNetworkServiceEndpointFindServer(clienturlctx.WithClientURL(ctx, url1), ch)
		err := s.Find(&registry.NetworkServiceEndpointQuery{
			NetworkServiceEndpoint: new(registry.NetworkServiceEndpoint),
			Watch:                  true,
		}, findSrv)
		require.Error(t, err)
	}()

	go func() {
		ch := make(chan *registry.NetworkServiceEndpoint, 1)
		findSrv := streamchannel.NewNetworkServiceEndpointFindServer(clienturlctx.WithClientURL(ctx, url2), ch)
		err := s.Find(&registry.NetworkServiceEndpointQuery{
			NetworkServiceEndpoint: new(registry.NetworkServiceEndpoint),
			Watch:                  true,
		}, findSrv)
		require.Error(t, err)
	}()

	cancel1()
	cancel2()

	for err, i := goleak.Find(), 0; err != nil && i < 3; err, i = goleak.Find(), i+1 {
	}
}
