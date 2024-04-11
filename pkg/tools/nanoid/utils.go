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

package nanoid

import (
	"fmt"

	kernelmech "github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
)

const (
	ifPrefix = "nsm"
)

// GenerateLinuxInterfaceName - returns a random interface name with "nsm" prefix
// to achieve a 1% chance of name collision, you need to generate approximately 68 billon names
func GenerateLinuxInterfaceName() (string, error) {
	ifIDLen := kernelmech.LinuxIfMaxLength - len(ifPrefix)
	id, err := GenerateString(ifIDLen)
	if err != nil {
		return "", err
	}
	name := fmt.Sprintf("%s%s", ifPrefix, id)

	return name, nil
}
