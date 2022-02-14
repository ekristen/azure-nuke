package resource

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/ekristen/azure-nuke/pkg/azure"
	"github.com/ekristen/azure-nuke/pkg/config"
	"github.com/ekristen/azure-nuke/pkg/types"
)

type ResourceListers map[string]ResourceLister

type ResourceLister func(authorizers azure.Authorizers, subscriptionId, resourceGroup string) ([]Resource, error)

type Resource interface {
	Remove() error
}

type Filter interface {
	Resource
	Filter() error
}

type LegacyStringer interface {
	Resource
	String() string
}

type ResourcePropertyGetter interface {
	Resource
	Properties() types.Properties
}

type FeatureFlagGetter interface {
	Resource
	FeatureFlags(config.FeatureFlags)
}

var resourceListers = make(ResourceListers)

func Register(name string, lister ResourceLister) {
	_, exists := resourceListers[name]
	if exists {
		panic(fmt.Sprintf("a resource with the name %s already exists", name))
	}

	logrus.WithField("name", name).Trace("registered listered")

	resourceListers[name] = lister
}

func GetListers() ResourceListers {
	return resourceListers
}

func GetLister(name string) ResourceLister {
	return resourceListers[name]
}

func GetListerNames() []string {
	names := []string{}
	for resourceType := range GetListers() {
		names = append(names, resourceType)
	}

	return names
}
