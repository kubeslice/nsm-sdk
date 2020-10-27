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

package fifosync

import (
	"sync"
)

// ChannelMutex is a FIFO order guaranteed mutex
type ChannelMutex struct {
	init    sync.Once
	channel chan struct{}
}

// Lock acquires lock or blocks until it gets free
func (m *ChannelMutex) Lock() {
	m.init.Do(func() {
		m.channel = make(chan struct{}, 1)
	})
	m.channel <- struct{}{}
}

// Unlock frees lock
func (m *ChannelMutex) Unlock() {
	<-m.channel
}
