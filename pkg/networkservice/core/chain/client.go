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

package chain

import (
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/sirupsen/logrus"

	"github.com/networkservicemesh/sdk/pkg/networkservice/common/tracehelper"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/trace"
)

// NewNetworkServiceClient - chains together a list of networkservice.NetworkServiceClient with tracing
func NewNetworkServiceClient(clients ...networkservice.NetworkServiceClient) networkservice.NetworkServiceClient {
	if logrus.GetLevel() == logrus.TraceLevel {
		clients = append([]networkservice.NetworkServiceClient{tracehelper.NewClient()}, clients...)
		return next.NewWrappedNetworkServiceClient(trace.NewNetworkServiceClient, clients...)
	}
	return next.NewNetworkServiceClient(clients...)
}
