// Copyright (c) 2022 Cisco Systems, Inc.
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

package trace

import (
	"context"
	"strconv"

	"github.com/miekg/dns"

	"github.com/networkservicemesh/sdk/pkg/tools/dnsutils"
	"github.com/networkservicemesh/sdk/pkg/tools/dnsutils/next"
	"github.com/networkservicemesh/sdk/pkg/tools/typeutils"
)

type beginTraceHandler struct {
	traced dnsutils.Handler
}

type endTraceHandler struct{}

// NewDNSHandler - wraps tracing around the supplied traced
func NewDNSHandler(traced dnsutils.Handler) dnsutils.Handler {
	return next.NewDNSHandler(
		&beginTraceHandler{traced: traced},
		&endTraceHandler{},
	)
}

func (t *beginTraceHandler) ServeDNS(ctx context.Context, rw dns.ResponseWriter, m *dns.Msg) {
	operation := typeutils.GetFuncName(t.traced, "ServeDNS")
	ctx, finish := withLog(ctx, operation, strconv.Itoa(int(m.Id)))
	defer finish()

	logMessage(ctx, m)
	t.traced.ServeDNS(ctx, rw, m)
}

func (t *endTraceHandler) ServeDNS(ctx context.Context, rw dns.ResponseWriter, m *dns.Msg) {
	logMessage(ctx, m)
	next.Handler(ctx).ServeDNS(ctx, rw, m)
}
