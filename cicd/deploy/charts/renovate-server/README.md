# renovate-server

`renovate-server` is a webhook server for `renovate`

## Introduction

This is the official helm chart for [renovate-server](https://github.com/arhat-dev/renovate-server), you can deploy `renovate-server` to your Kubernetes cluster to receive github/gitlab webhooks and schedule cron jobs to invoke renovate

## Prerequisites

- `helm` v3
- `Kubernetes` 1.15+

## Installing the Chart

```bash
helm install my-release arhat-dev/renovate-server
```

## Uninstalling the Chart

```bash
helm delete my-release
```

## Configuration

Please refer to the [`values.yaml`](https://github.com/arhat-dev/renovate-server/blob/v0.1.3/cicd/deploy/charts/renovate-server/values.yaml)
