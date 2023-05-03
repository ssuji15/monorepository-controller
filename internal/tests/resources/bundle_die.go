package resources

import (
	v1 "dies.dev/apis/meta/v1"
	"github.com/garethjevans/filter-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +die:object=true
type _ = v1alpha1.Filter

// +die
type _ = v1alpha1.FilterSpec

// +die
type _ = v1alpha1.FilterStatus

func (d *FilterStatusDie) ConditionsDie(conditions ...*v1.ConditionDie) *FilterStatusDie {
	return d.DieStamp(func(r *v1alpha1.FilterStatus) {
		r.Conditions = make([]metav1.Condition, len(conditions))
		for i := range conditions {
			r.Conditions[i] = conditions[i].DieRelease()
		}
	})
}

var (
	FilterReadyBlank           = v1.ConditionBlank.Type(v1alpha1.FilterConditionReady)
	FilterResourceMappingBlank = v1.ConditionBlank.Type(v1alpha1.FilterResourceMapping)
)
