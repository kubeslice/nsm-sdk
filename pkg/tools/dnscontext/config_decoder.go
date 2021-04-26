// Copyright (c) 2021 Doc.ai and/or its affiliates.
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

package dnscontext

import (
	"encoding/json"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
)

// Decoder allows to parse []*networkservice.DNSConfig from json string. Can be used for env configuration.
// See at https://github.com/kelseyhightower/envconfig#custom-decoders
type Decoder []*networkservice.DNSConfig

// Decode parses values from passed string.
func (d *Decoder) Decode(v string) error {
	var c []*networkservice.DNSConfig
	if err := json.Unmarshal([]byte(v), &c); err != nil {
		return err
	}
	*d = Decoder(c)
	return nil
}
