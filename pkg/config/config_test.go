package config

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/ekristen/libnuke/pkg/filter"
	"github.com/ekristen/libnuke/pkg/types"
)

func TestConfigBlocklist(t *testing.T) {
	cfg := new(Nuke)

	if cfg.HasBlocklist() {
		t.Errorf("HasBlocklist() returned true on a nil backlist.")
	}

	if cfg.InBlocklist("blubber") {
		t.Errorf("InBlocklist() returned true on a nil backlist.")
	}

	cfg.TenantBlocklist = []string{}

	if cfg.HasBlocklist() {
		t.Errorf("HasBlocklist() returned true on a empty backlist.")
	}

	if cfg.InBlocklist("foobar") {
		t.Errorf("InBlocklist() returned true on a empty backlist.")
	}

	cfg.TenantBlocklist = append(cfg.TenantBlocklist, "bim")

	if !cfg.HasBlocklist() {
		t.Errorf("HasBlocklist() returned false on a backlist with one element.")
	}

	if !cfg.InBlocklist("bim") {
		t.Errorf("InBlocklist() returned false on looking up an existing value.")
	}

	if cfg.InBlocklist("baz") {
		t.Errorf("InBlocklist() returned true on looking up an non existing value.")
	}
}

func TestLoadExampleConfig(t *testing.T) {
	cfg, err := Load("testdata/example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	expect := Nuke{
		TenantBlocklist: []string{"1234567890"},
		Tenants: map[string]Tenant{
			"555133742": {
				Presets: []string{"terraform"},
				Filters: filter.Filters{
					"IAMRole": {
						filter.NewExactFilter("uber.admin"),
					},
					"IAMRolePolicyAttachment": {
						filter.NewExactFilter("uber.admin -> AdministratorAccess"),
					},
				},
				ResourceTypes: ResourceTypes{
					Targets:  types.Collection{"S3Bucket"},
					Excludes: nil,
				},
			},
		},
		ResourceTypes: ResourceTypes{
			Targets:  types.Collection{"DynamoDBTable", "S3Bucket", "S3Object"},
			Excludes: types.Collection{"IAMRole"},
		},
		Presets: map[string]PresetDefinitions{
			"terraform": {
				Filters: filter.Filters{
					"S3Bucket": {
						filter.Filter{
							Type:  filter.Glob,
							Value: "my-statebucket-*",
						},
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(cfg, expect) {
		t.Errorf("Read struct mismatches:")
		t.Errorf("  Got:      %#v", cfg)
		t.Errorf("  Expected: %#v", expect)
	}
}

func TestResolveDeprecations(t *testing.T) {
	cfg := Nuke{
		TenantBlocklist: []string{"1234567890"},
		Tenants: map[string]Tenant{
			"555133742": {
				Filters: filter.Filters{
					"IamRole": {
						filter.NewExactFilter("uber.admin"),
						filter.NewExactFilter("foo.bar"),
					},
					"IAMRolePolicyAttachment": {
						filter.NewExactFilter("uber.admin -> AdministratorAccess"),
					},
				},
			},
			"2345678901": {
				Filters: filter.Filters{
					"ECRrepository": {
						filter.NewExactFilter("foo:bar"),
						filter.NewExactFilter("bar:foo"),
					},
					"IAMRolePolicyAttachment": {
						filter.NewExactFilter("uber.admin -> AdministratorAccess"),
					},
				},
			},
		},
	}

	expect := map[string]Tenant{
		"555133742": {
			Filters: filter.Filters{
				"IAMRole": {
					filter.NewExactFilter("uber.admin"),
					filter.NewExactFilter("foo.bar"),
				},
				"IAMRolePolicyAttachment": {
					filter.NewExactFilter("uber.admin -> AdministratorAccess"),
				},
			},
		},
		"2345678901": {
			Filters: filter.Filters{
				"ECRRepository": {
					filter.NewExactFilter("foo:bar"),
					filter.NewExactFilter("bar:foo"),
				},
				"IAMRolePolicyAttachment": {
					filter.NewExactFilter("uber.admin -> AdministratorAccess"),
				},
			},
		},
	}

	err := cfg.ResolveDeprecations()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg.Tenants, expect) {
		t.Errorf("Read struct mismatches:")
		t.Errorf("  Got:      %#v", cfg.Tenants)
		t.Errorf("  Expected: %#v", expect)
	}

	invalidConfig := Nuke{
		TenantBlocklist: []string{"1234567890"},
		Tenants: map[string]Tenant{
			"555133742": {
				Filters: filter.Filters{
					"IamUserAccessKeys": {
						filter.NewExactFilter("X")},
					"IAMUserAccessKey": {
						filter.NewExactFilter("Y")},
				},
			},
		},
	}

	err = invalidConfig.ResolveDeprecations()
	if err == nil || !strings.Contains(err.Error(), "using deprecated resource type and replacement") {
		t.Fatal("invalid config did not cause correct error")
	}
}

func TestConfigValidation(t *testing.T) {
	cfg, err := Load("testdata/example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		ID         string
		ShouldFail bool
	}{
		{ID: "555133742", ShouldFail: false},
		{ID: "1234567890", ShouldFail: true},
		{ID: "1111111111", ShouldFail: true},
	}

	for i, tc := range cases {
		name := fmt.Sprintf("%d_%s/%t", i, tc.ID, tc.ShouldFail)
		t.Run(name, func(t *testing.T) {
			err := cfg.Validate(tc.ID)
			if tc.ShouldFail && err == nil {
				t.Fatal("Expected an error but didn't get one.")
			}
			if !tc.ShouldFail && err != nil {
				t.Fatalf("Didn't excpect an error, but got one: %v", err)
			}
		})
	}
}

func TestDeprecatedConfigKeys(t *testing.T) {
	cfg, err := Load("testdata/deprecated-keys-config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if !cfg.InBlocklist("1234567890") {
		t.Errorf("Loading the config did not resolve the deprecated key 'account-blacklist' correctly")
	}
}

func TestFilterMerge(t *testing.T) {
	cfg, err := Load("testdata/example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	filters, err := cfg.Filters("555133742")
	if err != nil {
		t.Fatal(err)
	}

	expect := filter.Filters{
		"S3Bucket": []filter.Filter{
			{
				Type: "glob", Value: "my-statebucket-*",
			},
		},
		"IAMRole": []filter.Filter{
			{
				Type:  "exact",
				Value: "uber.admin",
			},
		},
		"IAMRolePolicyAttachment": []filter.Filter{
			{
				Type:  "exact",
				Value: "uber.admin -> AdministratorAccess",
			},
		},
	}

	if !reflect.DeepEqual(filters, expect) {
		t.Errorf("Read struct mismatches:")
		t.Errorf("  Got:      %#v", filters)
		t.Errorf("  Expected: %#v", expect)
	}
}
