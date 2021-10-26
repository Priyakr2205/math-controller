package main

import (
	"context"
	"fmt"
	mathv1alpha1 "math-controller/pkg/apis/mathematics/v1alpha1"
	clientset "math-controller/pkg/client/clientset/versioned"
	samplescheme "math-controller/pkg/client/clientset/versioned/scheme"
	informers "math-controller/pkg/client/informers/externalversions/mathematics/v1alpha1"
	listers "math-controller/pkg/client/listers/mathematics/v1alpha1"
	"reflect"
	"strconv"
	"time"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog/v2"

	//      "context"
	//	"k8s.io/client-go/kubernetes"
	//	appslisters "k8s.io/client-go/listers/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const controllerAgentName = "math-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Foo"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "Math  synced successfully"
)

type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	//kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	mathclientset clientset.Interface

	//deploymentsLister appslisters.DeploymentLister
	//deploymentsSynced cache.InformerSynced
	mathLister listers.MathLister
	mathSynced cache.InformerSynced

	//add workqueue for adding events
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	mathclientset clientset.Interface,
	//deploymentInformer appsinformers.DeploymentInformer,
	mathInformer informers.MathInformer) *Controller {

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		//kubeclientset:     kubeclientset,
		mathclientset: mathclientset,
		//deploymentsLister: deploymentInformer.Lister(),
		//deploymentsSynced: deploymentInformer.Informer().HasSynced,
		mathLister: mathInformer.Lister(),
		mathSynced: mathInformer.Informer().HasSynced,
		workqueue:  workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Maths"),
		recorder:   recorder,
	}

	klog.V(4).Info("Setting up event handlers")
	// Set up an event handler for when Math resources change
	mathInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueMath,
		UpdateFunc: func(old, new interface{}) {
			newMath := new.(*mathv1alpha1.Math)
			oldMath := old.(*mathv1alpha1.Math)
			if reflect.DeepEqual(newMath.Spec, oldMath.Spec) {
				klog.V(4).Info("Specs not modified. Ignoring update event")
				return
			}
			controller.enqueueMath(new)
		},
		DeleteFunc: func(obj interface{}) {
			if resource, ok := obj.(*mathv1alpha1.Math); ok {
				klog.Infof("Math Resource %s@%s deleted", resource.Name, resource.Namespace)

			}
		},
	})

	return controller
}

func (c *Controller) enqueueMath(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Math controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.mathSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Math resources
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {

	}

}

func (c *Controller) processNextWorkItem() bool {

	// get obj from worker queue
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Math resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) syncHandler(key string) error {

	defer func() {
		if r := recover(); r != nil {
			klog.Errorln("recovered in syncHandler(). Error : ", r)
		}
	}()

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Math resource with this namespace/name
	math, err := c.mathLister.Maths(namespace).Get(name)
	if err != nil {
		// The Math resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("foo '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	klog.Info("Numbers are", *math.Spec.Number1, *math.Spec.Number2)
	var result int32
	var oper string
	switch math.Spec.Operation {
	case "add", "Add", "Addition", "addition":
		result = *math.Spec.Number1 + *math.Spec.Number2
		oper = "Sum"
		klog.Info("Sum of ", *math.Spec.Number1, " and ", *math.Spec.Number2, " is ", result)

		break
	case "sub", "Sub", "Substraction", "substraction":
		result = *math.Spec.Number1 - *math.Spec.Number2
		oper = "Difference"
		klog.Info("Difference of ", *math.Spec.Number1, " and ", *math.Spec.Number2, " is ", result)

		break
	case "div", "Div", "Division", "division":
		result = *math.Spec.Number1 % *math.Spec.Number2
		oper = "Remainder"
		klog.Info("Remainder of ", *math.Spec.Number1, " divided by", *math.Spec.Number2, " is ", result)

		break
	case "multiply", "Multiply":
		result = *math.Spec.Number1 * *math.Spec.Number2
		oper = "Product"
		klog.Info("Product of ", *math.Spec.Number1, " and ", *math.Spec.Number2, " is ", result)

		break
	default:
		return fmt.Errorf("Invalid operation. Skipping processing for %s", math.Name)
	}

	err = c.updateMath(math, result, oper)
	if err != nil {
		return err
	}

	//number1 := math.Spec.Number1
	//number2 := math.Spec.Number2

	//fmt.Println("number1:",number1, "nu" number2)
	c.recorder.Event(math, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateMath(math *mathv1alpha1.Math, result int32, oper string) error {
	mathcopy := math.DeepCopy()
	mathcopy.Status.Status = "SUCCESS"
	mathcopy.Status.Message = oper + "=" + strconv.Itoa(int(result))
	mathcopy.Status.LastUpdateTime = metav1.Time{Time: time.Now()}

	klog.V(4).Info("service label ", mathcopy.ObjectMeta.Name)

	res := strconv.Itoa(int(result))

	mathcopy.ObjectMeta.Labels = map[string]string{
		oper: res,
	}

	_, err := c.mathclientset.MathematicsV1alpha1().Maths("default").UpdateStatus(context.TODO(), mathcopy, metav1.UpdateOptions{})

	return err
}
