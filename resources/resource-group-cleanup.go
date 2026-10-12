package resources

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armlocks"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azkeys"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const (
	resourceGroupCleanupTimeout = 5 * time.Minute
	deletedKeyPollInterval      = 2 * time.Second
)

// deleting reports whether the resource group delete has already started.
func (r *ResourceGroup) deleting(ctx context.Context) (bool, error) {
	group, err := r.client.Get(ctx, *r.Name, nil)
	if err != nil {
		return false, err
	}

	return group.Properties != nil && strings.EqualFold(ptr.ToString(group.Properties.ProvisioningState), "Deleting"), nil
}

// softDeleteKeys soft-deletes (never purges) the keys of the resource group's vaults whose tags match the
// SoftDeleteKeyVaultKeysTag setting (`key=value`, every vault when empty) and waits until Key Vault reports them
// deleted. A purge-protected vault with a fixed name is recovered by the next build, with its keys still live unless
// they were soft-deleted first.
func (r *ResourceGroup) softDeleteKeys(ctx context.Context) error {
	tagKey, tagValue := parseTag(r.settings.GetString("SoftDeleteKeyVaultKeysTag"))

	client, err := armkeyvault.NewVaultsClient(*r.SubscriptionID, r.creds, nil)
	if err != nil {
		return err
	}

	pager := client.NewListByResourceGroupPager(*r.Name, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return azure.IgnoreNotFound(err)
		}

		for _, vault := range page.Value {
			if !tagMatches(vault.Tags, tagKey, tagValue) || vault.Properties == nil || vault.Properties.VaultURI == nil {
				continue
			}
			// without soft delete a key delete is permanent
			if vault.Properties.EnableSoftDelete != nil && !*vault.Properties.EnableSoftDelete {
				logrus.WithField("r", ResourceGroupResource).Warnf("vault %s has soft delete disabled, keys left alone", *vault.Name)
				continue
			}

			if err := r.softDeleteVaultKeys(ctx, *vault.Name, *vault.Properties.VaultURI); err != nil {
				return fmt.Errorf("vault %s: %w", *vault.Name, err)
			}
		}
	}

	return nil
}

func (r *ResourceGroup) softDeleteVaultKeys(ctx context.Context, vaultName, vaultURI string) error {
	client, err := azkeys.NewClient(vaultURI, r.creds, nil)
	if err != nil {
		return err
	}

	names := make([]string, 0)
	pager := client.NewListKeyPropertiesPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return azure.IgnoreNotFound(err)
		}

		for _, key := range page.Value {
			// keys backing a certificate are managed by Key Vault and go with the certificate
			if key.KID == nil || (key.Managed != nil && *key.Managed) {
				continue
			}
			names = append(names, key.KID.Name())
		}
	}

	deleted := make([]string, 0, len(names))
	for _, name := range names {
		logrus.WithField("r", ResourceGroupResource).Infof("soft-deleting key %s/%s", vaultName, name)
		_, err := client.DeleteKey(ctx, name, nil)
		if azure.IsNotFound(err) {
			continue
		}
		// 409: a delete of this key is already in progress
		if err != nil && !azure.IsConflict(err) {
			return err
		}
		deleted = append(deleted, name)
	}

	for _, name := range deleted {
		if err := waitForDeletedKey(ctx, client, name); err != nil {
			return err
		}
	}

	return nil
}

// waitForDeletedKey waits until a key's soft delete has completed: it is listed as deleted only then.
func waitForDeletedKey(ctx context.Context, client *azkeys.Client, name string) error {
	for {
		_, err := client.GetDeletedKey(ctx, name, nil)
		if err == nil {
			return nil
		}
		if !azure.IsNotFound(err) {
			return err
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for key %s to be deleted: %w", name, ctx.Err())
		case <-time.After(deletedKeyPollInterval):
		}
	}
}

// deleteLocks deletes every management lock in the resource group, including locks on its resources.
func (r *ResourceGroup) deleteLocks(ctx context.Context) error {
	client, err := armlocks.NewManagementLocksClient(*r.SubscriptionID, r.creds, nil)
	if err != nil {
		return err
	}

	rgScope := strings.ToLower("/subscriptions/" + *r.SubscriptionID + "/resourceGroups/" + *r.Name)

	pager := client.NewListAtResourceGroupLevelPager(*r.Name, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}

		for _, lock := range page.Value {
			if lock.ID == nil || lock.Name == nil {
				continue
			}
			// only the resource group's own locks and its resources' locks, never inherited ones
			scope := strings.ToLower(lockScope(*lock.ID))
			if scope != rgScope && !strings.HasPrefix(scope, rgScope+"/") {
				continue
			}

			logrus.WithField("r", ResourceGroupResource).Infof("deleting lock %s", *lock.ID)
			if _, err := client.DeleteByScope(ctx, lockScope(*lock.ID), *lock.Name, nil); !azure.IsNotFound(err) && err != nil {
				return err
			}
		}
	}

	return nil
}

// parseTag splits a `key=value` setting; a value-less `key` matches any value.
func parseTag(setting string) (key, value string) {
	key, value, _ = strings.Cut(setting, "=")
	return strings.TrimSpace(key), strings.TrimSpace(value)
}

func tagMatches(tags map[string]*string, key, value string) bool {
	if key == "" {
		return true
	}

	v, ok := tags[key]
	if !ok || v == nil {
		return false
	}

	return value == "" || *v == value
}

// lockScope returns the scope a lock is attached to: its ID without the lock part.
func lockScope(lockID string) string {
	if i := strings.Index(strings.ToLower(lockID), "/providers/microsoft.authorization/locks/"); i > 0 {
		return lockID[:i]
	}
	return lockID
}
