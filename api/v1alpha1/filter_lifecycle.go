package v1alpha1

import (
	"fmt"
	"github.com/vmware-labs/reconciler-runtime/apis"
)

const (
	FilterConditionReady                       = apis.ConditionReady
	FilterResourceMapping                      = "FilterResourceMapping"
	FilterResourceMappingNoSuchComponentReason = "NoSuchResource"
)

var containerCondSet = apis.NewLivingConditionSet(
	FilterResourceMapping,
)

func (b *FilteredRepositoryStatus) MarkResourceMissing(resource string, component string, namespace string) {
	template := "resource `%s` missing. " +
		"filter is trying to find resource `%s` in namespace `%s`"

	message := fmt.Sprintf(template, resource, component, namespace)

	containerCondSet.Manage(b).MarkFalse(FilterResourceMapping, FilterResourceMappingNoSuchComponentReason, message)
}

func (b *FilteredRepositoryStatus) MarkFailed(err error) {
	containerCondSet.Manage(b).MarkFalse(FilterConditionReady, "Failed", err.Error())
}
