This repository is the repackaging of [node_exporter](https://github.com/prometheus/node_exporter) as an *Opentelemetry Collector Receiver*.

(Despite the name, it's a receiver)

Opentelemetry Collector already has a component to scrape node metrics. First of all, I recommend to check [hostmetrics](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/README.md).

`node_exporter` receiver is intended to be used when you want to reuse existing dashboards and alerting rules built for Prometheus `node_exporter` (or just need one of the collectors).

Unfortunately it's not easy to reuse `node_exporter` because the usage of global variables and flag parsing.

Grafana Alloy had a similar problem, and they created a fork: https://github.com/prometheus/node_exporter/pull/2812

Here, I use different approach. The collector directory is vanilla `node_exporter`, only the import path of the `kingpin` library is replaced.

That can make it easer to keep up to date with the original `node_exporter`.

Can be configured like the original node exporter using the `flags` field in configuration.

Example:

```yaml
receivers:
  nodeexporter:
    interval: 10s
    flags:
      collector.cpu: "false"
  
...

service:
  pipelines:
    metrics:
      receivers: [nodeexporter]
      exporters: [debug]
```

This is just a repackage. Kudos for the original contributors (can be found in the current Git history).