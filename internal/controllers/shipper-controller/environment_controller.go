package shippercontroller


import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	corev1alpha1 "github.com/ninoamine/shippercd/api/shipper-controller/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)



type EvironmentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *EvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error){

	logger := log.FromContext(ctx)

	var environment corev1alpha1.Environment

	if err := r.Get(ctx, req.NamespacedName, &environment); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Environment resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Environment")
		return ctrl.Result{}, err
	} 

	return ctrl.Result{}, nil

}


func (r *EvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Environment{}).
		Named("environment").Complete(r)
}