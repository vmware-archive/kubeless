/*
Copyright 2016 Skippbox, Ltd.

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

package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/skippbox/kubeless/pkg/function"
	"github.com/skippbox/kubeless/pkg/spec"
	"github.com/skippbox/kubeless/pkg/utils"

	k8sapi "k8s.io/kubernetes/pkg/api"
	unversionedAPI "k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/unversioned"
)

const (
	tprName = "lamb-da.k8s.io"
)

var (
	ErrVersionOutdated = errors.New("Requested version is outdated in apiserver")
	initRetryWaitTime  = 30 * time.Second
)

type rawEvent struct {
	Type   string
	Object json.RawMessage
}

type Event struct {
	Type   string
	Object *spec.Function
}

type Controller struct {
	logger       *logrus.Entry
	Config       Config
	waitFunction sync.WaitGroup
	Functions    map[string]*spec.Function
}

type Config struct {
	Namespace  string
	KubeCli    *unversioned.Client
	MasterHost string
}

func New(cfg Config) *Controller {
	return &Controller{
		logger:    logrus.WithField("pkg", "controller"),
		Config:    cfg,
		Functions: make(map[string]*spec.Function),
	}
}

func (c *Controller) Init() {
	c.logger.Infof("Initializing Kubeless controller...")
	for {
		//create TPR if it's not exists
		err := c.initResource()
		if err == nil {
			break
		}
		c.logger.Errorf("Initialization failed: %v", err)
		c.logger.Infof("Retry in %v...", initRetryWaitTime)
		time.Sleep(initRetryWaitTime)
	}
}

func (c *Controller) InstallKubeless() {
	c.logger.Infof("Installing Kubeless controller into Kubernetes deployment...")
	err := utils.DeployKubeless(c.Config.KubeCli)
	if err != nil {
		c.logger.Errorf("Kubeless controller installation failed: %v", err)
	} else {
		c.logger.Infof("Kubeless controller installation finished!")
	}
}

func (c *Controller) InstallMsgBroker() {
	c.logger.Infof("Installing Message Broker into Kubernetes deployment...")
	err := utils.DeployMsgBroker(c.Config.KubeCli)
	if err != nil {
		c.logger.Errorf("Message Broker installation failed: %v", err)
	} else {
		c.logger.Infof("Message Broker installation finished!")
	}
}

func (c *Controller) Run() error {
	var (
		watchVersion string
		err          error
	)

	watchVersion, err = c.FindResourceVersion()
	if err != nil {
		return err
	}

	c.logger.Infof("Start running Kubeless controller from watch version: %s", watchVersion)
	defer func() {
		c.waitFunction.Wait()
	}()

	//monitor user-defined functions
	eventCh, errCh := c.monitor(watchVersion)

	go func() {
		for event := range eventCh {
			functionName := event.Object.ObjectMeta.Name
			ns := event.Object.ObjectMeta.Namespace
			switch event.Type {
			case "ADDED":
				functionSpec := &event.Object.Spec
				//err := function.New(c.Config.KubeCli, functionName, c.Config.Namespace, functionSpec, &c.waitFunction)
				err := function.New(c.Config.KubeCli, functionName, ns, functionSpec, &c.waitFunction)
				if err != nil {
					break
				}
				c.Functions[functionName] = event.Object
				c.logger.Infof("A new function was added: %s", functionName)

			case "DELETED":
				if c.Functions[functionName] == nil {
					c.logger.Warningf("Ignore deletion: function %q not found", functionName)
					break
				}
				delete(c.Functions, functionName)
				//err := function.Delete(c.Config.KubeCli, functionName, c.Config.Namespace, &c.waitFunction)
				err := function.Delete(c.Config.KubeCli, functionName, ns, &c.waitFunction)
				if err != nil {
					break
				}
				c.logger.Infof("A function was deleted: %s", functionName)
			}
		}
	}()
	return <-errCh
}

func (c *Controller) initResource() error {
	if c.Config.MasterHost == "" {
		return fmt.Errorf("MasterHost is empty. Please check if k8s cluster is available.")
	}
	err := c.createTPR()
	if err != nil {
		if !utils.IsKubernetesResourceAlreadyExistError(err) {
			return fmt.Errorf("Fail to create TPR: %v", err)
		}
	}
	return nil
}

func (c *Controller) FindResourceVersion() (string, error) {
	resp, err := utils.ListResources(c.Config.MasterHost, c.Config.Namespace, c.Config.KubeCli.RESTClient.Client)
	if err != nil {
		return "", err
	}

	d := json.NewDecoder(resp.Body)
	list := &FunctionList{}
	if err := d.Decode(list); err != nil {
		return "", err
	}

	for _, item := range list.Items {
		funcName := item.Name
		c.Functions[funcName] = item
	}
	return list.ListMeta.ResourceVersion, nil
}

func (c *Controller) createTPR() error {
	tpr := &extensions.ThirdPartyResource{
		ObjectMeta: k8sapi.ObjectMeta{
			Name: tprName,
		},
		Versions: []extensions.APIVersion{
			{Name: "v1"},
		},
		Description: "Kubeless: Serverless framework for Kubernetes",
	}
	_, err := c.Config.KubeCli.ThirdPartyResources().Create(tpr)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) monitor(watchVersion string) (<-chan *Event, <-chan error) {
	host := c.Config.MasterHost
	ns := c.Config.Namespace
	httpClient := c.Config.KubeCli.RESTClient.Client

	eventCh := make(chan *Event)
	// On unexpected error case, controller should exit
	errCh := make(chan error, 1)

	go func() {
		defer close(eventCh)
		for {
			resp, err := utils.WatchResources(host, ns, httpClient, watchVersion)
			if err != nil {
				errCh <- err
				return
			}
			if resp.StatusCode != 200 {
				resp.Body.Close()
				errCh <- errors.New("Invalid status code: " + resp.Status)
				return
			}
			c.logger.Infof("Start watching at %v", watchVersion)
			decoder := json.NewDecoder(resp.Body)
			for {
				ev, st, err := pollEvent(decoder)

				if err != nil {
					if err == io.EOF { // apiserver will close stream periodically
						c.logger.Debug("Apiserver closed stream")
						break
					}

					c.logger.Errorf("Received invalid event from API server: %v", err)
					errCh <- err
					return
				}

				if st != nil {
					if st.Code == http.StatusGone { // event history is outdated
						errCh <- ErrVersionOutdated // go to recovery path
						return
					}
					c.logger.Fatalf("Unexpected status response from API server: %v", st.Message)
				}

				c.logger.Debugf("Function event: %v %v", ev.Type, ev.Object.Spec)

				watchVersion = ev.Object.ObjectMeta.ResourceVersion
				eventCh <- ev
			}

			resp.Body.Close()
		}
	}()

	return eventCh, errCh
}

func pollEvent(decoder *json.Decoder) (*Event, *unversionedAPI.Status, error) {
	re := &rawEvent{}
	err := decoder.Decode(re)
	if err != nil {
		if err == io.EOF {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("Fail to decode raw event from apiserver (%v)", err)
	}

	if re.Type == "ERROR" {
		status := &unversionedAPI.Status{}
		err = json.Unmarshal(re.Object, status)
		if err != nil {
			return nil, nil, fmt.Errorf("Fail to decode (%s) into unversioned.Status (%v)", re.Object, err)
		}
		return nil, status, nil
	}

	ev := &Event{
		Type:   re.Type,
		Object: &spec.Function{},
	}
	err = json.Unmarshal(re.Object, ev.Object)
	if err != nil {
		return nil, nil, fmt.Errorf("Fail to unmarshal function object from data (%s): %v", re.Object, err)
	}
	return ev, nil, nil
}

func NewControllerConfig(masterHost, ns string) Config {
	f := utils.GetFactory()
	kubecli, err := f.Client()
	if err != nil {
		fmt.Errorf("Can not get kubernetes config: %s", err)
	}
	if ns == "" {
		ns, _, err = f.DefaultNamespace()
		if err != nil {
			fmt.Errorf("Can not get kubernetes config: %s", err)
		}
	}
	if masterHost == "" {
		k8sConfig, err := f.ClientConfig()
		if err != nil {
			fmt.Errorf("Can not get kubernetes config: %s", err)
		}
		if k8sConfig == nil {
			fmt.Errorf("Got nil k8sConfig, please check if k8s cluster is available.")
		} else {
			masterHost = k8sConfig.Host
		}
	}

	cfg := Config{
		Namespace:  ns,
		KubeCli:    kubecli,
		MasterHost: masterHost,
	}

	return cfg
}