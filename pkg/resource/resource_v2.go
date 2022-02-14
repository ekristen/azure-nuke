package resource

import (
	"fmt"

	"github.com/ekristen/azure-nuke/pkg/azure"
	"github.com/sirupsen/logrus"
	"github.com/stevenle/topsort"
)

type Scope string

const (
	Tenant        Scope = "tenant"
	Subscription  Scope = "subscription"
	ResourceGroup Scope = "resource-group"
)

type Registrations map[string]Registration
type ResourceListersV2 map[string]ResourceListerV2

var resourceListersV2 = make(ResourceListersV2)
var registrations = make(Registrations)

type Registration struct {
	Name      string
	Scope     Scope
	Lister    ResourceListerV2
	DependsOn []string
}

type ListerOpts struct {
	Authorizers    azure.Authorizers
	TenantId       string
	SubscriptionId string
	ResourceGroup  string
}

type ResourceListerV2 func(lister ListerOpts) ([]Resource, error)

func RegisterV2(r Registration) {
	if r.Scope == "" {
		panic(fmt.Errorf("scope must be set"))
	}

	_, exists := registrations[r.Name]
	if exists {
		panic(fmt.Sprintf("a resource with the name %s already exists", r.Name))
	}

	logrus.WithField("name", r.Name).Trace("registered listered")

	registrations[r.Name] = r
	resourceListersV2[r.Name] = r.Lister
}

func GetListersV2() (listers ResourceListersV2) {
	listers = make(ResourceListersV2)
	for name, r := range registrations {
		listers[name] = r.Lister
	}
	return listers
}

func GetListersTS() {
	graph := topsort.NewGraph()

	for name := range registrations {
		graph.AddNode(name)
	}
	for name, r := range registrations {
		for _, dep := range r.DependsOn {
			graph.AddEdge(name, dep)
		}
	}
}

func GetListersForScope(scope Scope) (listers ResourceListersV2) {
	listers = make(ResourceListersV2)
	for name, r := range registrations {
		if r.Scope == scope {
			listers[name] = r.Lister
		}
	}
	return listers
}

func GetListerNamesV2() []string {
	names := []string{}
	for resourceType := range GetListersV2() {
		names = append(names, resourceType)
	}

	return names
}

func GetListersNameForScope(scope Scope) []string {
	names := []string{}
	for resourceType := range GetListersForScope(scope) {
		names = append(names, resourceType)
	}
	return names
}

func GetListerV2(name string) ResourceListerV2 {
	return resourceListersV2[name]
}
