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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	injectorv1alpha1 "argano.com/sidecar/api/v1alpha1"
)

// SidecarReconciler reconciles a Sidecar object
type SidecarReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Initialize logger
var logger = logf.Log.WithName("sidecar-controller")

//+kubebuilder:rbac:groups=injector.argano.com,resources=sidecars,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=injector.argano.com,resources=sidecars/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=injector.argano.com,resources=sidecars/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Sidecar object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *SidecarReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	logger = logf.Log.WithName(req.Namespace+"/"+req.Name)
	// Controller Logic Here
	// 1. Get Object of type Sidecar
	sidecar := &injectorv1alpha1.Sidecar{}
	err := r.Get(ctx,req.NamespacedName,sidecar)
	if(err != nil){
          logger.Error(err,"Unable to fetch Sidecar object ",req.Namespace,req.Name)
	  return ctrl.Result{},err
	}

	// 2. Get List of Active Pods with matching labels
	podList := &corev1.PodList{}
	opts := []client.ListOption{
	  client.InNamespace(req.NamespacedName.Namespace),
	  client.MatchingLabels(sidecar.Labels),
	}

	err = r.List(ctx,podList,opts...)
        if(err != nil){
	  logger.Error(err,"Unable to list Pods",req.Namespace,req.Name)
	  return ctrl.Result{},err
	}

	// 3. If Pod has annotation not marked, then replicate the pod with sidecar container and mark the annotations
	for i:=0;i<len(podList.Items);i++ {
          pod := podList.Items[i]
	  if ( pod.Status.Phase == "Running" && pod.Annotations["sidecar-injector"] != req.Name ) {
            logger.Info(pod.Name)
	    newPod := pod.DeepCopy()
	    annotations :=  newPod.Annotations
	    if(annotations == nil) {
	      annotations = make(map[string]string)
	    }
	    annotations["sidecar-injector"] = req.Name
	    newPod.ObjectMeta.Annotations = annotations
	    newPod.ObjectMeta.ResourceVersion = ""
	    newPod.Spec.Containers = append(newPod.Spec.Containers,sidecar.Spec.Containers...)
	    newPod.Spec.Volumes    = append(newPod.Spec.Volumes,sidecar.Spec.Volumes...)
	    newPod.Name = newPod.Name+"-injected"
	    // Recreate the POD with Sidecar Injected
	    err = r.Create(ctx,newPod)
	    if(err != nil){
              logger.Error(err,"Unable to recreate Pod "+newPod.Name)
	    }

	    oldPod := pod.DeepCopy()
            oldPod.ObjectMeta.Annotations = annotations
	    // Mark annotation on the old POD
	    err = r.Update(ctx,oldPod)
	    if(err != nil){
	      logger.Error(err,"Unable to update Pod "+oldPod.Name)
	    }

          } 
	}

	// 4. Get List of Pods and check if annotation is marked, then add the Pod name to Sidecar.Status.Nodes
	podList = &corev1.PodList{}
        err = r.List(ctx,podList,opts...)
	if(err != nil){
	  logger.Error(err,"Unable to list Pods",req.Namespace,req.Name)
	}

	var podNames []string
	for i:=0;i<len(podList.Items);i++ {
	  pod := podList.Items[i]
	  if(pod.Status.Phase == "Running" && pod.Annotations["sidecar-injector"] == req.Name) {
	    podNames = append(podNames,pod.Name)
	  }
	}

        if(len(podNames) == 0) {
          sidecar.Status.Nodes = make([]string,0,0)
	} else {
          sidecar.Status.Nodes = podNames
	}

	err = r.Status().Update(ctx,sidecar)
	if(err != nil){
          logger.Error(err,"Unable to update Sidecar.Status.Nodes")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SidecarReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&injectorv1alpha1.Sidecar{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
