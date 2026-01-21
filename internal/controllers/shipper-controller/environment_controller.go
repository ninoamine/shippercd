package shippercontroller


import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)



type EvironmentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *EvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error){

	return ctrl.Result{}, nil

}