package v1alpha1

import (
	"fmt"
	"github.com/vmware-labs/reconciler-runtime/apis"
)

const (
	FilteredRepositoryConditionReady = apis.ConditionReady

	FilteredRepositorySourceMapping      = "FilteredRepositorySourceMapping"
	FilteredRepositoryNoSuchSourceReason = "NoSuchSource"

	FilteredRepositoryArtifactResolved       = "FilteredRepositoryArtifactResolved"
	FilteredRepositoryArtifactResolvedReason = "Resolved"

	FilteredRepositorySucceededReason = "Succeeded"
	FilteredRepositoryFailedReason    = "Failed"
)

var containerCondSet = apis.NewLivingConditionSet(
	FilteredRepositoryConditionReady,
	FilteredRepositorySourceMapping,
	FilteredRepositoryArtifactResolved,
)

func (b *FilteredRepositoryStatus) MarkResourceMissing(resource string, component string, namespace string) {
	template := "resource `%s` missing. " +
		"filter is trying to find resource `%s` in namespace `%s`"

	message := fmt.Sprintf(template, resource, component, namespace)

	containerCondSet.Manage(b).MarkFalse(FilteredRepositorySourceMapping, FilteredRepositoryNoSuchSourceReason, message)
}

func (b *FilteredRepositoryStatus) MarkArtifactResolved(url string) {
	template := "resolved artifact from url %s"

	message := fmt.Sprintf(template, url)

	containerCondSet.Manage(b).MarkFalse(FilteredRepositoryArtifactResolved, FilteredRepositoryArtifactResolvedReason, message)
}

func (b *FilteredRepositoryStatus) MarkFailed(err error) {
	containerCondSet.Manage(b).MarkFalse(FilteredRepositoryConditionReady, FilteredRepositoryFailedReason, err.Error())
}

func (b *FilteredRepositoryStatus) MarkReady() {
	containerCondSet.Manage(b).MarkTrue(FilteredRepositoryConditionReady, FilteredRepositorySucceededReason, "Repository has been successfully filtered for change")
}
