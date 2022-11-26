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

package updatepath_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk/pkg/registry/common/grpcmetadata"
)

const (
	nse1           = "nse-1"
	nse2           = "nse-2"
	nse3           = "nse-3"
	pathSegmentID1 = "36ce7f0c-9f6d-40a4-8b39-6b56ff07eea9"
	pathSegmentID2 = "ece490ea-dfe8-4512-a3ca-5be7b39515c5"
	pathSegmentID3 = "f9a83e55-0a4f-3647-144a-98a9ee8fb231"
	different      = "different"
)

func makePath(pathIndex uint32, pathSegments int) *grpcmetadata.Path {
	if pathSegments == 0 {
		return nil
	}

	path := &grpcmetadata.Path{
		Index: pathIndex,
	}
	if pathSegments >= 1 {
		path.PathSegments = append(path.PathSegments, &networkservice.PathSegment{
			Name: nse1,
			Id:   pathSegmentID1,
		})
	}
	if pathSegments >= 2 {
		path.PathSegments = append(path.PathSegments, &networkservice.PathSegment{
			Name: nse2,
			Id:   pathSegmentID2,
		})
	}
	if pathSegments >= 3 {
		path.PathSegments = append(path.PathSegments, &networkservice.PathSegment{
			Name: nse3,
			Id:   pathSegmentID3,
		})
	}
	return path
}

func requirePathEqual(t *testing.T, expected, actual *grpcmetadata.Path, unknownIDs ...int) {
	expected = expected.Clone()
	actual = actual.Clone()
	for _, index := range unknownIDs {
		expected.PathSegments[index].Id = ""
		actual.PathSegments[index].Id = ""
	}

	expectedString, err := json.Marshal(expected)
	require.NoError(t, err)
	actualString, err := json.Marshal(actual)
	require.NoError(t, err)
	require.Equal(t, string(expectedString), string(actualString))
}
