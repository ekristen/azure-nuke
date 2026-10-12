package azure

import (
	"fmt"
	"strings"
	"time"

	libnuke "github.com/ekristen/libnuke/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/utils"
)

// Prompt is a struct that contains the parameters and tenant details use to craft a unique prompt
// for the user to confirm the nuke operation.
type Prompt struct {
	Parameters *libnuke.Parameters
	Tenant     *Tenant

	// TenantResourceTypes are the tenant-scoped (Entra ID) resource types the run includes
	TenantResourceTypes []string
}

func (p *Prompt) Prompt() error {
	forceSleep := time.Duration(p.Parameters.ForceSleep) * time.Second

	fmt.Printf("Do you really want to nuke these subscriptions of tenant %s?\n", Describe(p.Tenant.ID, p.Tenant.Name))

	subscriptions := make([]string, 0, len(p.Tenant.SubscriptionIds))
	for _, id := range p.Tenant.SubscriptionIds {
		subscriptions = append(subscriptions, Describe(id, p.Tenant.SubscriptionNames[id]))
	}
	fmt.Printf("Subscriptions: %s\n", strings.Join(subscriptions, ", "))

	if len(p.TenantResourceTypes) == 0 {
		fmt.Println("Tenant resources (Entra ID): none, all tenant resource types are excluded")
	} else {
		fmt.Printf("Tenant resources (Entra ID): %s\n", strings.Join(p.TenantResourceTypes, ", "))
	}
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
