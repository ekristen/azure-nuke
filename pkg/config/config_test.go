package config

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/ekristen/azure-nuke/pkg/types"
)

func TestConfigBlocklist(t *testing.T) {
	config := new(Nuke)

	if config.HasBlocklist() {
		t.Errorf("HasBlocklist() returned true on a nil backlist.")
	}

	if config.InBlocklist("blubber") {
		t.Errorf("InBlocklist() returned true on a nil backlist.")
	}

	config.TenantBlocklist = []string{}

	if config.HasBlocklist() {
		t.Errorf("HasBlocklist() returned true on a empty backlist.")
	}

	if config.InBlocklist("foobar") {
		t.Errorf("InBlocklist() returned true on a empty backlist.")
	}

	config.TenantBlocklist = append(config.TenantBlocklist, "bim")

	if !config.HasBlocklist() {
		t.Errorf("HasBlocklist() returned false on a backlist with one element.")
	}

	if !config.InBlocklist("bim") {
		t.Errorf("InBlocklist() returned false on looking up an existing value.")
	}

	if config.InBlocklist("baz") {
		t.Errorf("InBlocklist() returned true on looking up an non existing value.")
	}
}

func TestLoadExampleConfig(t *testing.T) {
	config, err := Load("testdata/example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	expect := Nuke{
		TenantBlocklist: []string{"1234567890"},
		Tenants: map[string]Tenant{
			"555133742": {
				Presets: []string{"terraform"},
				Filters: Filters{
					"IAMRole": {
						NewExactFilter("uber.admin"),
					},
					"IAMRolePolicyAttachment": {
						NewExactFilter("uber.admin -> AdministratorAccess"),
					},
				},
				ResourceTypes: ResourceTypes{
					types.Collection{"S3Bucket"},
					nil,
				},
			},
		},
		ResourceTypes: ResourceTypes{
			Targets:  types.Collection{"DynamoDBTable", "S3Bucket", "S3Object"},
			Excludes: types.Collection{"IAMRole"},
		},
		Presets: map[string]PresetDefinitions{
			"terraform": {
				Filters: Filters{
					"S3Bucket": {
						Filter{
							Type:  FilterTypeGlob,
							Value: "my-statebucket-*",
						},
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(*config, expect) {
		t.Errorf("Read struct mismatches:")
		t.Errorf("  Got:      %#v", *config)
		t.Errorf("  Expected: %#v", expect)
	}
}

func TestResolveDeprecations(t *testing.T) {
	config := Nuke{
		TenantBlocklist: []string{"1234567890"},
		Tenants: map[string]Tenant{
			"555133742": {
				Filters: Filters{
					"IamRole": {
						NewExactFilter("uber.admin"),
						NewExactFilter("foo.bar"),
					},
					"IAMRolePolicyAttachment": {
						NewExactFilter("uber.admin -> AdministratorAccess"),
					},
				},
			},
			"2345678901": {
				Filters: Filters{
					"ECRrepository": {
						NewExactFilter("foo:bar"),
						NewExactFilter("bar:foo"),
					},
					"IAMRolePolicyAttachment": {
						NewExactFilter("uber.admin -> AdministratorAccess"),
					},
				},
			},
		},
	}

	expect := map[string]Tenant{
		"555133742": {
			Filters: Filters{
				"IAMRole": {
					NewExactFilter("uber.admin"),
					NewExactFilter("foo.bar"),
				},
				"IAMRolePolicyAttachment": {
					NewExactFilter("uber.admin -> AdministratorAccess"),
				},
			},
		},
		"2345678901": {
			Filters: Filters{
				"ECRRepository": {
					NewExactFilter("foo:bar"),
					NewExactFilter("bar:foo"),
				},
				"IAMRolePolicyAttachment": {
					NewExactFilter("uber.admin -> AdministratorAccess"),
				},
			},
		},
	}

	err := config.resolveDeprecations()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(config.Tenants, expect) {
		t.Errorf("Read struct mismatches:")
		t.Errorf("  Got:      %#v", config.Tenants)
		t.Errorf("  Expected: %#v", expect)
	}

	invalidConfig := Nuke{
		TenantBlocklist: []string{"1234567890"},
		Tenants: map[string]Tenant{
			"555133742": {
				Filters: Filters{
					"IamUserAccessKeys": {
						NewExactFilter("X")},
					"IAMUserAccessKey": {
						NewExactFilter("Y")},
				},
			},
		},
	}

	err = invalidConfig.resolveDeprecations()
	if err == nil || !strings.Contains(err.Error(), "using deprecated resource type and replacement") {
		t.Fatal("invalid config did not cause correct error")
	}
}

func TestConfigValidation(t *testing.T) {
	config, err := Load("testdata/example.yaml")
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
			err := config.ValidateTenant(tc.ID)
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
	config, err := Load("testdata/deprecated-keys-config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if !config.InBlocklist("1234567890") {
		t.Errorf("Loading the config did not resolve the deprecated key 'account-blacklist' correctly")
	}
}

func TestFilterMerge(t *testing.T) {
	config, err := Load("testdata/example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	filters, err := config.Filters("555133742")
	if err != nil {
		t.Fatal(err)
	}

	expect := Filters{
		"S3Bucket": []Filter{
			{
				Type: "glob", Value: "my-statebucket-*",
			},
		},
		"IAMRole": []Filter{
			{
				Type:  "exact",
				Value: "uber.admin",
			},
		},
		"IAMRolePolicyAttachment": []Filter{
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
