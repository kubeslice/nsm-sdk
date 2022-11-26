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

package updatetoken

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/networkservicemesh/api/pkg/api/registry"

	"github.com/networkservicemesh/sdk/pkg/registry/common/grpcmetadata"
	"github.com/networkservicemesh/sdk/pkg/registry/core/next"
	"github.com/networkservicemesh/sdk/pkg/tools/log"
	"github.com/networkservicemesh/sdk/pkg/tools/token"
)

type updateTokenNSEServer struct {
	tokenGenerator token.GeneratorFunc
}

// NewNetworkServiceEndpointRegistryServer - creates a NetworkServiceEndpointRegistryServer chain element to update NetworkServiceEndpoint.Path token information
func NewNetworkServiceEndpointRegistryServer(tokenGenerator token.GeneratorFunc) registry.NetworkServiceEndpointRegistryServer {
	return &updateTokenNSEServer{
		tokenGenerator: tokenGenerator,
	}
}

func (s *updateTokenNSEServer) Register(ctx context.Context, nse *registry.NetworkServiceEndpoint) (*registry.NetworkServiceEndpoint, error) {
	path, err := grpcmetadata.PathFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if prev := path.GetPrevPathSegment(); prev != nil {
		tok, expireTime, tokenErr := token.FromContext(ctx)

		if tokenErr != nil {
			log.FromContext(ctx).Warnf("an error during getting token from the context: %+v", tokenErr)
		} else {
			expires := expireTime.Local()
			prev.Expires = timestamppb.New(expires)
			prev.Token = tok
			id, idErr := getIDFromToken(tok)
			if idErr != nil {
				return nil, idErr
			}
			nse.PathIds = updatePathIds(nse.PathIds, int(path.Index-1), id.String())
		}
	}

	err = updateToken(ctx, path, s.tokenGenerator)
	if err != nil {
		return nil, err
	}

	id, err := getIDFromToken(path.PathSegments[path.Index].Token)
	if err != nil {
		return nil, err
	}
	nse.PathIds = updatePathIds(nse.PathIds, int(path.Index), id.String())
	return next.NetworkServiceEndpointRegistryServer(ctx).Register(ctx, nse)
}

func (s *updateTokenNSEServer) Find(query *registry.NetworkServiceEndpointQuery, server registry.NetworkServiceEndpointRegistry_FindServer) error {
	return next.NetworkServiceEndpointRegistryServer(server.Context()).Find(query, server)
}

func (s *updateTokenNSEServer) Unregister(ctx context.Context, nse *registry.NetworkServiceEndpoint) (*empty.Empty, error) {
	path, err := grpcmetadata.PathFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if prev := path.GetPrevPathSegment(); prev != nil {
		tok, expireTime, tokenErr := token.FromContext(ctx)

		if tokenErr != nil {
			log.FromContext(ctx).Warnf("an error during getting token from the context: %+v", tokenErr)
		} else {
			expires := expireTime.Local()
			prev.Expires = timestamppb.New(expires)
			prev.Token = tok

			id, idErr := getIDFromToken(tok)
			if idErr != nil {
				return nil, idErr
			}
			nse.PathIds = updatePathIds(nse.PathIds, int(path.Index-1), id.String())
		}
	}
	err = updateToken(ctx, path, s.tokenGenerator)
	if err != nil {
		return nil, err
	}

	id, err := getIDFromToken(path.PathSegments[path.Index].Token)
	if err != nil {
		return nil, err
	}
	nse.PathIds = updatePathIds(nse.PathIds, int(path.Index), id.String())

	return next.NetworkServiceEndpointRegistryServer(ctx).Unregister(ctx, nse)
}
