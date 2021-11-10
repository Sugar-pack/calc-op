/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	calculator "github.com/example/calc-opr/api/v1"
)

// CalculatorReconciler reconciles a Calculator object
type CalculatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.example.com,resources=calculators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.example.com,resources=calculators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.example.com,resources=calculators/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Calculator object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *CalculatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	lg := log.FromContext(ctx)

	lg.Info("start Reconcile")
	calc := &calculator.Calculator{}
	err := r.Get(ctx, req.NamespacedName, calc)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			lg.Info("Calc resource not found. Ignoring since object must be deleted")
			// remove secret if Calc deleted
			secret := &corev1.Secret{}
			err := r.Get(ctx, req.NamespacedName, secret)
			if err == nil {
				lg.Info("Found old secret")
				err := r.Delete(ctx, secret)
				if err != nil {
					return ctrl.Result{}, err
				}
				lg.Info("Removed old secret")
			}
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		lg.Error(err, "Failed to get Calc")
		return ctrl.Result{}, err
	}

	calc.Status.Result = calc.Spec.X + calc.Spec.Y
	calc.Status.Processed = true

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: calc.Namespace,
			Name:      calc.Name,
			Annotations: map[string]string{
				"manage-by": "calc-operator",
			},
		},
		StringData: map[string]string{
			"result": strconv.FormatInt(calc.Status.Result, 10),
		},
		Type: corev1.SecretTypeOpaque,
	}

	err = r.Client.Create(ctx, secret)
	if err != nil {
		return ctrl.Result{}, err
	}

	RTString := os.Getenv("RECONCILIATION_TIME")
	RT, err := strconv.ParseUint(RTString, 10, 64)
	if err != nil {
		return ctrl.Result{RequeueAfter: time.Duration(RT * 1000 * 1000 * 1000)}, nil
	}

	return ctrl.Result{Requeue: true}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CalculatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&calculator.Calculator{}).
		Complete(r)
}
