// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package main

import (
	"github.com/openfaas-incubator/connector-sdk/types"
	"github.com/sticreations/nats-conn/config"
	"github.com/sticreations/nats-conn/nats"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"log"
)

func main() {
	creds := types.GetCredentials()

	config := config.Get()

	cfg := jaegercfg.Configuration{
		ServiceName: "Nats-Connector",
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jlog := jaegerlog.StdLogger

	// Initialize tracer with a logger and a metrics factory
	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jlog),
	)
	if err != nil {
		log.Printf("Could not initialize jaeger tracer: %s", err.Error())
		return
	}
	defer closer.Close()

	controllerConfig := &types.ControllerConfig{
		RebuildInterval: config.RebuildInterval,
		GatewayURL:      config.GatewayURL,
		PrintResponse:   config.PrintResponse,
	}

	controller := types.NewController(creds, controllerConfig)
	controller.BeginMapBuilder()

	brokerConfig := nats.BrokerConfig{
		Host:        config.Broker,
		ConnTimeout: config.UpstreamTimeout,
		Tracer:      tracer,
		GatewayURL:  config.GatewayURL,
		ConcurrentRequests: config.ConcurrentRequests
	}

	broker := nats.NewBroker(brokerConfig)
	broker.Subscribe(controller, config.Topics)
}
