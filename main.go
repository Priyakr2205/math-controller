package main

import (
	//      "flag"
	"flag"
	"os"

	"k8s.io/client-go/rest"

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

/*var (
	masterURL  string
	kubeconfig string
)*/

func main() {

	klog.InitFlags(nil)
	// By default klog writes to stderr. Setting logtostderr to false makes klog
	// write to a log file.
	flag.Set("logtostderr", "false")
	flag.Set("log_file", "controller.log")
	flag.Parse()

	currentTime := time.Now()

	klog.Info("Application start time : ", currentTime.String())

	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := os.Getenv("HOME") + "/.kube/config"
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			klog.Fatalf("getClusterConfig: %v", err)
		}
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("getClusterConfig: %v", err)
	}

	klog.V(4).Info("Successfully constructed k8s client")

	// generating custom resource client.
	exampleClient, err := clientset.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error building example clientset: %s", err.Error())
	}
	//fmt.Println(exampleClient)
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
		klog.Info("Controller exited")
	}

	klog.Flush()
}
