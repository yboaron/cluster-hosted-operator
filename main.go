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

package main

import (
	"flag"
	"os"

	osclientset "github.com/openshift/client-go/config/clientset/versioned"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	clusterstackv1beta1 "github.com/yboaron/cluster-hosted-operator/api/v1beta1"
	"github.com/yboaron/cluster-hosted-operator/controllers"
	"github.com/yboaron/cluster-hosted-operator/pkg/names"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(clusterstackv1beta1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var imagesJSONFilename string

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&imagesJSONFilename, "images-json", "/etc/cluster-hosted-operator/images/images.json",
		"The location of the file containing the images to use for our operands.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	releaseVersion := os.Getenv("RELEASE_VERSION")
	if releaseVersion == "" {
		ctrl.Log.Info("Environment variable RELEASE_VERSION not provided")
	}

	config := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "0b541b0c.my.domain",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	osClient := osclientset.NewForConfigOrDie(rest.AddUserAgent(config, names.ControllerComponentName))
	// Check the Platform Type to determine the state of the CO
	enabled, err := controllers.IsEnabled(osClient)
	if err != nil {
		setupLog.Error(err, "could not determine whether to run")
		os.Exit(1)
	}
	if !enabled {
		//Set ClusterOperator status to disabled=true, available=true
		err = controllers.SetCOInDisabledState(osClient, releaseVersion)
		if err != nil {
			setupLog.Error(err, "unable to set baremetal ClusterOperator to Disabled")
			os.Exit(1)
		}
	}

	if err = (&controllers.ConfigReconciler{
		Client:         mgr.GetClient(),
		Log:            ctrl.Log.WithName("controllers").WithName("Config"),
		Scheme:         mgr.GetScheme(),
		ImagesFilename: imagesJSONFilename,
		OSClient:       osClient,
		ReleaseVersion: releaseVersion,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Config")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
