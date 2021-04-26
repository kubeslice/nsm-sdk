// Copyright (c) 2020-2021 Doc.ai and/or its affiliates.
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
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/edwarnicke/serialize"

	"github.com/networkservicemesh/sdk/pkg/tools/log"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
	"github.com/networkservicemesh/sdk/pkg/tools/dnscontext"
)

type dnsContextClient struct {
	chainContext        context.Context
	coreFilePath        string
	resolveConfigPath   string
	defaultNameServerIP string
	dnsConfigManager    dnscontext.Manager
	updateCorefileQueue serialize.Executor
	once                sync.Once
}

// NewClient creates a new DNS client chain component. Setups all DNS traffic to the localhost. Monitors DNS configs from connections.
func NewClient(options ...DNSOption) networkservice.NetworkServiceClient {
	c := &dnsContextClient{
		chainContext:        context.Background(),
		defaultNameServerIP: "127.0.0.1",
		resolveConfigPath:   "/etc/resolv.conf",
		coreFilePath:        "/etc/coredns/Corefile",
	}
	for _, o := range options {
		o.apply(c)
	}

	return c
}

func (c *dnsContextClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*networkservice.Connection, error) {
	rv, err := next.Client(ctx).Request(ctx, request, opts...)
	if err != nil {
		return nil, err
	}
	var conifgs []*networkservice.DNSConfig
	if rv.GetContext().GetDnsContext() != nil {
		conifgs = rv.GetContext().GetDnsContext().GetConfigs()
	}
	if len(conifgs) > 0 {
		c.once.Do(c.initialize)
		c.dnsConfigManager.Store(rv.GetId(), conifgs...)
		c.updateCorefileQueue.AsyncExec(c.updateCorefile)
	}
	return rv, err
}

func (c *dnsContextClient) updateCorefile() {
	dir := filepath.Dir(c.coreFilePath)
	_ = os.MkdirAll(dir, os.ModePerm)
	err := ioutil.WriteFile(c.coreFilePath, []byte(c.dnsConfigManager.String()), os.ModePerm)
	if err != nil {
		log.FromContext(c.chainContext).Errorf("An error during update corefile: %v", err.Error())
	}
}

func (c *dnsContextClient) Close(ctx context.Context, conn *networkservice.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	var conifgs []*networkservice.DNSConfig
	if conn.GetContext().GetDnsContext() != nil {
		conifgs = conn.GetContext().GetDnsContext().GetConfigs()
	}
	if len(conifgs) > 0 {
		c.dnsConfigManager.Remove(conn.GetId())
		c.updateCorefileQueue.AsyncExec(c.updateCorefile)
	}
	return next.Client(ctx).Close(ctx, conn, opts...)
}

func (c *dnsContextClient) initialize() {
	r, err := dnscontext.OpenResolveConfig(c.resolveConfigPath)
	if err != nil {
		log.FromContext(c.chainContext).Errorf("An error during open resolve config: %v", err.Error())
		return
	}

	c.dnsConfigManager.Store("", &networkservice.DNSConfig{
		SearchDomains: r.Value(dnscontext.AnyDomain),
		DnsServerIps:  r.Value(dnscontext.NameserverProperty),
	})

	r.SetValue(dnscontext.NameserverProperty, c.defaultNameServerIP)

	if err = r.Save(); err != nil {
		log.FromContext(c.chainContext).Errorf("An error during save resolve config: %v", err.Error())
		return
	}

	c.updateCorefileQueue.AsyncExec(c.updateCorefile)
}
