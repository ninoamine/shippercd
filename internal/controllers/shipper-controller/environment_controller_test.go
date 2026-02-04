package shippercontroller_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	corev1alpha1 "github.com/ninoamine/shippercd/api/shipper-controller/v1alpha1"
	shippercontroller "github.com/ninoamine/shippercd/internal/controllers/shipper-controller"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func setupScheme(t *testing.T) *runtime.Scheme {
	scheme := runtime.NewScheme()
	err := corev1alpha1.AddToScheme(scheme)
	assert.NoError(t, err)
	return scheme
}

func TestReconcile_EnvironmentExists(t *testing.T) {
	scheme := setupScheme(t)

	environment := &corev1alpha1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-environment",
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
			Name:      "test-environment",
			Namespace: "default",
		},
	}

	res, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)

}

func TestReconcile_EnvironmentNotFound(t *testing.T) {
	scheme := setupScheme(t)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	reconciler := &shippercontroller.EnvironmentReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent-environment",
			Namespace: "default",
		},
	}

	res, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)
}

func TestReconcile_EnvironmentBeingDeleted(t *testing.T) {
	scheme := setupScheme(t)

	now := metav1.Now()

	environment := &corev1alpha1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "deleting-environment",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"environment.shippercd.io/finalizer"},
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
			Name:      "deleting-environment",
			Namespace: "default",
		},
	}

	res, err := reconciler.Reconcile(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)

	updated := &corev1alpha1.Environment{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "deleting-environment",
		Namespace: "default",
	}, updated)

	assert.Error(t, err)
}
