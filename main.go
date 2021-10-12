package main

import (
	//      "flag"
	"flag"
	"fmt"

	//"os"

	"time"

	//kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	//      "k8s.io/sample-controller/pkg/signals"
	clientset "math-controller/pkg/client/clientset/versioned"
	informers "math-controller/pkg/client/informers/externalversions"
	"math-controller/pkg/signals"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("getClusterConfig: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("getClusterConfig: %v", err)
	}

	klog.Info("Successfully constructed k8s client")

	// generating custom resource client.
	exampleClient, err := clientset.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error building example clientset: %s", err.Error())
	}
	fmt.Println(exampleClient)
	//creating sharedinformers for custom resource
	//exampleInformerFactory := informers.NewSharedInformerFactory(exampleClient, time.Second*30)

	/*pods, err := kubeClient.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	fmt.Println(pods)
	*/

	exampleInformerFactory := informers.NewSharedInformerFactory(exampleClient, time.Second*30)
	//kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)

	controller := NewController(kubeClient, exampleClient,
		exampleInformerFactory.Mathematics().V1alpha1().Maths())

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	exampleInformerFactory.Start(stopCh)
	err = controller.Run(3, stopCh)
	if err != nil {
		fmt.Println("Controller exited")
	}

}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}