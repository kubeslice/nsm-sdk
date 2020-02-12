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

package clientinfo

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
)

type clientInfoTestData struct {
	name    string
	envs    map[string]string
	request *networkservice.NetworkServiceRequest
	want    *networkservice.Connection
}

var tests = []clientInfoTestData{
	{
		"the-labels-map-is-not-present",
		map[string]string{
			"NODE_NAME":    "AAA",
			"POD_NAME":     "BBB",
			"CLUSTER_NAME": "CCC",
		},
		&networkservice.NetworkServiceRequest{
			Connection: &networkservice.Connection{},
		},
		&networkservice.Connection{
			Labels: map[string]string{
				"NodeNameKey":    "AAA",
				"PodNameKey":     "BBB",
				"ClusterNameKey": "CCC",
			},
		},
	},
	{
		"the-labels-are-overwritten",
		map[string]string{
			"NODE_NAME":    "A1",
			"POD_NAME":     "B2",
			"CLUSTER_NAME": "C3",
		},
		&networkservice.NetworkServiceRequest{
			Connection: &networkservice.Connection{
				Labels: map[string]string{
					"NodeNameKey":     "OLD_VAL1",
					"PodNameKey":      "OLD_VAL2",
					"ClusterNameKey":  "OLD_VAL3",
					"SomeOtherLabel1": "DDD",
					"SomeOtherLabel2": "EEE",
				},
			},
		},
		&networkservice.Connection{
			Labels: map[string]string{
				"NodeNameKey":     "A1",
				"PodNameKey":      "B2",
				"ClusterNameKey":  "C3",
				"SomeOtherLabel1": "DDD",
				"SomeOtherLabel2": "EEE",
			},
		},
	},
	{
		"some-of-the-envs-are-not-present",
		map[string]string{
			"CLUSTER_NAME": "ABC",
		},
		&networkservice.NetworkServiceRequest{
			Connection: &networkservice.Connection{
				Labels: map[string]string{
					"NodeNameKey":     "OLD_VAL1",
					"ClusterNameKey":  "OLD_VAL2",
					"SomeOtherLabel1": "DDD",
					"SomeOtherLabel2": "EEE",
				},
			},
		},
		&networkservice.Connection{
			Labels: map[string]string{
				"NodeNameKey":     "OLD_VAL1",
				"ClusterNameKey":  "ABC",
				"SomeOtherLabel1": "DDD",
				"SomeOtherLabel2": "EEE",
			},
		},
	},
}

func Test_clientInfo_Request(t *testing.T) {
	server := next.NewNetworkServiceClient(NewClient())
	for _, testData := range tests {
		for name, value := range testData.envs {
			err := os.Setenv(name, value)
			if err != nil {
				t.Errorf("%s: clientInfo.Request() unable to set up environment variable: %v", testData.name, err)
			}
		}

		got, _ := server.Request(context.Background(), testData.request)
		if !reflect.DeepEqual(got, testData.want) {
			t.Errorf("%s: clientInfo.Request() = %v, want %v", testData.name, got, testData.want)
		}

		for name := range testData.envs {
			err := os.Unsetenv(name)
			if err != nil {
				t.Errorf("%s: clientInfo.Request() unable to unset environment variable: %v", testData.name, err)
			}
		}
	}
}
