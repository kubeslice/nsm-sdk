// Copyright (c) 2021 Doc.ai and/or its affiliates.
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

// Package discoverforwarder discovers forwarder from the registry.
package discoverforwarder

import (
	"context"
	"net/url"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/registry"
	"github.com/pkg/errors"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
	"github.com/networkservicemesh/sdk/pkg/tools/clienturlctx"
	"github.com/networkservicemesh/sdk/pkg/tools/log"
	"github.com/networkservicemesh/sdk/pkg/tools/matchutils"
)

type discoverForwarderServer struct {
	nseClient            registry.NetworkServiceEndpointRegistryClient
	forwarderServiceName string
}

// NewServer creates new instance of discoverforwarder networkservice.NetworkServiceServer.
// Requires not nil nseClient.
func NewServer(nseClient registry.NetworkServiceEndpointRegistryClient, opts ...Option) networkservice.NetworkServiceServer {
	if nseClient == nil {
		panic("mseClient can not be nil")
	}
	var result = &discoverForwarderServer{
		nseClient:            nseClient,
		forwarderServiceName: "forwarder",
	}

	for _, opt := range opts {
		opt(result)
	}

	return result
}

func (d *discoverForwarderServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	var forwarderName = loadForwarderName(ctx)
	var logger = log.FromContext(ctx).WithField("discoverForwarderServer", "request")

	if forwarderName == "" {
		stream, err := d.nseClient.Find(ctx, &registry.NetworkServiceEndpointQuery{
			NetworkServiceEndpoint: &registry.NetworkServiceEndpoint{
				NetworkServiceNames: []string{
					d.forwarderServiceName,
				},
			},
		})

		if err != nil {
			logger.Errorf("can not open registry nse stream by networkservice. Error: %v", err.Error())
			return nil, errors.WithStack(err)
		}

		nses := d.matchForwarders(request.Connection.GetLabels(), registry.ReadNetworkServiceEndpointList(stream))

		if len(nses) == 0 {
			return nil, errors.New("no candidates found")
		}

		// TODO: Should we consider about load balancing?
		// https://github.com/networkservicemesh/sdk/issues/790
		for _, candidate := range nses {
			u, err := url.Parse(candidate.Url)

			if err != nil {
				logger.Errorf("can not parse forwarder=%v url=%v error=%v", candidate.Name, candidate.Url, err.Error())
				continue
			}

			resp, err := next.Server(ctx).Request(clienturlctx.WithClientURL(ctx, u), request.Clone())

			if err == nil {
				storeForwarderName(ctx, nses[0].Name)
				return resp, nil
			}
		}

		return nil, errors.New("all forwarders failed")
	}
	stream, err := d.nseClient.Find(ctx, &registry.NetworkServiceEndpointQuery{
		NetworkServiceEndpoint: &registry.NetworkServiceEndpoint{
			Name: forwarderName,
		},
	})

	if err != nil {
		logger.Errorf("can not open registry nse stream by forwarder name. Error: %v", err.Error())
		return nil, errors.WithStack(err)
	}

	nses := registry.ReadNetworkServiceEndpointList(stream)

	if len(nses) == 0 {
		storeForwarderName(ctx, "")
		return nil, errors.New("forwarder not found")
	}

	u, err := url.Parse(nses[0].Url)

	if err != nil {
		logger.Errorf("can not parse forwarder=%v url=%v error=%v", nses[0].Name, u, err.Error())
		return nil, errors.WithStack(err)
	}

	return next.Server(ctx).Request(clienturlctx.WithClientURL(ctx, u), request)
}

func (d *discoverForwarderServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	var forwarderName = loadForwarderName(ctx)
	var logger = log.FromContext(ctx).WithField("discoverForwarderServer", "request")

	if forwarderName == "" {
		return nil, errors.New("forwarder is not selected")
	}

	stream, err := d.nseClient.Find(ctx, &registry.NetworkServiceEndpointQuery{
		NetworkServiceEndpoint: &registry.NetworkServiceEndpoint{
			Name: forwarderName,
		},
	})

	if err != nil {
		logger.Errorf("can not open registry nse stream by forwarder name. Error: %v", err.Error())
		return nil, errors.WithStack(err)
	}

	nses := registry.ReadNetworkServiceEndpointList(stream)

	if len(nses) == 0 {
		return nil, errors.New("forwarder not found")
	}

	u, err := url.Parse(nses[0].Url)

	if err != nil {
		logger.Errorf("can not parse forwarder url %v", err.Error())
		return nil, errors.WithStack(err)
	}

	ctx = clienturlctx.WithClientURL(ctx, u)
	return next.Server(ctx).Close(ctx, conn)
}

func (d *discoverForwarderServer) matchForwarders(clientLabels map[string]string, canidates []*registry.NetworkServiceEndpoint) []*registry.NetworkServiceEndpoint {
	var result []*registry.NetworkServiceEndpoint

	for _, candidate := range canidates {
		var forwawrderLabels map[string]string

		if candidate.NetworkServiceLabels[d.forwarderServiceName] != nil {
			forwawrderLabels = candidate.NetworkServiceLabels[d.forwarderServiceName].Labels
		}

		if matchutils.IsSubset(clientLabels, forwawrderLabels, clientLabels) {
			result = append(result, candidate)
		}
	}

	return result
}
