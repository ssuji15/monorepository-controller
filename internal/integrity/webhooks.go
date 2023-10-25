package integrity

import (
	"context"
	"github.com/garethjevans/monorepository-controller/api/v1alpha1"

	"github.com/vmware-labs/reconciler-runtime/reconcilers"
	admissionv1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

//+kubebuilder:webhook:path=/integrity-source-garethjevans-org-monorepository,mutating=false,failurePolicy=fail,sideEffects=None,groups=source.garethjevans.org,resources=managedresources,verbs=create;update;delete,versions={v1alpha1},matchPolicy=equivalent,name=integrity.managedresource.source.garethjevans.org,admissionReviewVersions={v1,v1beta1}

func RegisterReferentialIntegrityWebhooks(mgr manager.Manager) {
	c := reconcilers.NewConfig(mgr, nil, 0)
	mgr.GetWebhookServer().Register("/integrity-source-garethjevans-org-managedresource", MonoRepositoryReferentialIntegrityWebhook(c).Build())
}

// +kubebuilder:rbac:groups=source.garethjevans.org,resources=monorepositories,verbs=get;list;watch
func MonoRepositoryReferentialIntegrityWebhook(c reconcilers.Config) *reconcilers.AdmissionWebhookAdapter[*v1alpha1.MonoRepository] {
	return &reconcilers.AdmissionWebhookAdapter[*v1alpha1.MonoRepository]{
		Name: "MonoRepositoryReferentialIntegrityWebhook",
		Reconciler: &reconcilers.SyncReconciler[*v1alpha1.MonoRepository]{
			Setup: func(ctx context.Context, mgr manager.Manager, bldr *builder.Builder) error {
				return nil
			},
			Sync: func(ctx context.Context, resource *v1alpha1.MonoRepository) error {
				//c := reconcilers.RetrieveConfigOrDie(ctx)
				req := reconcilers.RetrieveAdmissionRequest(ctx)
				//resp := reconcilers.RetrieveAdmissionResponse(ctx)

				switch req.Operation {
				case admissionv1.Create, admissionv1.Update:
					// FIXME
				case admissionv1.Delete:
					// FIXME
				}

				return nil
			},
		},
		Config: c,
	}
}
