package resources

import (
	v1 "dies.dev/apis/meta/v1"
	"github.com/garethjevans/filter-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +die:object=true
type _ = v1alpha1.FilteredRepository

// +die
type _ = v1alpha1.FilteredRepositorySpec

// +die
type _ = v1alpha1.FilteredRepositoryStatus

func (d *FilteredRepositoryStatusDie) ConditionsDie(conditions ...*v1.ConditionDie) *FilteredRepositoryStatusDie {
	return d.DieStamp(func(r *v1alpha1.FilteredRepositoryStatus) {
		r.Conditions = make([]metav1.Condition, len(conditions))
		for i := range conditions {
			r.Conditions[i] = conditions[i].DieRelease()
		}
	})
}

var (
	FilterReadyBlank           = v1.ConditionBlank.Type(v1alpha1.FilteredRepositoryConditionReady)
	FilterResourceMappingBlank = v1.ConditionBlank.Type(v1alpha1.FilteredRepositorySourceMapping)
)
