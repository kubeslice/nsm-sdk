// Copyright (c) 2022-2023 Cisco and/or its affiliates.
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

package cache

import (
	"github.com/edwarnicke/genericsync"
	"github.com/miekg/dns"
)

type responseWriterWrapper struct {
	dns.ResponseWriter
	cache *genericsync.Map[dns.Question, *dns.Msg]
}

func (r *responseWriterWrapper) WriteMsg(m *dns.Msg) error {
	if m != nil && m.Rcode == dns.RcodeSuccess {
		r.cache.Store(m.Question[0], m)
	}
	return r.ResponseWriter.WriteMsg(m)
}
