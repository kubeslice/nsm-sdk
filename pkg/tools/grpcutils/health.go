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

package grpcutils

import (
	"context"
	"github.com/networkservicemesh/api/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"time"
)

// RegisterHealthServices registers grpc health probe for each passed service
func RegisterHealthServices(s *grpc.Server, services ...interface{}) {
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	for _, service := range services {
		for _, serviceName := range api.ServiceNames(service) {
			healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)
		}
	}
}

func WaitStatusServing(ctx context.Context, c grpc_health_v1.HealthClient, services ...interface{}) error {
	for _, service := range services {
		response := &grpc_health_v1.HealthCheckResponse{}
		for _, serviceName := range api.ServiceNames(service) {
			var err error
			for {
				response, err = c.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: serviceName})
				if err != nil {
					return err
				}
				if response.Status == grpc_health_v1.HealthCheckResponse_SERVING {
					break
				}
				time.Sleep(time.Millisecond * 100)
			}
		}
	}
	return nil
}
