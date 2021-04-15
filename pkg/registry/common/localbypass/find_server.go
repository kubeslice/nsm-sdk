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

package localbypass

import (
	"github.com/networkservicemesh/api/pkg/api/registry"
)

type localBypassNSEFindServer struct {
	*localBypassNSEServer
	registry.NetworkServiceEndpointRegistry_FindServer
}

func (s *localBypassNSEFindServer) Send(nse *registry.NetworkServiceEndpoint) error {
	if nse.Url != s.nsmgrURL {
		return s.NetworkServiceEndpointRegistry_FindServer.Send(nse)
	}
	if nse.ExpirationTime != nil && nse.ExpirationTime.Seconds == -1 {
		return s.NetworkServiceEndpointRegistry_FindServer.Send(nse)
	}

	u, ok := s.nseURLs.Load(nse.Name)
	if !ok {
		return nil
	}

	nse.Url = u.String()

	return s.NetworkServiceEndpointRegistry_FindServer.Send(nse)
}
