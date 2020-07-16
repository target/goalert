package app

import (
	"context"

	"github.com/target/goalert/util/log"

	"cloud.google.com/go/compute/metadata"
	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/monitoredresource"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

func configTracing(ctx context.Context, c Config) ([]trace.Exporter, error) {
	var exporters []trace.Exporter
	if c.JaegerEndpoint != "" || c.JaegerAgentEndpoint != "" {
		exporter, err := jaeger.NewExporter(jaeger.Options{
			Endpoint:      c.JaegerEndpoint,
			AgentEndpoint: c.JaegerAgentEndpoint,
			ServiceName:   "goalert",
		})
		if err != nil {
			return nil, errors.Wrap(err, "init jaeger exporter")
		}
		e := c.wrapExporter(exporter)
		exporters = append(exporters, e)
		trace.RegisterExporter(recoverExporter{exp: e})
	}

	if c.StackdriverProjectID != "" {
		opts := stackdriver.Options{
			ProjectID: c.StackdriverProjectID,
		}
		if c.TracingClusterName != "" {
			instanceID, err := metadata.InstanceID()
			if err != nil {
				log.Log(ctx, errors.Wrap(err, "get instance ID"))
				instanceID = "unknown"
			}
			zone, err := metadata.Zone()
			if err != nil {
				log.Log(ctx, errors.Wrap(err, "get zone"))
				zone = "unknown"
			}
			opts.MonitoredResource = &monitoredresource.GKEContainer{
				ProjectID:     c.StackdriverProjectID,
				InstanceID:    instanceID,
				ClusterName:   c.TracingClusterName,
				ContainerName: c.TracingContainerName,
				NamespaceID:   c.TracingPodNamespace,
				PodID:         c.TracingPodName,
				Zone:          zone,
			}
		}
		exporter, err := stackdriver.NewExporter(opts)
		if err != nil {
			return nil, errors.Wrap(err, "init stackdriver exporter")
		}
		exporters = append(exporters, exporter)
		trace.RegisterExporter(recoverExporter{exp: exporter})
	}

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(c.TraceProbability)})

	if c.LogTraces {
		e := c.wrapExporter(&logExporter{})
		exporters = append(exporters, e)
		trace.RegisterExporter(recoverExporter{exp: e})
	}

	return exporters, nil
}
