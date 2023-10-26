package resources

import (
	v1 "dies.dev/apis/meta/v1"
	"github.com/garethjevans/monorepository-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +die:object=true
type _ = v1alpha1.MonoRepository

// +die
type _ = v1alpha1.MonoRepositorySpec

// +die
type _ = v1alpha1.MonoRepositoryStatus

func (d *MonoRepositoryStatusDie) ConditionsDie(conditions ...*v1.ConditionDie) *MonoRepositoryStatusDie {
	return d.DieStamp(func(r *v1alpha1.MonoRepositoryStatus) {
		r.Conditions = make([]metav1.Condition, len(conditions))
		for i := range conditions {
			r.Conditions[i] = conditions[i].DieRelease()
		}
	})
}

var (
	MonoRepositoryConditionBlank = v1.ConditionBlank.Type(v1alpha1.MonoRepositoryConditionReady)
)
