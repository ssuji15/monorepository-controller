package v1alpha1

import (
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

func (b *MonoRepositoryStatus) MarkResourceMissing(resource string, component string, namespace string) {
	template := "resource `%s` missing. " +
		"filter is trying to find resource `%s` in namespace `%s`"

	message := fmt.Sprintf(template, resource, component, namespace)

	containerCondSet.Manage(b).MarkFalse(MonoRepositorySourceMapping, MonoRepositoryNoSuchSourceReason, message)
}

func (b *MonoRepositoryStatus) MarkArtifactResolved(url string) {
	template := "resolved artifact from url %s"

	message := fmt.Sprintf(template, url)

	containerCondSet.Manage(b).MarkTrue(MonoRepositoryArtifactResolved, MonoRepositoryArtifactResolvedReason, message)
}

func (b *MonoRepositoryStatus) MarkFailed(err error) {
	containerCondSet.Manage(b).MarkFalse(MonoRepositoryConditionReady, MonoRepositoryFailedReason, err.Error())
}

func (b *MonoRepositoryStatus) MarkReady() {
	containerCondSet.Manage(b).MarkTrue(MonoRepositoryConditionReady, MonoRepositorySucceededReason, "Repository has been successfully filtered for change")
}
