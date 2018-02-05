package getServerConfig

import (
	"os"
	"strings"

	"github.com/kubeless/kubeless/pkg/langruntime"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//GetServerConfigCmd contains first-class command for displaying the current server config
var GetServerConfigCmd = &cobra.Command{
	Use:   "get-server-config",
	Short: "Print the current configuration of the controller",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cli := utils.GetClientOutOfCluster()
		controllerNamespace := os.Getenv("KUBELESS_NAMESPACE")
		kubelessConfig := os.Getenv("KUBELESS_CONFIG")

		if len(controllerNamespace) == 0 {
			controllerNamespace = "kubeless"
		}

		if len(kubelessConfig) == 0 {
			kubelessConfig = "kubeless-config"
		}
		config, err := cli.CoreV1().ConfigMaps(controllerNamespace).Get(kubelessConfig, metav1.GetOptions{})
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
