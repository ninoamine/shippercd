package main


import (
	ctrl "sigs.k8s.io/controller-runtime"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	scheme = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func main() {

	mgr, err := ctrl.NewManager( ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,	
	})

	if err != nil {
		setupLog.Error(err, "unable to start manager")
	}

}