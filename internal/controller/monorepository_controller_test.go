package controller_test

import (
	"github.com/garethjevans/monorepository-controller/api/v1alpha1"
	"github.com/garethjevans/monorepository-controller/internal/controller"
	"github.com/garethjevans/monorepository-controller/internal/tests/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"

	apiv1 "github.com/fluxcd/source-controller/api/v1"
	apiv1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	apiv1beta2 "github.com/fluxcd/source-controller/api/v1beta2"

	v1 "dies.dev/apis/meta/v1"
	"github.com/vmware-labs/reconciler-runtime/reconcilers"
	rtesting "github.com/vmware-labs/reconciler-runtime/testing"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

func TestMonoRepository(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)

	// mono repository
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	// flux
	utilruntime.Must(apiv1beta1.AddToScheme(scheme))
	utilruntime.Must(apiv1beta2.AddToScheme(scheme))
	utilruntime.Must(apiv1.AddToScheme(scheme))

	baseMonoRepo := resources.MonoRepositoryBlank.
		MetadataDie(func(d *v1.ObjectMetaDie) {
			d.Name("mono-repository")
			d.Namespace("dev")
		})

	ts := rtesting.SubReconcilerTests[*v1alpha1.MonoRepository]{
		"Contains a sub resource": {
			Resource: baseMonoRepo.SpecDie(func(d *resources.MonoRepositorySpecDie) {
				d.GitRepository(apiv1beta2.GitRepositorySpec{
					URL: "https://github.com/org/repo",
				})
			}).DieReleasePtr(),

			ExpectResource: baseMonoRepo.SpecDie(func(d *resources.MonoRepositorySpecDie) {
				d.GitRepository(apiv1beta2.GitRepositorySpec{
					URL: "https://github.com/org/repo",
				})
			}).StatusDie(func(d *resources.MonoRepositoryStatusDie) {
				d.ConditionsDie(
				//resources.ManagedResourceHealthyBlank.Reason("ReadyCondition").Unknown().Message("condition with type [Ready] not found on resource status"),
				)
			}).DieReleasePtr(),

			ExpectCreates: []client.Object{
				&apiv1beta2.GitRepository{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: "mono-repository-mr-",
						Namespace:    "dev",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion:         "source.garethjevans.org/v1alpha1",
								Kind:               "MonoRepository",
								Name:               "mono-repository",
								Controller:         pointer.Bool(true),
								BlockOwnerDeletion: pointer.Bool(true),
							},
						},
					},
					Spec: apiv1beta2.GitRepositorySpec{
						URL: "https://github.com/org/repo",
					},
					Status: apiv1beta2.GitRepositoryStatus{},
				},
			},

			ExpectEvents: []rtesting.Event{
				rtesting.NewEvent(baseMonoRepo, scheme, corev1.EventTypeNormal, "Created", "Created GitRepository %q", "mono-repository-mr-001"),
			},
		},
	}

	ts.Run(t, scheme, func(t *testing.T, rtc *rtesting.SubReconcilerTestCase[*v1alpha1.MonoRepository], c reconcilers.Config) reconcilers.SubReconciler[*v1alpha1.MonoRepository] {
		return controller.NewResourceValidator(c)
	})
}
