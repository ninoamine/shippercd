package shipperpostgresqlcontroller

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	corev1alpha1 "github.com/ninoamine/shippercd/api/shipper-postgresql-controller/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const databaseFinalizer = "database.shippercd.io/finalizer"

type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	PgPool *pgxpool.Pool
}

func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var database corev1alpha1.Database

	if err := r.Get(ctx, req.NamespacedName, &database); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Database resource already deleted", "name", req.Name, "namespace", req.Namespace)
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Database")
		return ctrl.Result{}, err
	}

	if database.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&database, databaseFinalizer) {
			logger.Info("Adding finalizer to Database", "name", database.Name)

			controllerutil.AddFinalizer(&database, databaseFinalizer)
			if err := r.Update(ctx, &database); err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}

		logger.Info("Reconciling Database normally", "name", database.Name)
		return ctrl.Result{}, nil
	}

	if controllerutil.ContainsFinalizer(&database, databaseFinalizer) {
		logger.Info("Handling deletion of Database", "name", database.Name)

		controllerutil.RemoveFinalizer(&database, databaseFinalizer)
		if err := r.Update(ctx, &database); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.createDatabase(ctx, database.Name); err != nil {
		logger.Error(err, "Failed to create database", "name", database.Name)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Database{}).
		Named("database").Complete(r)
}

var validDatabaseName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`)

func (r *DatabaseReconciler) createDatabase(ctx context.Context, databaseName string) error {
	if !validDatabaseName.MatchString(databaseName) {
		return fmt.Errorf("invalid database name: %s", databaseName)
	}

	databaseCheckQuery := `SELECT 1 FROM pg_database WHERE datname = $1;`

	var exists int

	err := r.PgPool.QueryRow(ctx, databaseCheckQuery, databaseName).Scan(&exists)
	if err == nil {
		return fmt.Errorf("database %s already exists", databaseName)
	}

	databaseCreateQuery := fmt.Sprintf(`CREATE DATABASE "%s";`, databaseName)

	_, err = r.PgPool.Exec(ctx, databaseCreateQuery)
	if err != nil {
		return fmt.Errorf("failed to create database %s: %v", databaseName, err)
	}

	return nil
}
