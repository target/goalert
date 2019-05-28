package app

import "go.opencensus.io/trace"

type clusterExporter struct {
	*appConfig
	e trace.Exporter
}

func (c *appConfig) wrapExporter(e trace.Exporter) trace.Exporter {
	if c.TracingClusterName == "" {
		return e
	}
	return &clusterExporter{
		appConfig: c,
		e:         e,
	}
}
func (c *clusterExporter) Flush() {
	type flusher interface {
		Flush()
	}
	if f, ok := c.e.(flusher); ok {
		f.Flush()
	}
}
func (c *clusterExporter) ExportSpan(s *trace.SpanData) {
	if s.Attributes == nil {
		s.Attributes = make(map[string]interface{}, 5)
	}
	s.Attributes["cluster_name"] = c.TracingClusterName
	s.Attributes["namespace_id"] = c.TracingPodNamespace
	s.Attributes["pod_id"] = c.TracingPodName
	s.Attributes["container_name"] = c.TracingContainerName
	s.Attributes["instance_id"] = c.TracingNodeName

	c.e.ExportSpan(s)
}
