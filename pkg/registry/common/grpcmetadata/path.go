// Copyright (c) 2022 Cisco and/or its affiliates.
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

package grpcmetadata

import (
	"github.com/networkservicemesh/api/pkg/api/networkservice"
)

// Path represents private path that is passed via grpcmetadata during NS and NSE registration
type Path networkservice.Path

// GetPrevPathSegment returns path.Index - 1 segments if it exists
func (p *Path) GetPrevPathSegment() *networkservice.PathSegment {
	if p == nil {
		return nil
	}
	if len(p.PathSegments) == 0 {
		return nil
	}
	if int(p.Index) == 0 {
		return nil
	}
	if int(p.Index)-1 > len(p.PathSegments) {
		return nil
	}
	return p.PathSegments[p.Index-1]
}

// Clone clones Path
func (p *Path) Clone() *Path {
	path := (*networkservice.Path)(p)
	return (*Path)(path.Clone())
}
