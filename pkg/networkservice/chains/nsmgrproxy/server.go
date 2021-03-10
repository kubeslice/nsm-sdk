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

// Package nsmgrproxy provides chain of networkservice.NetworkServiceServer chain elements to creating NSMgrProxy
package nsmgrproxy

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk/pkg/networkservice/chains/client"
	"github.com/networkservicemesh/sdk/pkg/networkservice/chains/endpoint"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/connect"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/externalips"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/interdomainurl"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/swapip"
	"github.com/networkservicemesh/sdk/pkg/tools/token"
)

// NewServer creates new proxy NSMgr
func NewServer(ctx context.Context, name string, authorizeServer networkservice.NetworkServiceServer, tokenGenerator token.GeneratorFunc, options ...grpc.DialOption) endpoint.Endpoint {
	return endpoint.NewServer(ctx, tokenGenerator,
		endpoint.WithName(name),
		endpoint.WithAuthorizeServer(authorizeServer),
		endpoint.WithAdditionalFunctionality(
			interdomainurl.NewServer(),
			externalips.NewServer(ctx),
			swapip.NewServer(),
			connect.NewServer(ctx, 100*time.Millisecond,
				client.NewClientFactory(client.WithName(name)),
				options...,
			),
		),
	)
}
