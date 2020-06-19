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

// Package adapters provides adapters to translate between networkservice.NetworkService{Server,Client}
package adapters

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
)

type clientToServer struct {
	client networkservice.NetworkServiceClient
}

// NewClientToServer - returns a networkservice.NetworkServiceServer wrapped around the supplied client
func NewClientToServer(client networkservice.NetworkServiceClient) networkservice.NetworkServiceServer {
	return &clientToServer{client: client}
}

func (c *clientToServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	conn, err := c.client.Request(ctx, request)
	if err != nil {
		return nil, err
	}
	if request == nil {
		request = &networkservice.NetworkServiceRequest{}
	}
	request.Connection = conn
	if !next.Done(ctx) {
		return conn, nil
	}
	return next.Server(ctx).Request(ctx, request)
}

func (c *clientToServer) Close(ctx context.Context, request *networkservice.Connection) (*empty.Empty, error) {
	conn, err := c.client.Close(ctx, request)
	if err != nil {
		return nil, err
	}
	if !next.Done(ctx) {
		return conn, nil
	}
	return next.Server(ctx).Close(ctx, request)
}
