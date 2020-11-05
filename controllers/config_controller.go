/*


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
	"path/filepath"

	"github.com/go-logr/logr"

	"github.com/openshift/cluster-network-operator/pkg/apply"
	"github.com/openshift/cluster-network-operator/pkg/render"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/pkg/errors"
	"github.com/yboaron/cluster-hosted-operator/pkg/names"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	//clusterstackv1beta1 "github.com/yboaron/cluster-hosted-operator/api/v1beta1"
	clusterstackv1beta1 "github.com/yboaron/cluster-hosted-operator/api/v1beta1"
)

// ConfigReconciler reconciles a Config object
type ConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=clusterstack.openshift.io,resources=configs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=clusterstack.openshift.io,resources=configs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=daemonsets/status,verbs=get;list;watch;create;update;patch;delete

func (r *ConfigReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctxt := context.Background()
	_ = r.Log.WithValues("config", req.NamespacedName)

	instance := &clusterstackv1beta1.Config{}
	err := r.Get(ctxt, req.NamespacedName, instance)
	if err != nil {
		// Error reading the object - requeue the req.
		return ctrl.Result{}, err
	}

	r.Log.Info("Returned object name", "name", req.NamespacedName.Name)

	err = r.applyKeepalived(instance)
	if err != nil {
		errors.Wrap(err, "failed applying LB")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterstackv1beta1.Config{}).
		Complete(r)
}

func (r *ConfigReconciler) applyKeepalived(instance *clusterstackv1beta1.Config) error {
	if instance.Spec.LoadBalancer == "Enable" {

		data := render.MakeRenderData()

		data.Data["HandlerNamespace"] = "cluster-config"
		data.Data["HandlerPrefix"] = "tst1"
		data.Data["onPremPlatformAPIServerInternalIP"] = "192.168.111.5"
		data.Data["onPremPlatformIngressIP"] = "192.168.111.4"

		err := r.renderAndApply(instance, data, "handler")
		return err
	} else {
		// TODO disable Keepalived code
	}
	return nil
}

func (r *ConfigReconciler) renderAndApply(instance *clusterstackv1beta1.Config, data render.RenderData, sourceDirectory string) error {
	var err error
	objs := []*uns.Unstructured{}

	sourceFullDirectory := filepath.Join(names.ManifestDir, "kubernetes-nmstate", sourceDirectory)

	objs, err = render.RenderDir(sourceFullDirectory, &data)
	if err != nil {
		return errors.Wrapf(err, "failed to render kubernetes-nmstate %s", sourceDirectory)
	}

	// If no file found in directory - return error
	if len(objs) == 0 {
		return fmt.Errorf("No manifests rendered from %s", sourceFullDirectory)
	}

	for _, obj := range objs {
		// RenderDir seems to add an extra null entry to the list. It appears to be because of the
		// nested templates. This just makes sure we don't try to apply an empty obj.
		if obj.GetName() == "" {
			continue
		}

		// Now apply the object
		err = apply.ApplyObject(context.TODO(), r.Client, obj)
		if err != nil {
			return errors.Wrapf(err, "failed to apply object %v", obj)
		}
	}

	return nil
}
