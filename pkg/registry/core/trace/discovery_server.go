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

package trace

import (
	"context"
	"fmt"

	"github.com/networkservicemesh/api/pkg/api/registry"
	"github.com/pkg/errors"

	"github.com/networkservicemesh/sdk/pkg/tools/spanhelper"
	"github.com/networkservicemesh/sdk/pkg/tools/typeutils"
)

type traceDiscoveryServer struct {
	traced registry.NetworkServiceDiscoveryServer
}

// NewDiscoveryServer - wraps registry.NetworkServiceDiscoveryServer with tracing
func NewDiscoveryServer(traced registry.NetworkServiceDiscoveryServer) registry.NetworkServiceDiscoveryServer {
	return &traceDiscoveryServer{traced: traced}
}

func (t *traceDiscoveryServer) FindNetworkService(ctx context.Context, request *registry.FindNetworkServiceRequest) (*registry.FindNetworkServiceResponse, error) {
	// Create a new span
	operation := fmt.Sprintf("%s/%s.FindNetworkService", typeutils.GetPkgPath(t.traced), typeutils.GetTypeName(t.traced))
	span := spanhelper.FromContext(ctx, operation)
	defer span.Finish()

	// Make sure we log to span

	ctx = withLog(span.Context(), span.Logger())

	span.LogObject("request", request)

	// Actually call the next
	rv, err := t.traced.FindNetworkService(ctx, request)

	if err != nil {
		if _, ok := err.(stackTracer); !ok {
			err = errors.Wrapf(err, "Error returned from %s/", operation)
		}
		span.LogError(err)
		return nil, err
	}
	span.LogObject("response", rv)
	return rv, err
}
