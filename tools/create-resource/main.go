package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/gertd/go-pluralize"
)

const resourceTemplate = `package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/{{.PackagePath}}"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const {{.ResourceType}}Resource = "{{.ResourceType}}"

func init() {
	registry.Register(&registry.Registration{
		Name:     {{.ResourceType}}Resource,
		Scope:    azure.ResourceGroupScope,
		Resource: &{{.ResourceType}}{},
		Lister:   &{{.ResourceType}}Lister{},
	})
}

type {{.ResourceType}}Lister struct{}

// NOTE: this is a jumping off point, check the client and pager names against the azure sdk for go package
func (l {{.ResourceType}}Lister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", {{.ResourceType}}Resource).WithField("s", opts.SubscriptionID)

	client, err := {{.Package}}.New{{.ResourceTypePlural}}Client(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list {{.ResourceTypePlural}}")

	pager := client.NewListPager(opts.ResourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entity := range page.Value {
			resources = append(resources, &{{.ResourceType}}{
				BaseResource: &BaseResource{
					Region:         entity.Location,
					ResourceGroup:  &opts.ResourceGroup,
					SubscriptionID: &opts.SubscriptionID,
				},
				client: client,
				Name:   entity.Name,
				Tags:   entity.Tags,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}

// ---------------------------------------------

type {{.ResourceType}} struct {
	*BaseResource ` + "`property:\",inline\"`" + `

	client   *{{.Package}}.{{.ResourceTypePlural}}Client
	removeOp *runtime.Poller[{{.Package}}.{{.ResourceTypePlural}}ClientDeleteResponse]
	Name     *string
	Tags     map[string]*string
}

func (r *{{.ResourceType}}) Remove(ctx context.Context) error {
	poller, err := r.client.BeginDelete(ctx, *r.ResourceGroup, *r.Name, nil)
	if err != nil {
		return err
	}

	r.removeOp = poller
	return nil
}

func (r *{{.ResourceType}}) HandleWait(ctx context.Context) error {
	return azure.WaitForPoller(ctx, r.removeOp)
}

func (r *{{.ResourceType}}) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *{{.ResourceType}}) String() string {
	return *r.Name
}
`

func main() {
	args := os.Args[1:]

	if len(args) != 2 {
		fmt.Println("usage: create-resource <package-path> <ResourceType>")
		fmt.Println("example: create-resource network/armnetwork/v12 VirtualNetwork")
		os.Exit(1)
	}

	packagePath := strings.Trim(args[0], "/")
	resourceType := args[1]

	pkg := ""
	for _, part := range strings.Split(packagePath, "/") {
		if strings.HasPrefix(part, "arm") {
			pkg = part
		}
	}
	if pkg == "" {
		pkg = path.Base(packagePath)
	}

	data := struct {
		PackagePath        string
		Package            string
		ResourceType       string
		ResourceTypePlural string
	}{
		PackagePath:        packagePath,
		Package:            pkg,
		ResourceType:       resourceType,
		ResourceTypePlural: pluralize.NewClient().Plural(resourceType),
	}

	tmpl, err := template.New("resource").Parse(resourceTemplate)
	if err != nil {
		panic(err)
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		panic(err)
	}

	fmt.Println(tpl.String())
}
