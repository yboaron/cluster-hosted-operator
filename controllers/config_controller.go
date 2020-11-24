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
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/openshift/cluster-network-operator/pkg/apply"
	"github.com/openshift/cluster-network-operator/pkg/render"
	corev1 "k8s.io/api/core/v1"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/pkg/errors"
	clusterstackv1beta1 "github.com/yboaron/cluster-hosted-operator/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ConfigReconciler reconciles a Config object
type ConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=daemonsets/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=namespaces;configmaps;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings;rolebindings;roles,verbs="*"
// +kubebuilder:rbac:groups="security.openshift.io",resources=securitycontextconstraints,verbs=get;list;watch;create;update;patch;delete
+// +kubebuilder:rbac:groups=clusterstack.openshift.io,resources=configs,verbs=get;list;watch;create;update;patch;delete
+// +kubebuilder:rbac:groups=clusterstack.openshift.io,resources=configs/status,verbs=get;update;patch

// +kubebuilder:rbac:namespace=tst-cluster-hosted-operator,groups=clusterstack.openshift.io,resources=configs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:namespace=tst-cluster-hosted-operator,groups=clusterstack.openshift.io,resources=configs/status,verbs=get;update;patch

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

	err = r.syncNamespace(instance)
	if err != nil {
		errors.Wrap(err, "failed applying Namespace")
		return ctrl.Result{}, err
	}

	err = r.syncRBAC(instance)
	if err != nil {
		errors.Wrap(err, "failed applying Namespace")
		return ctrl.Result{}, err
	}

	err = r.syncKeepalived(instance)
	if err != nil {
		errors.Wrap(err, "failed applying Keepalived")
		return ctrl.Result{}, err
	}

	err = r.syncHaproxy(instance)
	if err != nil {
		errors.Wrap(err, "failed applying Haproxy")
		return ctrl.Result{}, err
	}

	err = r.syncMDNS(instance)
	if err != nil {
		errors.Wrap(err, "failed applying MDNS")
		return ctrl.Result{}, err
	}

	err = r.syncCoreDNS(instance)
	if err != nil {
		errors.Wrap(err, "failed applying CoreDNS")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ConfigReconciler) syncRBAC(instance *clusterstackv1beta1.Config) error {

	// TODO:  add here code to check if RBAC resources already exist
	data := render.MakeRenderData()
	data.Data["HandlerNamespace"] = os.Getenv("HANDLER_NAMESPACE")

	err := r.renderAndApply(instance, data, "rbac")
	if err != nil {
		errors.Wrap(err, "failed applying RBAC")
		return err
	}
	return r.renderAndApply(instance, data, "rbac")
}

func (r *ConfigReconciler) syncKeepalived(instance *clusterstackv1beta1.Config) error {

	// TODO:  add here code to check if Keepalived resources already exist
	data := render.MakeRenderData()
	data.Data["HandlerNamespace"] = os.Getenv("HANDLER_NAMESPACE")
	data.Data["OnPremPlatformAPIServerInternalIP"] = os.Getenv("ON_PREM_API_VIP")
	data.Data["OnPremPlatformIngressIP"] = os.Getenv("ON_PREM_INGRESS_VIP")
	data.Data["BaremetalRuntimeCfgImage"] = os.Getenv("BAREMETAL_RUNTIMECFG_IMAGE")
	data.Data["KeepalivedImage"] = os.Getenv("KEEPALIVED_IMAGE")

	err := r.renderAndApply(instance, data, "keepalived-configmap")
	if err != nil {
		errors.Wrap(err, "failed applying keepalived-configmap ")
		return err
	}
	return r.renderAndApply(instance, data, "keepalived-daemonset")
}

func (r *ConfigReconciler) syncCoreDNS(instance *clusterstackv1beta1.Config) error {

	// TODO:  add here code to check if CoreDNS resources already exist
	data := render.MakeRenderData()
	data.Data["HandlerNamespace"] = os.Getenv("HANDLER_NAMESPACE")
	data.Data["OnPremPlatformAPIServerInternalIP"] = os.Getenv("ON_PREM_API_VIP")
	data.Data["OnPremPlatformIngressIP"] = os.Getenv("ON_PREM_INGRESS_VIP")
	data.Data["BaremetalRuntimeCfgImage"] = os.Getenv("BAREMETAL_RUNTIMECFG_IMAGE")
	data.Data["CorednsImage"] = os.Getenv("COREDNS_IMAGE")
	data.Data["DnsBaseDomain"] = os.Getenv("DNS_BASE_DOMAIN")

	err := r.renderAndApply(instance, data, "coredns-configmap")
	if err != nil {
		errors.Wrap(err, "failed applying CoreDNS-configmap ")
		return err
	}
	return r.renderAndApply(instance, data, "coredns-daemonset")
}

func (r *ConfigReconciler) syncMDNS(instance *clusterstackv1beta1.Config) error {

	// TODO:  add here code to check if MDNS resources already exist
	data := render.MakeRenderData()
	data.Data["HandlerNamespace"] = os.Getenv("HANDLER_NAMESPACE")
	data.Data["OnPremPlatformAPIServerInternalIP"] = os.Getenv("ON_PREM_API_VIP")
	data.Data["OnPremPlatformIngressIP"] = os.Getenv("ON_PREM_INGRESS_VIP")
	data.Data["BaremetalRuntimeCfgImage"] = os.Getenv("BAREMETAL_RUNTIMECFG_IMAGE")
	data.Data["MdnsPublisherImage"] = os.Getenv("MDNS_PUBLISHER_IMAGE")

	err := r.renderAndApply(instance, data, "mdns-configmap")
	if err != nil {
		errors.Wrap(err, "failed applying Mdns-configmap ")
		return err
	}
	return r.renderAndApply(instance, data, "mdns-daemonset")
}

func (r *ConfigReconciler) syncHaproxy(instance *clusterstackv1beta1.Config) error {

	// TODO:  add here code to check if HAProxy resources already exist
	data := render.MakeRenderData()
	data.Data["HandlerNamespace"] = os.Getenv("HANDLER_NAMESPACE")
	data.Data["OnPremPlatformAPIServerInternalIP"] = os.Getenv("ON_PREM_API_VIP")
	data.Data["BaremetalRuntimeCfgImage"] = os.Getenv("BAREMETAL_RUNTIMECFG_IMAGE")
	data.Data["HaproxyImage"] = os.Getenv("HAPROXY_IMAGE")

	err := r.renderAndApply(instance, data, "haproxy-configmap")
	if err != nil {
		errors.Wrap(err, "failed applying haproxy-configmap ")
		return err
	}
	return r.renderAndApply(instance, data, "haproxy-daemonset")
}

func (r *ConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterstackv1beta1.Config{}).
		Owns(&corev1.Namespace{}).
		Complete(r)
}
func (r *ConfigReconciler) syncNamespace(instance *clusterstackv1beta1.Config) error {

	// TODO:  add here code to check if namespace exists
	data := render.MakeRenderData()
	data.Data["HandlerNamespace"] = os.Getenv("HANDLER_NAMESPACE")
	return r.renderAndApply(instance, data, "namespace")
}

func (r *ConfigReconciler) renderAndApply(instance *clusterstackv1beta1.Config, data render.RenderData, sourceDirectory string) error {
	var err error
	objs := []*uns.Unstructured{}

	sourceFullDirectory := filepath.Join( /*names.ManifestDir*/ "./bindata", "cluster-hosted", sourceDirectory)

	objs, err = render.RenderDir(sourceFullDirectory, &data)
	if err != nil {
		return errors.Wrapf(err, "failed to render cluster-hosted %s", sourceDirectory)
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
