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

package setid_test

import (
	"context"
	"testing"

	"github.com/networkservicemesh/sdk/pkg/networkservice/utils/checks/checkrequest"

	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk/pkg/networkservice/common/setid"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
)

func Test_idServer_Request(t *testing.T) {
	for _, data := range testData {
		test := data
		t.Run(test.name, func(t *testing.T) {
			testServerRequest(t, test.clientName, test.path, test.connectionID, test.testFuncProvider(test.connectionID))
		})
	}
}

func testServerRequest(t *testing.T, serverName string, path *networkservice.Path, connectionID string, testFunc func(*testing.T, *networkservice.NetworkServiceRequest)) {
	client := next.NewNetworkServiceServer(setid.NewServer(serverName), checkrequest.NewServer(t, testFunc))
	request := &networkservice.NetworkServiceRequest{
		Connection: &networkservice.Connection{
			Id:   connectionID,
			Path: path,
		},
	}
	_, _ = client.Request(context.Background(), request)
}
