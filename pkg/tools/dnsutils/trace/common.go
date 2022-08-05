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
	"fmt"
	"strings"

	"github.com/miekg/dns"

	"github.com/networkservicemesh/sdk/pkg/tools/log"
)

func logRequest(ctx context.Context, message *dns.Msg, prefixes ...string) {
	msg := strings.Join(append(prefixes, "request"), "-")
	logObjectTrace(ctx, msg, message)

	// diffMsg := strings.Join(append(prefixes, "diff"), "-")

	// messageInfo, ok := trace(ctx)
	// if ok && !cmp.Equal(messageInfo.Message, message) {
	// 	if messageInfo.Message != nil {
	// 		messageDiff, _ := diff.Diff(messageInfo.Message, message)
	// 		if len(messageDiff) > 0 {
	// 			logObjectTrace(ctx, diffMsg, messageDiff)
	// 		}
	// 	} else {
	// 		logObjectTrace(ctx, msg, message)
	// 	}
	// 	messageInfo.Message = message.Copy()
	// 	return
	// }
}

func logResponse(ctx context.Context, rw dns.ResponseWriter, prefixes ...string) {
	msg := strings.Join(append(prefixes, "response"), "-")

	traceRW, ok := rw.(*traceResponseWriter)
	if ok {
		logObjectTrace(ctx, msg, traceRW.responseMsg)
	}
}

func logObjectTrace(ctx context.Context, k, v interface{}) {
	s := log.FromContext(ctx)
	s.Tracef(fmt.Sprintf("%v=%#v", k, v))
}
