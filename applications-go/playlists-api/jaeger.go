package main

import "github.com/uber/jaeger-client-go/config"

func newJaegerConfig() *config.Configuration {
	return &config.Configuration{
		ServiceName: serviceName,

		// "const" sampler is a binary sampling strategy: 0=never sample, 1=always sample.
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},

		// Log the emitted spans to stdout.
		Reporter: &config.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: jaeger_host_port,
		},
	}
}
