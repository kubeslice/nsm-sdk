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

package roundrobin

import (
	"context"
	"net/url"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	"github.com/networkservicemesh/networkservicemesh/controlplane/api/connection"
	"github.com/networkservicemesh/networkservicemesh/controlplane/api/networkservice"

	"github.com/networkservicemesh/sdk/pkg/networkservice/common/clienturl"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/discover"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
)

type selectEndpointServer struct {
	selector *roundRobinSelector
}

// NewServer - provides a NetworkServiceServer chain element that round robins among candidates provided by
// discover.Candidate(ctx) in the context.
func NewServer() networkservice.NetworkServiceServer {
	return &selectEndpointServer{
		selector: newRoundRobinSelector(),
	}
}

func (s *selectEndpointServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*connection.Connection, error) {
	ctx, err := s.withClientURL(ctx)
	if err != nil {
		return nil, err
	}
	// TODO - set Request.Connection.NetworkServiceEndpoint if unset
	return next.Server(ctx).Request(ctx, request)
}

func (s *selectEndpointServer) Close(ctx context.Context, conn *connection.Connection) (*empty.Empty, error) {
	// TODO - we should remember the previous selection here.
	ctx, err := s.withClientURL(ctx)
	if err != nil {
		return nil, err
	}
	return next.Server(ctx).Close(ctx, conn)
}

func (s *selectEndpointServer) withClientURL(ctx context.Context) (context.Context, error) {
	candidates := discover.Candidates(ctx)
	endpoint := s.selector.selectEndpoint(candidates.GetNetworkService(), candidates.GetNetworkServiceEndpoints())
	urlString := candidates.GetNetworkServiceManagers()[endpoint.GetNetworkServiceManagerName()].GetUrl()
	u, err := url.Parse(urlString)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	ctx = clienturl.WithClientURL(ctx, u)
	return ctx, nil
}
