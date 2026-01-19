package main


import (
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {

	mgr, err := ctrl.NewManager( ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		
	})

}