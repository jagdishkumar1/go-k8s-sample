package main

import (
	"fmt"
	"strings"
	"testing"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetDeploymentList(t *testing.T) {
	t.Parallel()
	testData := []struct {
		clientset                kubernetes.Interface
		countExpectedDeployments int
		inputNamespace           string
		err                      error
	}{
		// deployments found in the namespace
		{
			clientset: fake.NewSimpleClientset(&v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-app-1",
					Namespace:   "test",
					Annotations: map[string]string{},
				},
			}, &v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-app-2",
					Namespace:   "test",
					Annotations: map[string]string{},
				},
			}),
			inputNamespace:           "test",
			countExpectedDeployments: 2,
		},
		// No deployments in the namespace
		{
			clientset: fake.NewSimpleClientset(&v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-app-1",
					Namespace:   "test",
					Annotations: map[string]string{},
				},
			}, &v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-app-2",
					Namespace:   "test",
					Annotations: map[string]string{},
				},
			}),
			inputNamespace:           "test-1",
			countExpectedDeployments: 0,
		},
	}

	for _, data := range testData {
		t.Run("", func(data struct {
			clientset                kubernetes.Interface
			countExpectedDeployments int
			inputNamespace           string
			err                      error
		}) func(t *testing.T) {
			return func(t *testing.T) {
				deployments, err := GetDeploymentList(data.clientset, &data.inputNamespace)
				if err != nil {
					if data.err == nil {
						t.Fatalf(err.Error())
					}
					if !strings.EqualFold(data.err.Error(), err.Error()) {
						t.Fatalf("expected err: %s got err: %s", data.err, err)
					}
				} else {
					if len(deployments.Items) != data.countExpectedDeployments {
						t.Fatalf("expected %d deployments, got %d", data.countExpectedDeployments, len(deployments.Items))
					}
				}
			}
		}(data))
	}
}

func TestCreateDeployment(t *testing.T) {
	t.Parallel()
	testData := []struct {
		clientset      kubernetes.Interface
		deploymentName string
		inputNamespace string
		deployment     v1.Deployment
		err            error
	}{
		// deployments created in the namespace
		{
			clientset: fake.NewSimpleClientset(&v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-app-2",
					Namespace:   "test",
					Annotations: map[string]string{},
				},
			}),
			inputNamespace: "test",
			deploymentName: "test-app-1",
			deployment: v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-app-1",
					Namespace:   "test",
					Annotations: map[string]string{},
				},
			},
		},
		// deployments already exists in the namespace
		{
			clientset: fake.NewSimpleClientset(&v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-app-1",
					Namespace:   "test",
					Annotations: map[string]string{},
				},
			}),
			inputNamespace: "test",
			deploymentName: "test-app-1",
			deployment: v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-app-1",
					Namespace:   "test",
					Annotations: map[string]string{},
				},
			},
			err: fmt.Errorf("deployments.apps \"test-app-1\" already exists"),
		},
	}

	for _, data := range testData {
		t.Run("", func(data struct {
			clientset      kubernetes.Interface
			deploymentName string
			inputNamespace string
			deployment     v1.Deployment
			err            error
		}) func(t *testing.T) {
			return func(t *testing.T) {
				result, err := CreateDeployment(data.clientset, &data.inputNamespace, data.deployment)
				if err != nil {
					if data.err == nil {
						t.Fatalf(err.Error())
					}
					if !strings.EqualFold(data.err.Error(), err.Error()) {
						t.Fatalf("expected err: %s got err: %s", data.err, err)
					}
				} else {
					if data.deploymentName != result.GetObjectMeta().GetName() {
						t.Fatalf("expected %s deployments, got %s", data.deploymentName, result.GetObjectMeta().GetName())
					}
				}
			}
		}(data))
	}
}

func TestCreateNamespace(t *testing.T) {
	t.Parallel()
	testData := []struct {
		clientset      kubernetes.Interface
		inputNamespace string
		err            error
	}{
		// namespace created
		{
			clientset: fake.NewSimpleClientset(&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-nm-1",
					Annotations: map[string]string{},
				},
			}),
			inputNamespace: "test-nm",
		},
		// namespace already exists
		{
			clientset: fake.NewSimpleClientset(&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-nm-1",
					Annotations: map[string]string{},
				},
			}),
			inputNamespace: "test-nm-1",
			err:            fmt.Errorf("namespaces \"test-nm-1\" already exists"),
		},
	}

	for _, data := range testData {
		t.Run("", func(data struct {
			clientset      kubernetes.Interface
			inputNamespace string
			err            error
		}) func(t *testing.T) {
			return func(t *testing.T) {
				result, err := CreateNamespace(data.clientset, &data.inputNamespace)
				if err != nil {
					if data.err == nil {
						t.Fatalf(err.Error())
					}
					if !strings.EqualFold(data.err.Error(), err.Error()) {
						t.Fatalf("expected err: %s got err: %s", data.err, err)
					}
				} else {
					if data.inputNamespace != result.GetObjectMeta().GetName() {
						t.Fatalf("expected %s deployments, got %s", data.inputNamespace, result.GetObjectMeta().GetName())
					}
				}
			}
		}(data))
	}
}
