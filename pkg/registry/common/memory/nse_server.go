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

package memory

import (
	"context"
	"io"

	"github.com/edwarnicke/serialize"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"

	"github.com/networkservicemesh/api/pkg/api/registry"

	"github.com/networkservicemesh/sdk/pkg/registry/core/next"
	"github.com/networkservicemesh/sdk/pkg/tools/matchutils"
)

type memoryNSEServer struct {
	networkServiceEndpoints NetworkServiceEndpointSyncMap
	executor                serialize.Executor
	eventChannels           map[string]chan *registry.NetworkServiceEndpoint
	eventChannelSize        int
}

// NewNetworkServiceEndpointRegistryServer creates new memory based NetworkServiceEndpointRegistryServer
func NewNetworkServiceEndpointRegistryServer(options ...Option) registry.NetworkServiceEndpointRegistryServer {
	s := &memoryNSEServer{
		eventChannelSize: defaultEventChannelSize,
		eventChannels:    make(map[string]chan *registry.NetworkServiceEndpoint),
	}
	for _, o := range options {
		o.apply(s)
	}
	return s
}

func (s *memoryNSEServer) setEventChannelSize(l int) {
	s.eventChannelSize = l
}

func (s *memoryNSEServer) Register(ctx context.Context, nse *registry.NetworkServiceEndpoint) (*registry.NetworkServiceEndpoint, error) {
	r, err := next.NetworkServiceEndpointRegistryServer(ctx).Register(ctx, nse)
	if err != nil {
		return nil, err
	}

	s.networkServiceEndpoints.Store(r.Name, r.Clone())

	s.sendEvent(r)

	return r, err
}

func (s *memoryNSEServer) sendEvent(event *registry.NetworkServiceEndpoint) {
	event = event.Clone()
	s.executor.AsyncExec(func() {
		for _, ch := range s.eventChannels {
			ch <- event.Clone()
		}
	})
}

func (s *memoryNSEServer) Find(query *registry.NetworkServiceEndpointQuery, server registry.NetworkServiceEndpointRegistry_FindServer) error {
	if !query.Watch {
		for _, ns := range s.allMatches(query) {
			if err := server.Send(ns); err != nil {
				return err
			}
		}
		return next.NetworkServiceEndpointRegistryServer(server.Context()).Find(query, server)
	}

	eventCh := make(chan *registry.NetworkServiceEndpoint, s.eventChannelSize)
	id := uuid.New().String()

	s.executor.AsyncExec(func() {
		s.eventChannels[id] = eventCh
		for _, entity := range s.allMatches(query) {
			eventCh <- entity
		}
	})
	defer s.closeEventChannel(id, eventCh)

	var err error
	for ; err == nil; err = s.receiveEvent(query, server, eventCh) {
	}
	if err != io.EOF {
		return err
	}
	return next.NetworkServiceEndpointRegistryServer(server.Context()).Find(query, server)
}

func (s *memoryNSEServer) allMatches(query *registry.NetworkServiceEndpointQuery) (matches []*registry.NetworkServiceEndpoint) {
	s.networkServiceEndpoints.Range(func(_ string, nse *registry.NetworkServiceEndpoint) bool {
		if matchutils.MatchNetworkServiceEndpoints(query.NetworkServiceEndpoint, nse) {
			matches = append(matches, nse.Clone())
		}
		return true
	})
	return matches
}

func (s *memoryNSEServer) closeEventChannel(id string, eventCh <-chan *registry.NetworkServiceEndpoint) {
	ctx, cancel := context.WithCancel(context.Background())

	s.executor.AsyncExec(func() {
		delete(s.eventChannels, id)
		cancel()
	})

	for {
		select {
		case <-ctx.Done():
			return
		case <-eventCh:
		}
	}
}

func (s *memoryNSEServer) receiveEvent(
	query *registry.NetworkServiceEndpointQuery,
	server registry.NetworkServiceEndpointRegistry_FindServer,
	eventCh <-chan *registry.NetworkServiceEndpoint,
) error {
	select {
	case <-server.Context().Done():
		return io.EOF
	case event := <-eventCh:
		if matchutils.MatchNetworkServiceEndpoints(query.NetworkServiceEndpoint, event) {
			if err := server.Send(event); err != nil {
				if server.Context().Err() != nil {
					return io.EOF
				}
				return err
			}
		}
		return nil
	}
}

func (s *memoryNSEServer) Unregister(ctx context.Context, nse *registry.NetworkServiceEndpoint) (*empty.Empty, error) {
	if unregisterNSE, ok := s.networkServiceEndpoints.LoadAndDelete(nse.Name); ok {
		unregisterNSE = unregisterNSE.Clone()
		unregisterNSE.ExpirationTime = &timestamp.Timestamp{
			Seconds: -1,
		}
		s.sendEvent(unregisterNSE)
	}
	return next.NetworkServiceEndpointRegistryServer(ctx).Unregister(ctx, nse)
}
