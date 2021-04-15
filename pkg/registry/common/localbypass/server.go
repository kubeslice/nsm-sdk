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

// Package localbypass implements a chain element to set NSMgr URL to endpoints on registration and set back endpoints
// URLs on find
package localbypass

import (
	"context"
	"net/url"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	"github.com/networkservicemesh/api/pkg/api/registry"

	"github.com/networkservicemesh/sdk/pkg/registry/core/next"
	"github.com/networkservicemesh/sdk/pkg/tools/stringurl"
)

type localBypassNSEServer struct {
	nsmgrURL string
	nseURLs  stringurl.Map
}

// NewNetworkServiceEndpointRegistryServer creates new instance of NetworkServiceEndpointRegistryServer which sets
// NSMgr URL to endpoints on registration and sets back endpoints URLs on find
func NewNetworkServiceEndpointRegistryServer(nsmgrURL string) registry.NetworkServiceEndpointRegistryServer {
	return &localBypassNSEServer{
		nsmgrURL: nsmgrURL,
	}
}

func (s *localBypassNSEServer) Register(ctx context.Context, nse *registry.NetworkServiceEndpoint) (reg *registry.NetworkServiceEndpoint, err error) {
	u, loaded := s.nseURLs.Load(nse.Name)
	if !loaded {
		u, err = url.Parse(nse.Url)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot register NSE with passed URL: %s", nse.Url)
		}
		if u.String() == "" {
			return nil, errors.Errorf("cannot register NSE with passed URL: %s", nse.Url)
		}

		s.nseURLs.Store(nse.Name, u)
	}

	nse.Url = s.nsmgrURL

	reg, err = next.NetworkServiceEndpointRegistryServer(ctx).Register(ctx, nse)
	if err != nil {
		if !loaded {
			s.nseURLs.Delete(nse.Name)
		}
		return nil, err
	}

	reg.Url = u.String()

	return reg, nil
}

func (s *localBypassNSEServer) Find(query *registry.NetworkServiceEndpointQuery, server registry.NetworkServiceEndpointRegistry_FindServer) error {
	return next.NetworkServiceEndpointRegistryServer(server.Context()).Find(query, &localBypassNSEFindServer{
		localBypassNSEServer:                      s,
		NetworkServiceEndpointRegistry_FindServer: server,
	})
}

func (s *localBypassNSEServer) Unregister(ctx context.Context, nse *registry.NetworkServiceEndpoint) (_ *empty.Empty, err error) {
	if _, ok := s.nseURLs.Load(nse.Name); ok {
		nse.Url = s.nsmgrURL

		_, err = next.NetworkServiceEndpointRegistryServer(ctx).Unregister(ctx, nse)

		s.nseURLs.Delete(nse.Name)
	}
	return new(empty.Empty), err
}
