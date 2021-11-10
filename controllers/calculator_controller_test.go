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
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strconv"
	"testing"

	calculator "github.com/example/calc-opr/api/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestCalculatorReconcilerValid(t *testing.T) {
	calc := &calculator.Calculator{
		TypeMeta: metav1.TypeMeta{
			Kind: "Calculator",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name: "test",
		},
		Spec: calculator.CalculatorSpec{
			X: 5,
			Y: 7,
		},
		Status: calculator.CalculatorStatus{
			Processed: false,
			Result:    0,
		},
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: calc.Namespace,
			Name:      calc.Name,
			Annotations: map[string]string{
				"manage-by": "calc-operator",
			},
		},
		StringData: map[string]string{
			"result": strconv.FormatInt(calc.Spec.X+calc.Spec.Y, 10),
		},
		Type: corev1.SecretTypeOpaque,
	}

	s := scheme.Scheme
	s.AddKnownTypes(appsv1.SchemeGroupVersion, calc, &calculator.Calculator{}, &calculator.CalculatorList{})

	cl := fake.NewClientBuilder().WithObjects(calc).Build()

	r := CalculatorReconciler{
		Client: cl,
		Scheme: s,
	}
	ctx := context.TODO()
	nsn := types.NamespacedName{
		Namespace: calc.Namespace,
		Name:      calc.Name,
	}
	req := reconcile.Request{
		NamespacedName: nsn,
	}
	_, err := r.Reconcile(ctx, req)

	assert.NoError(t, err)

	madeSecret := &corev1.Secret{}
	cl.Get(context.TODO(), types.NamespacedName{
		Namespace: calc.Namespace,
		Name:      calc.Name,
	}, madeSecret)
	assert.NoError(t, err)
	assert.Equal(t, secret.StringData,madeSecret.StringData)
	assert.Equal(t, secret.Annotations, madeSecret.Annotations)
}
