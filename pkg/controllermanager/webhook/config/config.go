/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved.
 * This file is licensed under the Apache Software License, v. 2 except as noted
 * otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *
 */

package config

import (
	"fmt"
	"strings"

	"github.com/gardener/controller-manager-library/pkg/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cert"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	areacfg "github.com/gardener/controller-manager-library/pkg/controllermanager/config"
)

const OPTION_SOURCE = "webhooks"

type Config struct {
	cert.CertConfig
	Webhooks               string
	Cluster                string
	Port                   int
	ServicePort            int
	RegistrationName       string
	DedicatedRegistrations bool
	OmitRegistrations      bool
	Labels                 []string

	config.OptionSet
}

var _ config.OptionSource = (*Config)(nil)

func NewConfig() *Config {
	cfg := &Config{
		CertConfig: *cert.NewCertConfig("webhook", ""),
		OptionSet:  config.NewDefaultOptionSet(OPTION_SOURCE, OPTION_SOURCE),
	}
	cfg.CertConfig.AddOptionsToSet(cfg.OptionSet)
	cfg.AddStringOption(&cfg.Webhooks, "webhooks", "w", "all", "comma separated list of webhooks to start (<name>,<group>,all)")
	cfg.AddStringOption(&cfg.Cluster, "cluster", "", cluster.DEFAULT, "cluster to maintain webhook server secret")
	cfg.AddIntOption(&cfg.Port, "port", "", 8443, "port to use for webhook server")
	cfg.AddIntOption(&cfg.ServicePort, "service-port", "", 443, "port used on service")
	cfg.AddStringOption(&cfg.RegistrationName, "registration-name", "", "", "webhook registration name for grouped registrations")
	cfg.AddBoolOption(&cfg.OmitRegistrations, "omit-webhook-registration", "", false, "omit webhook registration")
	cfg.AddBoolOption(&cfg.DedicatedRegistrations, "dedicated-webhook-registrations", "", false, "uses separate registrations for every configured webhook")
	cfg.AddStringArrayOption(&cfg.Labels, "label", "", nil, "additional labels for the webhook registrations")
	return cfg
}

func (this *Config) AddOptionsToSet(set config.OptionSet) {
	this.OptionSet.AddOptionsToSet(set)
}

func (this *Config) Evaluate() error {
	if this.Secret != "" {
		if this.Cluster == "" {
			return fmt.Errorf("web hook cluster name must be specified for automated secret maintenance")
		}
	}
	if len(this.Hostnames) == 0 && this.Service == "" {
		return fmt.Errorf("web hook server requires at least service name or hostname")
	}

	if len(this.Hostnames) > 1 {
		return fmt.Errorf("web hook server requires only one hostname")
	}
	if this.CertFile != "" && this.KeyFile == "" {
		return fmt.Errorf("specifying webhook server certficate require key file, also")
	}
	if this.Secret != "" && this.CertFile != "" {
		return fmt.Errorf("only one of webhook server certificate file or secret name possible")
	}
	if this.Secret == "" && this.CertFile == "" {
		return fmt.Errorf("one of webhook server certificate file or secret name must be specified")
	}
	for _, l := range this.Labels {
		a := strings.Split(l, "=")
		if len(a) != 2 {
			return fmt.Errorf("invalid label spec (%s): must contain excactly one = character")
		}
	}
	return this.OptionSet.Evaluate()
}

func GetConfig(cfg *areacfg.Config) *Config {
	return cfg.GetSource(OPTION_SOURCE).(*Config)
}
