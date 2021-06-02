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
        "fmt" 

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	injectorv1alpha1 "argano.com/sidecar/api/v1alpha1"
)

// SidecarReconciler reconciles a Sidecar object
type SidecarReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

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
	_ = log.FromContext(ctx)

	// your logic here
        fmt.Println("###Looping inside sidecar-controller!###")
	sidecarList := &injectorv1alpha1.SidecarList{}
	err :=  r.List(ctx,sidecarList)
	if(err != nil){
          fmt.Println("!!! Unable to find sidecars",err)
	}

	for i:=0;i<len(sidecarList.Items);i++ {
	  fmt.Println("****")
          fmt.Println("Name of sidecar[",i,"] = ",sidecarList.Items[i].Name)
	  fmt.Println("  Labels = ",sidecarList.Items[i].Labels)

	  podList := &corev1.PodList{}
          err := r.List(ctx,podList,client.MatchingLabels(sidecarList.Items[i].Labels))
	  if(err != nil) {
	    fmt.Println("!!! Unable to find Pods",err)
	  }

	  var podNames []string
	  for j:=0;j<len(podList.Items);j++ {
	    fmt.Println("    Pods = ",podList.Items[j].Name)
	    podNames = append(podNames,podList.Items[j].Name)
	  }

	  sidecarList.Items[i].Status.Nodes = podNames
	  sidecar := sidecarList.Items[i].DeepCopy()
	  r.Status().Update(ctx,sidecar)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SidecarReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&injectorv1alpha1.Sidecar{}).
//		Owns(&corev1.Pod{}).
		Complete(r)
}
