package getserverconfig

import (
	"strings"

	"github.com/kubeless/kubeless/pkg/langruntime"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// GetServerConfigCmd contains first-class command for displaying the current server config
var GetServerConfigCmd = &cobra.Command{
	Use:   "get-server-config",
	Short: "Print the current configuration of the controller",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cli := utils.GetClientOutOfCluster()
		apiExtensionsClientset := utils.GetAPIExtensionsClientOutOfCluster()
		config, err := utils.GetKubelessConfig(cli, apiExtensionsClientset)
		if err != nil {
			logrus.Fatalf("Unable to read the configmap: %v", err)
		}

		var lr = langruntime.New(config)
		lr.ReadConfigMap()

		logrus.Info("Current Server Config:")
		logrus.Infof("Supported Runtimes are: %s",
			strings.Join(lr.GetRuntimes(), ", "))
	},
}
