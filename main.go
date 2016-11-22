package main

import (
	"flag"
	"fmt"

	"github.com/skippbox/kubeless/pkg/controller"
	"github.com/skippbox/kubeless/pkg/utils"
)

var (
	masterHost string
)

func init() {
	flag.StringVar(&masterHost, "master", "", "API Server addr")
	flag.Parse()
}

func main() {
	cfg := newControllerConfig()
	c := controller.New(cfg)
	err := c.Run()
	if err != nil {
		fmt.Errorf("Kubeless controller running failed: %s", err)
	}
}

func newControllerConfig() controller.Config {

	kubecli, ns, err := utils.GetClient()
	if err != nil {
		fmt.Errorf("Can not get kubernetes config: %s", err)
	}
	cfg := controller.Config{
		Namespace: ns,
		KubeCli:   kubecli,
		MasterHost: masterHost,
	}

	return cfg
}
