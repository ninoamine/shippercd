package shippercontroller_test


import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1alpha1 "github.com/ninoamine/shippercd/api/shipper-controller/v1alpha1"
	shippercontroller "github.com/ninoamine/shippercd/internal/controllers/shipper-controller"
	ctrl "sigs.k8s.io/controller-runtime"
	"k8s.io/apimachinery/pkg/types"
)


func setupScheme(t *testing.T) *runtime.Scheme {
	scheme := runtime.NewScheme()
	err := corev1alpha1.AddToScheme(scheme)
	assert.NoError(t, err)
	return scheme
}


func TestReconcile_EnvironmentExists(t *testing.T){
	scheme := setupScheme(t)

	environment := &corev1alpha1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: 	"test-environment",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(environment).
		Build()

	
	reconciler := &shippercontroller.EnvironmentReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-environment",
			Namespace: "default",
		},
	}

	res, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)

}