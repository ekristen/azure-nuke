package nuke

import (
	"github.com/ekristen/azure-nuke/pkg/azure"
	"github.com/ekristen/azure-nuke/pkg/config"
	"github.com/ekristen/libnuke/pkg/featureflag"
	"github.com/ekristen/libnuke/pkg/filter"
	sdknuke "github.com/ekristen/libnuke/pkg/nuke"
	"github.com/sirupsen/logrus"
)

type Parameters struct {
	sdknuke.Parameters

	Targets      []string
	Excludes     []string
	CloudControl []string
}

type Nuke struct {
	sdknuke.Nuke
	Parameters     Parameters
	Config         *config.Nuke
	Tenant         *azure.Tenant
	TenantId       string
	SubscriptionId string
}

func New(params Parameters, config *config.Nuke, filters filter.Filters, tenant *azure.Tenant) *Nuke {
	n := Nuke{
		Nuke: sdknuke.Nuke{
			Parameters:   params.Parameters,
			FeatureFlags: &featureflag.FeatureFlags{},
			Filters:      filters,
		},
		Parameters: params,
		Config:     config,
		Tenant:     tenant,
	}

	n.SetLogger(logrus.WithField("component", "nuke"))

	return &n
}
