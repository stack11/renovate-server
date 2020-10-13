/*
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
*/

package conf

import (
	"arhat.dev/pkg/confhelper"
	"arhat.dev/pkg/log"
	"github.com/spf13/pflag"

	"arhat.dev/renovate-server/pkg/constant"
)

type Config struct {
	Server ServerConfig `json:"server" yaml:"server"`

	GitHub []GitHubConfig `json:"github" yaml:"github"`
	GitLab []GitLabConfig `json:"gitlab" yaml:"gitlab"`
}

type ServerConfig struct {
	Log log.ConfigSet `json:"log" yaml:"log"`

	Webhook struct {
		Listen string               `json:"listen" yaml:"listen"`
		TLS    confhelper.TLSConfig `json:"tls" yaml:"tls"`
	} `json:"webhook" yaml:"webhook"`
}

func FlagsForServer(prefix string, config *ServerConfig) *pflag.FlagSet {
	fs := pflag.NewFlagSet("app", pflag.ExitOnError)

	fs.StringVar(&config.Webhook.Listen, prefix+"webhook.listen",
		constant.DefaultWebhookListenAddress, "set webhook listener address",
	)
	fs.AddFlagSet(confhelper.FlagsForTLSConfig(prefix+"webhook.tls", &config.Webhook.TLS))

	return fs
}
