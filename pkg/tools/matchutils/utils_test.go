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

package matchutils_test

import (
	"testing"

	"github.com/networkservicemesh/api/pkg/api/networkservice/payload"
	"github.com/networkservicemesh/api/pkg/api/registry"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/networkservicemesh/sdk/pkg/tools/matchutils"
)

func TestNSMatch(t *testing.T) {
	referenceService := &registry.NetworkService{
		Name:    "ns-1-substring-match",
		Payload: payload.IP,
		Matches: []*registry.Match{
			{
				SourceSelector: map[string]string{
					"app": "firewall",
				},
				Routes: []*registry.Destination{
					{
						DestinationSelector: map[string]string{
							"app": "some-middle-app",
						},
					},
				},
			},
		},
	}

	type test struct {
		name string
		svc  *registry.NetworkService
		want bool
	}

	tests := []test{
		{
			name: "empty",
			svc:  &registry.NetworkService{},
			want: true,
		},
		{
			name: "same",
			svc:  referenceService,
			want: true,
		},
		{
			name: "matchName",
			svc: &registry.NetworkService{
				Name: referenceService.Name,
			},
			want: true,
		},
		{
			name: "matchNameSubstring",
			svc: &registry.NetworkService{
				Name: "substring-match",
			},
			want: true,
		},
		{
			name: "noMatchName",
			svc: &registry.NetworkService{
				Name: "different-name",
			},
			want: false,
		},
		{
			name: "matchPayload",
			svc: &registry.NetworkService{
				Payload: referenceService.Payload,
			},
			want: true,
		},
		{
			name: "noMatchPayload",
			svc: &registry.NetworkService{
				Payload: payload.Ethernet,
			},
			want: false,
		},
		{
			name: "matchMatches",
			svc: &registry.NetworkService{
				Matches: referenceService.Matches,
			},
			want: true,
		},
		{
			name: "noMatchPayload",
			svc: &registry.NetworkService{
				Matches: []*registry.Match{
					{
						SourceSelector: map[string]string{
							"app": "vpn",
						},
						Routes: []*registry.Destination{
							{
								DestinationSelector: map[string]string{
									"app": "some-middle-app",
								},
							},
						},
					},
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := matchutils.MatchNetworkServices(tc.svc, referenceService)
			if tc.want != got {
				t.Logf("matching right: %v", referenceService)
				t.Logf("matching left: %v", tc.svc)
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func TestNSEMatch(t *testing.T) {
	referenceEndpoint := &registry.NetworkServiceEndpoint{
		Name:                "nse-1-substring-match",
		NetworkServiceNames: []string{"nse-service-1", "nse-service-2"},
		NetworkServiceLabels: map[string]*registry.NetworkServiceLabels{
			"Service1": {
				Labels: map[string]string{
					"foo": "bar",
				},
			},
			"Service2": {
				Labels: map[string]string{
					"foo2": "bar2",
				},
			},
		},
		Url:            "tcp://1.1.1.1",
		ExpirationTime: &timestamppb.Timestamp{Seconds: 600},
	}

	type test struct {
		name     string
		endpoint *registry.NetworkServiceEndpoint
		want     bool
	}

	tests := []test{
		{
			name:     "empty",
			endpoint: &registry.NetworkServiceEndpoint{},
			want:     true,
		},
		{
			name:     "same",
			endpoint: referenceEndpoint,
			want:     true,
		},
		{
			name: "matchName",
			endpoint: &registry.NetworkServiceEndpoint{
				Name: referenceEndpoint.Name,
			},
			want: true,
		},
		{
			name: "matchNameSubstring",
			endpoint: &registry.NetworkServiceEndpoint{
				Name: "substring-match",
			},
			want: true,
		},
		{
			name: "noMatchName",
			endpoint: &registry.NetworkServiceEndpoint{
				Name: "different-name",
			},
			want: false,
		},
		{
			name: "matchLabels",
			endpoint: &registry.NetworkServiceEndpoint{
				NetworkServiceLabels: referenceEndpoint.NetworkServiceLabels,
			},
			want: true,
		},
		{
			name: "noMatchLabels",
			endpoint: &registry.NetworkServiceEndpoint{
				NetworkServiceLabels: map[string]*registry.NetworkServiceLabels{
					"Service3": {
						Labels: map[string]string{
							"foo": "bar",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "matchExpirationTime",
			endpoint: &registry.NetworkServiceEndpoint{
				ExpirationTime: referenceEndpoint.ExpirationTime,
			},
			want: true,
		},
		{
			name: "matchExpirationTimeWithoutNanos",
			endpoint: &registry.NetworkServiceEndpoint{
				ExpirationTime: &timestamppb.Timestamp{
					Seconds: 600,
					Nanos:   100000,
				},
			},
			want: true,
		},
		{
			name: "noMatchExpirationTime",
			endpoint: &registry.NetworkServiceEndpoint{
				ExpirationTime: &timestamppb.Timestamp{Seconds: 601},
			},
			want: false,
		},
		{
			name: "matchNetworkServiceNames",
			endpoint: &registry.NetworkServiceEndpoint{
				NetworkServiceNames: referenceEndpoint.NetworkServiceNames,
			},
			want: true,
		},
		{
			name: "matchNetworkServiceNamesPartial",
			endpoint: &registry.NetworkServiceEndpoint{
				NetworkServiceNames: []string{"nse-service-2"},
			},
			want: true,
		},
		{
			name: "noMatchNetworkServiceNames",
			endpoint: &registry.NetworkServiceEndpoint{
				NetworkServiceNames: []string{"nse-service-3"},
			},
			want: false,
		},
		{
			name: "matchUrl",
			endpoint: &registry.NetworkServiceEndpoint{
				Url: referenceEndpoint.Url,
			},
			want: true,
		},
		{
			name: "matchUrlPartial",
			endpoint: &registry.NetworkServiceEndpoint{
				Url: "://1",
			},
			want: true,
		},
		{
			name: "noMatchUrl",
			endpoint: &registry.NetworkServiceEndpoint{
				Url: "udp://",
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := matchutils.MatchNetworkServiceEndpoints(tc.endpoint, referenceEndpoint)
			if tc.want != got {
				t.Logf("matching right: %v", referenceEndpoint)
				t.Logf("matching left: %v", tc.endpoint)
				require.Equal(t, tc.want, got)
			}
		})
	}
}
