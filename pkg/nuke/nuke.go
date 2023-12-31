package nuke

import (
	"github.com/ekristen/azure-nuke/pkg/azure"
	sdknuke "github.com/ekristen/cloud-nuke-sdk/pkg/nuke"
)

type Nuke struct {
	sdknuke.Nuke
	Tenant         *azure.Tenant
	TenantId       string
	SubscriptionId string
}

func New(params sdknuke.Parameters, tenant *azure.Tenant) *Nuke {
	n := Nuke{
		Nuke: sdknuke.Nuke{
			Parameters: params,
		},
		Tenant: tenant,
	}

	return &n
}
