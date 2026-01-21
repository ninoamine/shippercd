package shippercontroller


import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
)



type EvironmentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}