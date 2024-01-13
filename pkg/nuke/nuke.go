package nuke

import (
	"github.com/ekristen/azure-nuke/pkg/azure"
	"github.com/ekristen/azure-nuke/pkg/config"
	sdknuke "github.com/ekristen/libnuke/pkg/nuke"
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

func New(params Parameters, tenant *azure.Tenant) *Nuke {
	n := Nuke{
		Nuke: sdknuke.Nuke{
			Parameters: params.Parameters,
		},
		Parameters: params,
		Tenant:     tenant,
	}

	return &n
}
