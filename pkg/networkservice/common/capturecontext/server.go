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

package capturecontext

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
)

type contextServer struct{}

func (c *contextServer) Request(ctx context.Context, in *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	captureContext(ctx)
	return next.Server(ctx).Request(ctx, in)
}

func (c *contextServer) Close(ctx context.Context, in *networkservice.Connection) (*empty.Empty, error) {
	captureContext(ctx)
	return next.Server(ctx).Close(ctx, in)
}

// NewServer - creates a new networkservice.NetworkServiceServer chain element that store context
// from the adapter server/client and pass it to the next client/server to avoid the problem with losing
// values from adapted server/client context.
func NewServer() networkservice.NetworkServiceServer {
	return &contextServer{}
}
