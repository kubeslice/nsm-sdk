// Copyright (c) 2021 Cisco and/or its affiliates.
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

package dial

import (
	"context"
	"net/url"

	"github.com/networkservicemesh/sdk/pkg/networkservice/utils/metadata"
)

type key struct{}

// storeClientURL sets the *url.URL stored in per Connection.Id metadata.
func storeClientURL(ctx context.Context, isClient bool, u *url.URL) {
	metadata.Map(ctx, isClient).Store(key{}, u)
}

// deleteClientURL deletes the *url.URL stored in per Connection.Id metadata
func deleteClientURL(ctx context.Context, isClient bool) {
	metadata.Map(ctx, isClient).Delete(key{})
}

// loadClientURL returns the *url.URL stored in per Connection.Id metadata, or nil if no
// value is present.
// The ok result indicates whether value was found in the per Connection.Id metadata.
func loadClientURL(ctx context.Context, isClient bool) (value *url.URL, ok bool) {
	rawValue, ok := metadata.Map(ctx, isClient).Load(key{})
	if !ok {
		return
	}
	value, ok = rawValue.(*url.URL)
	return value, ok
}
