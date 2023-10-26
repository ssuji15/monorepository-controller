package v1alpha1

import (
	"context"
	"fmt"

	"github.com/vmware-labs/reconciler-runtime/apis"
)

const (
	MonoRepositoryConditionReady = apis.ConditionReady

	MonoRepositorySourceMapping      = "MonoRepositorySourceMapping"
	MonoRepositoryNoSuchSourceReason = "NoSuchSource"

	MonoRepositoryArtifactResolved       = "MonoRepositoryArtifactResolved"
	MonoRepositoryArtifactResolvedReason = "Resolved"

	MonoRepositorySucceededReason = "Succeeded"
	MonoRepositoryFailedReason    = "Failed"
)

var containerCondSet = apis.NewLivingConditionSet(
	MonoRepositoryConditionReady,
	MonoRepositorySourceMapping,
	MonoRepositoryArtifactResolved,
)

func (b *MonoRepositoryStatus) MarkResourceMissing(ctx context.Context, resource string, component string, namespace string) {
	template := "resource `%s` missing. " +
		"filter is trying to find resource `%s` in namespace `%s`"

	message := fmt.Sprintf(template, resource, component, namespace)

	containerCondSet.ManageWithContext(ctx, b).MarkFalse(MonoRepositorySourceMapping, MonoRepositoryNoSuchSourceReason, message)
}

func (b *MonoRepositoryStatus) MarkArtifactResolved(ctx context.Context, url string) {
	template := "resolved artifact from url %s"

	message := fmt.Sprintf(template, url)

	containerCondSet.ManageWithContext(ctx, b).MarkTrue(MonoRepositoryArtifactResolved, MonoRepositoryArtifactResolvedReason, message)
}

func (b *MonoRepositoryStatus) MarkFailed(ctx context.Context, err error) {
	containerCondSet.ManageWithContext(ctx, b).MarkFalse(MonoRepositoryConditionReady, MonoRepositoryFailedReason, err.Error())
}

func (b *MonoRepositoryStatus) MarkReady(ctx context.Context, checksum string) {
	containerCondSet.ManageWithContext(ctx, b).MarkTrue(MonoRepositoryConditionReady, MonoRepositorySucceededReason, "Repository has been successfully filtered with checksum %s", checksum)
}
