package v1alpha1

import (
	"context"

	"github.com/vmware-labs/reconciler-runtime/apis"
)

const (
	MonoRepositoryConditionReady = apis.ConditionReady

	MonoRepositorySucceededReason = "Succeeded"
	MonoRepositoryFailedReason    = "Failed"
)

var containerCondSet = apis.NewLivingConditionSet(
	MonoRepositoryConditionReady,
)

func (b *MonoRepositoryStatus) MarkFailed(ctx context.Context, err error) {
	containerCondSet.ManageWithContext(ctx, b).MarkFalse(MonoRepositoryConditionReady, MonoRepositoryFailedReason, err.Error())
}

func (b *MonoRepositoryStatus) MarkReady(ctx context.Context, checksum string) {
	containerCondSet.ManageWithContext(ctx, b).MarkTrue(MonoRepositoryConditionReady, MonoRepositorySucceededReason, "Repository has been successfully filtered with checksum %s", checksum)
}
