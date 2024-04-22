package nuke

import (
	"fmt"
	"time"

	libnuke "github.com/ekristen/libnuke/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/utils"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

type Prompt struct {
	Parameters *libnuke.Parameters
	Tenant     *azure.Tenant
}

func (p *Prompt) Prompt() error {
	forceSleep := time.Duration(p.Parameters.ForceSleep) * time.Second

	fmt.Printf("Do you really want to nuke the tenant and subscriptions with "+
		"the ID %s?\n", p.Tenant.ID)
	if p.Parameters.Force {
		fmt.Printf("Waiting %v before continuing.\n", forceSleep)
		time.Sleep(forceSleep)
	} else {
		fmt.Printf("Do you want to continue? Enter tenant ID to continue.\n")
		if err := utils.Prompt(p.Tenant.ID); err != nil {
			return err
		}
	}

	return nil
}
