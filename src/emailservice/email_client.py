#!/usr/bin/python
#
# Copyright 2018 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import grpc

import demo_pb2
import demo_pb2_grpc

from logger import getJSONLogger
logger = getJSONLogger('emailservice-client')

from opentelemetry import trace
from opentelemetry.instrumentation.grpc import client_interceptor
from opentelemetry.instrumentation.grpc.grpcext import intercept_channel
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.exporter import jaeger
from opentelemetry.sdk.trace.export import BatchExportSpanProcessor
from opentelemetry.sdk.trace.export import Span, SpanExporter, SpanExportResult
from opentelemetry.exporter.jaeger import _translate_to_jaeger
from opentelemetry.exporter.jaeger.gen.jaeger import Collector as jaegerType

trace.set_tracer_provider(TracerProvider())

# create a JaegerSpanExporter
jaeger_exporter = jaeger.JaegerSpanExporter(
    service_name=os.environ.get('SERVICE_NAME', "emailservice") + "-client",
    # configure agent
    #agent_host_name="localhost",
    #agent_port=6831,
    # optional: configure also collector
    collector_host_name=os.environ.get('JAEGER_HOST', 'jaeger-collector'),
    collector_port=14268,
    collector_endpoint="/api/traces?format=jaeger.thrift",
    # username=xxxx, # optional
    # password=xxxx, # optional
)

def new_export(spans):
	global jaeger_exporter
	self = jaeger_exporter
	jaeger_spans = _translate_to_jaeger(spans)
	podIp = os.environ.get('POD_IP')
	podName = os.environ.get('POD_NAME')
	nodeName = os.environ.get('NODE_NAME')

	tags = [
		jaegerType.Tag(key="exporter", vType=jaegerType.TagType.STRING, vStr="jaeger"),
		jaegerType.Tag(key="float", vType=jaegerType.TagType.DOUBLE, vDouble=312.23),
		jaegerType.Tag(key="ip", vType=jaegerType.TagType.STRING, vStr=podIp),
		jaegerType.Tag(key="name", vType=jaegerType.TagType.STRING, vStr=podName),
		jaegerType.Tag(key="node_name", vType=jaegerType.TagType.STRING, vStr=nodeName)
	]

	batch = jaegerType.Batch(
		spans=jaegerType,
		process=jaegerType.Process(serviceName=self.service_name, tags=tags),
	)

	if self.collector is not None:
		self.collector.submit(batch)
	self.agent_client.emit(batch)

	return SpanExportResult.SUCCESS
jaeger_exporter.export = new_export

# create a BatchExportSpanProcessor and add the exporter to it
span_processor = BatchExportSpanProcessor(jaeger_exporter)
trace.get_tracer_provider().add_span_processor(span_processor)

def send_confirmation_email(email, order):
    channel = grpc.insecure_channel('0.0.0.0:8080')
    channel = intercept_channel(channel, client_interceptor())
    stub = demo_pb2_grpc.EmailServiceStub(channel)
    try:
        response = stub.SendOrderConfirmation(demo_pb2.SendOrderConfirmationRequest(
            email=email,
            order=order
        ))
        logger.info('Request sent.')
    except grpc.RpcError as err:
        logger.error(err.details())
        logger.error('{}, {}'.format(err.code().name, err.code().value))


if __name__ == '__main__':
    logger.info('Client for email service.')
