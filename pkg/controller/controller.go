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
	"github.com/bitnami/kubeless/pkg/function"
	"github.com/bitnami/kubeless/pkg/spec"
	"github.com/bitnami/kubeless/pkg/utils"

	"k8s.io/client-go/kubernetes"
	k8sErrors "k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
)

const (
	tprName = "function.k8s.io"
)

var (
	errVersionOutdated = errors.New("Requested version is outdated in apiserver")
	initRetryWaitTime  = 30 * time.Second
)

type rawEvent struct {
	Type   string
	Object json.RawMessage
}

// Event object
type Event struct {
	Type   string
	Object *spec.Function
}

// Controller object
type Controller struct {
	logger       *logrus.Entry
	Config       Config
	waitFunction sync.WaitGroup
	Functions    map[string]*spec.Function
}

// Config contains k8s client of a controller
type Config struct {
	KubeCli *kubernetes.Clientset
}

// New initializes a controller object
func New(cfg Config) *Controller {
	return &Controller{
		logger:    logrus.WithField("pkg", "controller"),
		Config:    cfg,
		Functions: make(map[string]*spec.Function),
	}
}

// Init creates tpr functions.k8s.io
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

// InstallKubeless deploys kubeless-controller
func (c *Controller) InstallKubeless(ctlImage string, ctlNamespace string) {
	c.logger.Infof("Installing Kubeless controller into Kubernetes deployment...")
	err := utils.DeployKubeless(c.Config.KubeCli, ctlImage, ctlNamespace)
	if err != nil {
		c.logger.Errorf("Kubeless controller installation failed: %v", err)
	} else {
		c.logger.Infof("Kubeless controller installation successful!")
	}
}

// InstallMsgBroker deploys kafka-controller
func (c *Controller) InstallMsgBroker(kafkaVer string, ctlNamespace string) {
	c.logger.Infof("Installing Message Broker into Kubernetes deployment...")
	err := utils.DeployMsgBroker(c.Config.KubeCli, kafkaVer, ctlNamespace)
	if err != nil {
		c.logger.Errorf("Message Broker installation failed: %v", err)
	} else {
		c.logger.Infof("Message Broker installation successful!")
	}
}

// Run starts the kubeless controller
func (c *Controller) Run() error {
	var (
		watchVersion string
		err          error
	)

	// make a new config for the extension's API group, using the first config as a baseline
	tprClient, err := utils.GetTPRClient()
	if err != nil {
		return err
	}

	watchVersion, err = c.FindResourceVersion(tprClient)
	if err != nil {
		return err
	}

	c.logger.Infof("Start running Kubeless controller from watch version: %s", watchVersion)
	defer func() {
		c.waitFunction.Wait()
	}()

	//monitor user-defined functions
	eventCh := c.monitor(tprClient.Client, watchVersion)

	go func() {
		for event := range eventCh {
			functionName := event.Object.Metadata.Name
			ns := event.Object.Metadata.Namespace
			switch event.Type {
			case "ADDED":
				functionSpec := &event.Object.Spec
				err := function.New(c.Config.KubeCli, functionName, ns, functionSpec, &c.waitFunction)
				if err != nil {
					c.logger.Errorf("A new function is detected but can't be added: %v", err)
					break
				}
				c.Functions[functionName+"."+ns] = event.Object
				c.logger.Infof("A new function was added: %s", functionName)

			case "DELETED":
				if c.Functions[functionName+"."+ns] == nil {
					c.logger.Warningf("Ignore deletion: function %q not found", functionName)
					break
				}
				delete(c.Functions, functionName)
				err := function.Delete(c.Config.KubeCli, functionName, ns, &c.waitFunction)
				if err != nil {
					c.logger.Errorf("Can't delete function: %v", err)
					break
				}
				c.logger.Infof("A function was deleted: %s", functionName)
			}
		}
	}()

	return nil
}

func (c *Controller) initResource() error {
	_, err := c.Config.KubeCli.Extensions().ThirdPartyResources().Get(tprName)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			tpr := &v1beta1.ThirdPartyResource{
				ObjectMeta: v1.ObjectMeta{
					Name: tprName,
				},
				Versions: []v1beta1.APIVersion{
					{Name: "v1"},
				},
				Description: "Kubeless: Serverless framework for Kubernetes",
			}

			_, err := c.Config.KubeCli.Extensions().ThirdPartyResources().Create(tpr)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		fmt.Println("The functions.k8s.io tpr already exists")
	}

	return nil
}

// FindResourceVersion looks up the current resource version
func (c *Controller) FindResourceVersion(tprClient *rest.RESTClient) (string, error) {
	list := spec.FunctionList{}
	err := tprClient.Get().Resource("functions").Do().Into(&list)
	if err != nil {
		return "", err
	}

	for _, item := range list.Items {
		funcName := item.Metadata.Name + "." + item.Metadata.Namespace
		c.Functions[funcName] = item
	}

	return list.Metadata.ResourceVersion, nil
}

// monitor continuously watches for changes to custom function objects
func (c *Controller) monitor(httpClient *http.Client, watchVersion string) <-chan *Event {
	eventCh := make(chan *Event)

	go func() {
		defer close(eventCh)

		// per-watch loop: start watching resources and collecting functions (custom objects)
		for {
			resp, err := utils.WatchResources(httpClient, watchVersion)
			if err != nil {
				c.logger.Errorf("Fail to watch resources: %v. Try again", err)
				// go to the next round
				continue
			}
			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				c.logger.Errorf("Invalid status code: %s. Try again", resp.Status)
				// go to the next round
				continue
			}
			c.logger.Infof("Start watching at %v", watchVersion)
			decoder := json.NewDecoder(resp.Body)

			// per-event loop: pick up function and put to eventCh channel
			for {
				ev, st, err := pollEvent(decoder)

				if err != nil {
					if err == io.EOF { // apiserver will close stream periodically
						c.logger.Errorf("Apiserver closed stream. Try to watch resource again")
						// go to the next watch loop
						break
					}

					c.logger.Errorf("Received invalid event from API server: %v. Go to next event", err)
					// polling next event
					continue
				}

				if st != nil {
					if st.Code == http.StatusGone { // event history is outdated
						err = errVersionOutdated // go to recovery path
						// keep polling next event
						continue
					}
					c.logger.Errorf("Unexpected status response from API server: %v", st.Message)
				}

				c.logger.Debugf("Function event: %v %v", ev.Type, ev.Object.Spec)

				watchVersion = ev.Object.Metadata.ResourceVersion
				eventCh <- ev
			}

			resp.Body.Close()
		}
	}()

	return eventCh
}

func pollEvent(decoder *json.Decoder) (*Event, *unversioned.Status, error) {
	re := &rawEvent{}
	err := decoder.Decode(re)
	if err != nil {
		if err == io.EOF {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("Fail to decode raw event from apiserver (%v)", err)
	}

	if re.Type == "ERROR" {
		status := &unversioned.Status{}
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
