// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package nats

import (
	"context"
	"fmt"
	nats "github.com/nats-io/nats.go"
	"github.com/openfaas-incubator/connector-sdk/types"
	"github.com/opentracing/opentracing-go"
	"github.com/sticreations/not.go"
	"log"
	"sync"
	"time"
)

const queueGroup = "openfaas_nats_worker_group"
const clientName = "openfaas_connector"

type BrokerConfig struct {
	Host        string
	ConnTimeout time.Duration
	Tracer      opentracing.Tracer
	GatewayURL  string
}

type broker struct {
	client     *nats.Conn
	tracer     opentracing.Tracer
	gatewayURL string
}

func NewBroker(config BrokerConfig) *broker {
	broker := &broker{}

	brokerURL := "nats://" + config.Host + ":4222"
	for {
		client, err := nats.Connect(brokerURL, nats.Timeout(config.ConnTimeout), nats.Name(clientName))
		if client != nil && err == nil {
			broker.client = client
			break
		}

		if client != nil {
			client.Close()
		}
		log.Println("Wait for brokers to come up.. ", brokerURL)
		time.Sleep(1 * time.Second)
		// TODO Add healthcheck
	}
	return broker
}

func (b *broker) Subscribe(controller *types.Controller, topics []string) {
	log.Printf("Configured topics: %v", topics)

	wg := sync.WaitGroup{}
	wg.Add(1)

	for _, topic := range topics {
		log.Printf("Binding to topic: %v", topic)
		b.client.QueueSubscribe(topic, queueGroup, func(m *nats.Msg) {

			log.Printf("Received topic: %s, message: %s", m.Subject, string(m.Data))
			t := not.NewTraceMsg(m)

			sc, err := b.tracer.Extract(opentracing.Binary, t)
			if err != nil {
				log.Printf("Could not Extract Opentracing, continuing with normal Message : %v", err)
				go b.handleMessageWithoutTracing(m.Subject, m.Data, controller)
				return
			}
			childSpan := b.tracer.StartSpan("Nats-Event", opentracing.ChildOf(sc))
			defer childSpan.Finish()
			go b.handleMessageWithTracing(m.Subject, m.Data, controller, childspan)

		})
	}

	// interrupt handling
	wg.Wait()
	b.client.Close()
}

func (b *broker) handleMessageWithoutTracing(subject string, data []byte, controller *types.Controller) {
	controller.Invoke(subject, &data)
}

func (b *broker) handleMessageWithTracing(subject string, data []byte, controller *types.Controller, span opentracing.Span) {
	url := fmt.Sprintf("%s/%s/%s", b.gatewayURL, "function", subject)
	var ctx context.Context
	ctx = opentracing.ContextWithSpan(ctx, span)
	controller.InvokeWithContext(ctx, subject, &data)
	// Not possible to map subject to Function right now :()
}
