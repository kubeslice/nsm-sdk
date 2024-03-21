// Copyright (c) 2024 Cisco and/or its affiliates.
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

package kernel_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	kernelmech "github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"

	"github.com/networkservicemesh/sdk/pkg/networkservice/common/mechanisms/kernel"

	"github.com/networkservicemesh/sdk/pkg/tools/nanoid"
)

func TestKernelMechanismServer_ShouldSetInterfaceName(t *testing.T) {
	var expectedIfaceName string
	for i := 0; i < kernelmech.LinuxIfMaxLength; i++ {
		expectedIfaceName += "a"
	}

	c := kernel.NewClient(kernel.WithInterfaceName(expectedIfaceName + "long-suffix"))

	req := &networkservice.NetworkServiceRequest{}
	_, err := c.Request(context.Background(), req)
	require.NoError(t, err)

	require.Len(t, req.MechanismPreferences, 1)
	require.Equal(t, expectedIfaceName, req.MechanismPreferences[0].Parameters[kernelmech.InterfaceNameKey])
}

func TestKernelMechanismServer_ShouldSetValidNetNSURL(t *testing.T) {
	c := kernel.NewClient()

	req := &networkservice.NetworkServiceRequest{
		MechanismPreferences: []*networkservice.Mechanism{
			kernelmech.New("invalid-url"),
		},
	}

	_, err := c.Request(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, netNSURL, req.MechanismPreferences[0].Parameters[kernelmech.NetNSURL])
}

func TestKernelMechanismServer_ShouldSetRandomInteraceName(t *testing.T) {
	c := kernel.NewClient()
	req := &networkservice.NetworkServiceRequest{}

	_, err := c.Request(context.Background(), req)
	require.NoError(t, err)

	ifname := req.MechanismPreferences[0].Parameters[kernelmech.InterfaceNameKey]

	require.Len(t, ifname, kernelmech.LinuxIfMaxLength)
	require.True(t, strings.HasPrefix(ifname, "nsm"))
	for i := 0; i < kernelmech.LinuxIfMaxLength; i++ {
		require.Contains(t, nanoid.DefaultAlphabet, string(ifname[i]))
	}
}

func TestKernelMechanismServer_FailedToGenerateRandomName(t *testing.T) {
	c := kernel.NewClient(kernel.WithInterfaceNameGenerator(nanoid.New(nanoid.WithRandomByteGenerator(&brokenByteGenerator{}))))
	req := &networkservice.NetworkServiceRequest{}

	_, err := c.Request(context.Background(), req)
	require.Error(t, err)
}
