package azure

import (
	"context"
	"errors"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"

	liberror "github.com/ekristen/libnuke/pkg/errors"
)

// IsNotFound reports whether err is an Azure 404 response.
func IsNotFound(err error) bool {
	var respErr *azcore.ResponseError
	return errors.As(err, &respErr) && respErr.StatusCode == http.StatusNotFound
}

// IsConflict reports whether err is an Azure 409 response.
func IsConflict(err error) bool {
	var respErr *azcore.ResponseError
	return errors.As(err, &respErr) && respErr.StatusCode == http.StatusConflict
}

// IgnoreNotFound returns nil for an Azure 404 response: the resource is already gone.
func IgnoreNotFound(err error) error {
	if IsNotFound(err) {
		return nil
	}
	return err
}

// WaitForPoller polls a long-running delete operation once per libnuke wait cycle, returning ErrWaitResource while
// it is still running and the operation error if it failed.
func WaitForPoller[T any](ctx context.Context, poller *runtime.Poller[T]) error {
	if poller == nil {
		return nil
	}

	if _, err := poller.Poll(ctx); err != nil { //nolint:bodyclose // the poller owns the response, Result reads it
		return IgnoreNotFound(err)
	}

	if !poller.Done() {
		return liberror.ErrWaitResource("waiting for delete operation to complete")
	}

	_, err := poller.Result(ctx)
	return IgnoreNotFound(err)
}
