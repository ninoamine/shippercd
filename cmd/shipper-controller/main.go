package main


import (
	ctrl "sigs.k8s.io/controller-runtime"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"github.com/ninoamine/shippercd/internal/controllers/shipper-controller"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	corev1alpha1 "github.com/ninoamine/shippercd/api/shipper-controller/v1alpha1"

)

var (
	scheme = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(corev1alpha1.AddToScheme(scheme))
}


func main() {

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager( ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,	
	})

	if err != nil {
		setupLog.Error(err, "unable to start manager")
	}

	if err = (&shippercontroller.EvironmentReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err,"unable to create controller", "controller", "Environment")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}

}