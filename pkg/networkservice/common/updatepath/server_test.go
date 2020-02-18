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

package updatepath_test

import (
	"context"
	"testing"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/stretchr/testify/assert"

	"github.com/networkservicemesh/sdk/pkg/networkservice/common/updatepath"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
)

var serverTestData = []struct {
	name    string
	nscName string
	request *networkservice.NetworkServiceRequest
	want    *networkservice.Connection
}{
	{
		"add new segment when index was in the last position",
		"nsc-2",
		&networkservice.NetworkServiceRequest{
			Connection: &networkservice.Connection{
				Id: "conn-2",
				Path: &networkservice.Path{
					Index: 1,
					PathSegments: []*networkservice.PathSegment{
						{
							Name: "nsc-0",
							Id:   "conn-0",
						}, {
							Name: "nsc-1",
							Id:   "conn-1",
						},
					},
				},
			},
		},
		&networkservice.Connection{
			Id: "conn-2",
			Path: &networkservice.Path{
				Index: 2,
				PathSegments: []*networkservice.PathSegment{
					{
						Name: "nsc-0",
						Id:   "conn-0",
					}, {
						Name: "nsc-1",
						Id:   "conn-1",
					}, {
						Name: "nsc-2",
						Id:   "conn-2",
					},
				},
			},
		},
	},
	{
		"override values when index is valid in path array",
		"nsc-2",
		&networkservice.NetworkServiceRequest{
			Connection: &networkservice.Connection{
				Id: "conn-2",
				Path: &networkservice.Path{
					Index: 1,
					PathSegments: []*networkservice.PathSegment{
						{
							Name: "nsc-0",
							Id:   "conn-0",
						}, {
							Name: "nsc-1",
							Id:   "conn-1",
						}, {
							Name: "nsc-will-be-overridden",
							Id:   "conn-will-be-overridden",
						},
					},
				},
			},
		},
		&networkservice.Connection{
			Id: "conn-2",
			Path: &networkservice.Path{
				Index: 2,
				PathSegments: []*networkservice.PathSegment{
					{
						Name: "nsc-0",
						Id:   "conn-0",
					}, {
						Name: "nsc-1",
						Id:   "conn-1",
					}, {
						Name: "nsc-2",
						Id:   "conn-2",
					},
				},
			},
		},
	},
	{
		"index is greater or equal to path length",
		"nsc-1",
		&networkservice.NetworkServiceRequest{
			Connection: &networkservice.Connection{
				Id: "conn-1",
				Path: &networkservice.Path{
					Index: 2,
					PathSegments: []*networkservice.PathSegment{
						{
							Name: "nsc-0",
							Id:   "conn-0",
						}, {
							Name: "nsc-1",
							Id:   "conn-1",
						},
					},
				},
			},
		},
		nil,
	},
}

func Test_updatePathServer_Request(t *testing.T) {
	for _, data := range serverTestData {
		test := data
		t.Run(test.name, func(t *testing.T) {
			testServerRequest(t, test.nscName, test.request, test.want)
		})
	}
}

func testServerRequest(t *testing.T, nscName string, request *networkservice.NetworkServiceRequest, want *networkservice.Connection) {
	client := next.NewNetworkServiceServer(updatepath.NewServer(nscName))
	got, _ := client.Request(context.Background(), request)
	assert.Equal(t, want, got)
}
