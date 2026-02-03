package shippercontroller


import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	corev1alpha1 "github.com/ninoamine/shippercd/api/shipper-controller/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const environmentFinalizer = "environment.shippercd.io/finalizer"

type EnvironmentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *EnvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var environment corev1alpha1.Environment

	if err := r.Get(ctx, req.NamespacedName, &environment); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Environment resource already deleted", "name", req.Name, "namespace", req.Namespace)
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Environment")
		return ctrl.Result{}, err
	}

	if environment.DeletionTimestamp.IsZero() {

		if !controllerutil.ContainsFinalizer(&environment, environmentFinalizer) {
			logger.Info("Adding finalizer to Environment", "name", environment.Name)

			controllerutil.AddFinalizer(&environment, environmentFinalizer)
			if err := r.Update(ctx, &environment); err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}

		logger.Info("Reconciling Environment normally", "name", environment.Name)
		return ctrl.Result{}, nil
	}

	if controllerutil.ContainsFinalizer(&environment, environmentFinalizer) {

		logger.Info("Handling deletion of Environment", "name", environment.Name)

		controllerutil.RemoveFinalizer(&environment, environmentFinalizer)
		if err := r.Update(ctx, &environment); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}



func (r *EnvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Environment{}).
		Named("environment").Complete(r)
}