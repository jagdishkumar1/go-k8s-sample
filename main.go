package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	CLUSTER_A_DEPLOYMENT *v1.DeploymentList
)

func updateContext(kubeconfig *string, cluster_context *string) {

	configData, err := ioutil.ReadFile(*kubeconfig)
	if err != nil {
		log.Panicf("failed reading data from file: %s", err)
	}

	data := make(map[interface{}]interface{})

	err2 := yaml.Unmarshal(configData, &data)

	if err2 != nil {

		log.Fatal(err2)
	}

	for k, _ := range data {
		if k == "current-context" {
			data[k] = *cluster_context
			break
		}
	}

	b, err := yaml.Marshal(&data)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(*kubeconfig, b, 0)

	if err != nil {
		log.Fatal(err)
	}
}

// GetDeploymentList returns a list of deployment in a namepsace
func GetDeploymentList(clientset kubernetes.Interface, namespace *string) (*v1.DeploymentList, error) {
	deploymentsClient := clientset.AppsV1().Deployments(*namespace)
	deployments, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return deployments, nil
}

// CreateDeployment creates of deployment in a namepsace
func CreateDeployment(clientset kubernetes.Interface, namespace *string, deployment v1.Deployment) (*v1.Deployment, error) {
	deploymentsClient := clientset.AppsV1().Deployments(*namespace)
	result, err := deploymentsClient.Create(context.TODO(), &deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateNamespace creates a namepsace
func CreateNamespace(clientset kubernetes.Interface, namespace *string) (*corev1.Namespace, error) {
	nsSpec := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: *namespace}}
	result, err := clientset.CoreV1().Namespaces().Create(context.TODO(), nsSpec, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	cluster_a_context := flag.String("cluster_a_context", "", "cluster A context")
	cluster_b_context := flag.String("cluster_b_context", "", "cluster B context")
	namespace := flag.String("namespace", "", "cluster namespaces")
	flag.Parse()

	fmt.Printf("\nCluster A: %v\n", *cluster_a_context)
	fmt.Printf("Cluster B: %v\n", *cluster_b_context)
	fmt.Printf("Namespace: %v\n\n", *namespace)

	if *namespace == "" || *cluster_b_context == "" {
		fmt.Printf("Help : go run main.go -kubeconfig <> -namespace <> -cluster_a_context <> -cluster_b_context <>\n")
		fmt.Printf("namespace and cluster_b_context are required args\n\n")
		os.Exit(1)
	}
	if *cluster_a_context != "" {
		updateContext(kubeconfig, cluster_a_context)
	}

	for _, v := range []int{0, 1} {
		// use the current context in kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}

		// create the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		deployments, err := GetDeploymentList(clientset, namespace)
		if err != nil {
			panic(err.Error())
		}

		if v == 0 {
			fmt.Printf("Currently there are %d deployments in the cluster %s in namespace %s\n\n", len(deployments.Items), strings.Split(*cluster_a_context, "/")[0], *namespace)
			CLUSTER_A_DEPLOYMENT = deployments
			updateContext(kubeconfig, cluster_b_context)
		}

		if v == 1 {
			result, err := CreateNamespace(clientset, namespace)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("Created namespace %q in the cluster %s .\n", result.GetObjectMeta().GetName(), strings.Split(*cluster_b_context, "/")[0])
			}

			fmt.Printf("Currently there are %d deployments in the cluster %s in namespace %s\n", len(deployments.Items), strings.Split(*cluster_b_context, "/")[0], *namespace)
			for _, deployment := range CLUSTER_A_DEPLOYMENT.Items {
				deployment.ObjectMeta = metav1.ObjectMeta{
					Name:      deployment.ObjectMeta.Name,
					Namespace: deployment.ObjectMeta.Namespace,
					Labels:    deployment.ObjectMeta.Labels,
				}
				result, err := CreateDeployment(clientset, namespace, deployment)
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
				}
			}

			deployments, err := GetDeploymentList(clientset, namespace)
			if err != nil {
				panic(err.Error())
			}
			fmt.Printf("Now there are %d deployments in the cluster %s in namespace %s\n\n", len(deployments.Items), strings.Split(*cluster_b_context, "/")[0], *namespace)
		}
	}
}
