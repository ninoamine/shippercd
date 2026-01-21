package shippercontroller


import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)



type EvironmentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *EvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error){

	logger := log.FromContext(ctx)

	var environment corev1alpha1.Encironment

	return ctrl.Result{}, nil

}