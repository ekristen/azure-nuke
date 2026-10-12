package azure

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/stretchr/testify/assert"
)

func TestIgnoreNotFound(t *testing.T) {
	notFound := &azcore.ResponseError{StatusCode: http.StatusNotFound}
	conflict := &azcore.ResponseError{StatusCode: http.StatusConflict}
	other := errors.New("boom")

	assert.NoError(t, IgnoreNotFound(nil))
	assert.NoError(t, IgnoreNotFound(notFound))
	assert.NoError(t, IgnoreNotFound(fmt.Errorf("wrapped: %w", notFound)))
	assert.Equal(t, conflict, IgnoreNotFound(conflict))
	assert.Equal(t, other, IgnoreNotFound(other))
}

func TestIsConflict(t *testing.T) {
	assert.True(t, IsConflict(&azcore.ResponseError{StatusCode: http.StatusConflict}))
	assert.False(t, IsConflict(&azcore.ResponseError{StatusCode: http.StatusNotFound}))
	assert.False(t, IsConflict(errors.New("boom")))
	assert.False(t, IsConflict(nil))
}

func TestWaitForPollerNil(t *testing.T) {
	var poller *runtime.Poller[struct{}]
	assert.NoError(t, WaitForPoller(context.Background(), poller))
}
