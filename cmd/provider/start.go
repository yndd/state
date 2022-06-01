/*
Copyright 2022 NDD.

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

package provider

import (
	"os"
	"strconv"
	"strings"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/pkg/errors"
	"github.com/pkg/profile"
	"github.com/spf13/cobra"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	pkgmetav1 "github.com/yndd/ndd-core/apis/pkg/meta/v1"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-runtime/pkg/ratelimiter"
	"github.com/yndd/ndd-runtime/pkg/shared"
	"github.com/yndd/state/internal/config"
	"github.com/yndd/state/internal/controllers"
	"github.com/yndd/state/internal/gnmiserver"
	//+kubebuilder:scaffold:imports
)

var (
	metricsAddr          string
	probeAddr            string
	enableLeaderElection bool
	concurrency          int
	pollInterval         time.Duration
	namespace            string
	podname              string
	grpcServerAddress    string
	grpcQueryAddress     string
	autoPilot            bool
)

// startCmd represents the start command for the network device driver
var startCmd = &cobra.Command{
	Use:          "start",
	Short:        "start the state ndd provider manager",
	Long:         "start the state ndd provider manager",
	Aliases:      []string{"start"},
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		zlog := zap.New(zap.UseDevMode(debug), zap.JSONEncoder())
		if debug {
			// Only use a logr.Logger when debug is on
			ctrl.SetLogger(zlog)
		}

		if profiler {
			defer profile.Start().Stop()
			go func() {
				http.ListenAndServe(":8000", nil)
			}()
		}

		zlog.Info("create manager")
		mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
			Scheme:             scheme,
			MetricsBindAddress: metricsAddr,
			WebhookServer: &webhook.Server{
				Port: 9443,
			},
			Port:                   7443,
			HealthProbeBindAddress: probeAddr,
			LeaderElection:         enableLeaderElection,
			LeaderElectionID:       "c66ce353.ndd.yndd.io",
		})
		if err != nil {
			return errors.Wrap(err, "Cannot create manager")
		}

		// assign gnmi address
		var gnmiAddress string
		if grpcQueryAddress != "" {
			gnmiAddress = grpcQueryAddress
		} else {
			gnmiAddress = strings.Join([]string{"127.0.0.1", strconv.Itoa(pkgmetav1.GnmiServerPort)}, ":")
		}

		zlog.Info("gnmi address", "address", gnmiAddress)

		// initialize controllers
		if err := controllers.Setup(mgr, &shared.NddControllerOptions{
			Logger:      logging.NewLogrLogger(zlog.WithName("state")),
			Poll:        pollInterval,
			Namespace:   namespace,
			GnmiAddress: gnmiAddress,
		}); err != nil {
			return errors.Wrap(err, "Cannot add ndd controllers to manager")
		}

		// initialize the gnmiserver
		s := gnmiserver.New(
			cmd.Context(),
			gnmiserver.WithLogger(logging.NewLogrLogger(zlog.WithName("gnmi server"))),
			gnmiserver.WithConfig(config.New()),
			gnmiserver.WithK8sClient(mgr.GetClient()),
		)
		if err := s.Start(); err != nil {
			return errors.Wrap(err, "Cannot start gnmi server")
		}

		// +kubebuilder:scaffold:builder

		if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
			return errors.Wrap(err, "unable to set up health check")
		}
		if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
			return errors.Wrap(err, "unable to set up ready check")
		}

		zlog.Info("starting manager")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			return errors.Wrap(err, "problem running manager")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVarP(&metricsAddr, "metrics-bind-address", "m", ":8080", "The address the metric endpoint binds to.")
	startCmd.Flags().StringVarP(&probeAddr, "health-probe-bind-address", "p", ":8081", "The address the probe endpoint binds to.")
	startCmd.Flags().BoolVarP(&enableLeaderElection, "leader-elect", "l", false, "Enable leader election for controller manager. "+
		"Enabling this will ensure there is only one active controller manager.")
	startCmd.Flags().IntVarP(&concurrency, "concurrency", "", 1, "Number of items to process simultaneously")
	startCmd.Flags().DurationVarP(&pollInterval, "poll-interval", "", 10*time.Minute, "Poll interval controls how often an individual resource should be checked for drift.")
	startCmd.Flags().StringVarP(&namespace, "namespace", "n", os.Getenv("POD_NAMESPACE"), "Namespace used to unpack and run packages.")
	startCmd.Flags().StringVarP(&podname, "podname", "", os.Getenv("POD_NAME"), "Name from the pod")
	startCmd.Flags().StringVarP(&grpcServerAddress, "grpc-server-address", "s", "", "The address of the grpc server binds to.")
	startCmd.Flags().StringVarP(&grpcQueryAddress, "grpc-query-address", "", "", "Validation query address.")
	startCmd.Flags().BoolVarP(&autoPilot, "autopilot", "a", true,
		"Apply delta/diff changes to the config automatically when set to true, if set to false the provider will report the delta and the operator should intervene what to do with the delta/diffs")
}

func nddCtlrOptions(c int) controller.Options {
	return controller.Options{
		MaxConcurrentReconciles: c,
		RateLimiter:             ratelimiter.NewDefaultProviderRateLimiter(ratelimiter.DefaultProviderRPS),
	}
}
