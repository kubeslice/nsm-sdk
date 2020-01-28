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
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/protobuf/ptypes/empty"

	
	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
)

type authorizeServer struct {
	p *rego.PreparedEvalQuery
}

// NewServer - returns a new authorization networkservicemesh.NetworkServiceServers
func NewServer(p *rego.PreparedEvalQuery) networkservice.NetworkServiceServer {
	return &authorizeServer{
		p: p,
	}
}

func (a *authorizeServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	if err := check(ctx, a.p, request.GetConnection().GetPath().GetPathSegments()[request.GetConnection().GetPath().GetIndex()].GetToken()); err != nil {
		return nil, err
	}

	return next.Server(ctx).Request(ctx, request)
}

func (a *authorizeServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	if err := check(ctx, a.p, conn.GetPath().GetPathSegments()[conn.GetPath().GetIndex()].GetToken()); err != nil {
		return nil, err
	}

	return next.Server(ctx).Close(ctx, conn)
}

func check(ctx context.Context, p *rego.PreparedEvalQuery, input interface{}) error {
	rs, err := p.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	hasAccess, err := hasAccess(rs)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if !hasAccess {
		return status.Error(codes.PermissionDenied, "no sufficient privileges to call Request")
	}

	return nil
}

func hasAccess(rs rego.ResultSet) (bool, error) {
	for _, r := range rs {
		for _, e := range r.Expressions {
			t, ok := e.Value.(bool)
			if !ok {
				return false, errors.New("policy contains non boolean expression")
			}

			if !t {
				return false, nil
			}
		}
	}

	return true, nil
}
