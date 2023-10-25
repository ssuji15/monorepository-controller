package dies

import (
	"github.com/garethjevans/monorepository-controller/api/v1alpha1"
)

// +die:object=true
type _ = v1alpha1.MonoRepository

// +die
type _ = v1alpha1.MonoRepositorySpec

//func (d *ComponentSpecDie) PipelineRunDie(fn func(d *PipelineRunDie)) *ComponentSpecDie {
//	return d.DieStamp(func(r *cartographerv1alpha1.ComponentSpec) {
//		d := PipelineRunBlank.DieImmutable(false).DieFeed(r.PipelineRun)
//		fn(d)
//		r.PipelineRun = d.DieRelease()
//	})
//}

//var (
//	ManagedResourceSucceededBlank = diecorev1.ConditionBlank.Type(v1alpha1.ManagedResourceConditionSucceeded)
//)
