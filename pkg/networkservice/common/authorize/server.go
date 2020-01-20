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

// Package authorize provides authz checks for incoming or returning connections.
package authorize

import (
	"context"
	"github.com/open-policy-agent/opa/rego"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/networkservicemesh/networkservicemesh/controlplane/api/connection"
	"github.com/networkservicemesh/networkservicemesh/controlplane/api/networkservice"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
)

type authorizeServer struct {
	policy *rego.PreparedEvalQuery
}

// NewServer - returns a new authorization networkservicemesh.NetworkServiceServers
func NewServer(p *rego.PreparedEvalQuery) networkservice.NetworkServiceServer {
	return &authorizeServer{
		policy: p,
	}
}

func (a *authorizeServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*connection.Connection, error) {
	rs, err := a.policy.Eval(ctx, rego.EvalInput(request.GetConnection().GetPath().GetPathSegments()[request.GetConnection().GetPath().GetIndex()].GetToken()))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := check(rs); err != nil {
		return nil, err
	}

	return next.Server(ctx).Request(ctx, request)
}

func (a *authorizeServer) Close(ctx context.Context, conn *connection.Connection) (*empty.Empty, error) {
	rs, err := a.policy.Eval(ctx, rego.EvalInput(conn.GetPath().GetPathSegments()[conn.GetPath().GetIndex()].GetToken()))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := check(rs); err != nil {
		return nil, err
	}

	return next.Server(ctx).Close(ctx, conn)
}

func check(rs rego.ResultSet) error {
	for _, r := range rs {
		for _, e := range r.Expressions {
			t, ok := e.Value.(bool)
			if !ok {
				return status.Error(codes.Internal, "non boolean expression")
			}

			if !t {
				return status.Error(codes.PermissionDenied, "no sufficient privileges to call Request")
			}
		}
	}
	return nil
}
