# renovate-server

[![CI](https://github.com/arhat-dev/renovate-server/workflows/CI/badge.svg)](https://github.com/arhat-dev/renovate-server/actions?query=workflow%3ACI)
[![Build](https://github.com/arhat-dev/renovate-server/workflows/Build/badge.svg)](https://github.com/arhat-dev/renovate-server/actions?query=workflow%3ABuild)
[![PkgGoDev](https://pkg.go.dev/badge/arhat.dev/renovate-server)](https://pkg.go.dev/arhat.dev/renovate-server)
[![GoReportCard](https://goreportcard.com/badge/arhat.dev/renovate-server)](https://goreportcard.com/report/arhat.dev/renovate-server)
[![codecov](https://codecov.io/gh/arhat-dev/renovate-server/branch/master/graph/badge.svg)](https://codecov.io/gh/arhat-dev/renovate-server)

Self-Hosted renovate server to automate renovate actions

## Support Matrix

- Job Types
  - Cron Job
  - Webhook Events
    - `issue` with dashboard title: edited/closed/reopened
    - `pull/merge request` with checkbox edited/closed/reopened
    - `push`
- Platforms
  - `gitlab`
  - `github`
- Executors
  - `kubernetes` (creates kubernetes jobs to execute renovate)

## Usage

1. Create a `renovate` config in your repository, you can use [shareable config presets](https://docs.renovatebot.com/config-presets/) to save your time
   - see [`arhat-dev/renovate-presets`](https://github.com/arhat-dev/renovate-presets) for example

2. Deploy `renovate-server` to your local/cloud environment, then you will get a webhook endpoint exposed via your ingress controller
   - for kubernetes, you can customize your installation with [helm chart](./cicd/deploy/charts/renovate-server)

3. Configure your repository or organization, create a webhook for `renovate-server` with desired events, say `issues`, `pull requests` and `push`

4. Now you are good to go, every time you trigger the webhook with desired event payload, `renovate-server` will execute renovate for you

## LICENSE

```text
Copyright 2020 The arhat.dev Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
