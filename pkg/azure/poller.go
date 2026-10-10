package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"

	liberror "github.com/ekristen/libnuke/pkg/errors"
)

// WaitForPoller polls a long-running delete operation once per libnuke wait cycle, returning ErrWaitResource while
// it is still running and the operation error if it failed.
func WaitForPoller[T any](ctx context.Context, poller *runtime.Poller[T]) error {
	if poller == nil {
		return nil
	}

	if _, err := poller.Poll(ctx); err != nil {
		return err
	}

	if !poller.Done() {
		return liberror.ErrWaitResource("waiting for delete operation to complete")
	}

	_, err := poller.Result(ctx)
	return err
}
