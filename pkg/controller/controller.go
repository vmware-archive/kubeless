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
	"github.com/skippbox/kubeless/pkg/utils"
	"github.com/skippbox/kubeless/pkg/spec"

	unversionedAPI "k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	k8sapi "k8s.io/kubernetes/pkg/api"
)

const (
	tprName = "lamb-da.k8s.io"
)

var (
	ErrVersionOutdated = errors.New("requested version is outdated in apiserver")
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
	stopChMap    map[string]chan struct{}
	waitFunction sync.WaitGroup
	functions    map[string]*spec.Function
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
		functions: make(map[string]*spec.Function),
		stopChMap: map[string]chan struct{}{},
	}
}

func (c *Controller) Run() error {
	var (
		watchVersion string
		err          error
	)

	for {
		//create TPR, if exists, get the current version
		watchVersion, err = c.initResource()
		if err == nil {
			break
		}
		c.logger.Errorf("initialization failed: %v", err)
		c.logger.Infof("retry in %v...", initRetryWaitTime)
		time.Sleep(initRetryWaitTime)
	}
	c.logger.Infof("starts running Kubeless controller from watch version: %s", watchVersion)
	defer func() {
		for _, stopC := range c.stopChMap {
			close(stopC)
		}
		c.waitFunction.Wait()
	}()

	//monitor user-defined functions
	eventCh, errCh := c.monitor(watchVersion)

	go func() {
		for event := range eventCh {
			functionName := event.Object.ObjectMeta.Name
			switch event.Type {
			case "ADDED":
				functionSpec := &event.Object.Spec
				stopC := make(chan struct{})
				c.stopChMap[functionName] = stopC
				err := function.New(c.Config.KubeCli, functionName, c.Config.Namespace, functionSpec, stopC, &c.waitFunction)
				if err != nil {
					break
				}
				c.functions[functionName] = event.Object
				fmt.Println(c.functions)
				c.logger.Infof("a new function was added: %s", functionName)

			case "DELETED":
				if c.functions[functionName] == nil {
					c.logger.Warningf("ignore deletion: function %q not found (or dead)", functionName)
					break
				}
				stopC := make(chan struct{})
				delete(c.functions, functionName)
				err := function.Delete(c.Config.KubeCli, functionName, c.Config.Namespace, stopC, &c.waitFunction)
				if err != nil {
					break
				}
				fmt.Println(c.functions)
				c.logger.Infof("a function was deleted: %s", functionName)
			}
		}
	}()
	return <-errCh
}

//create TPR, if exists, get the current version
func (c *Controller) initResource() (string, error) {
	watchVersion := "0"
	err := c.createTPR()
	if err != nil {
		if utils.IsKubernetesResourceAlreadyExistError(err) {
			watchVersion, err = c.findResourceVersion()
			if err != nil {
				return "", err
			}
		} else {
			return "", fmt.Errorf("fail to create TPR: %v", err)
		}
	}
	return watchVersion, nil
}

func (c *Controller) findResourceVersion() (string, error) {
	c.logger.Info("finding current resource version...")
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
		c.functions[funcName] = &item
	}
	fmt.Println(c.functions)
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
		Description: "Kubeless: Manage serverless functions in Kubernetes",
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
			c.logger.Infof("start watching at %v", watchVersion)
			decoder := json.NewDecoder(resp.Body)
			for {
				ev, st, err := pollEvent(decoder)

				if err != nil {
					if err == io.EOF { // apiserver will close stream periodically
						c.logger.Debug("apiserver closed stream")
						break
					}

					c.logger.Errorf("received invalid event from API server: %v", err)
					errCh <- err
					return
				}

				if st != nil {
					if st.Code == http.StatusGone { // event history is outdated
						errCh <- ErrVersionOutdated // go to recovery path
						return
					}
					c.logger.Fatalf("unexpected status response from API server: %v", st.Message)
				}

				c.logger.Debugf("function event: %v %v", ev.Type, ev.Object.Spec)

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
		return nil, nil, fmt.Errorf("fail to decode raw event from apiserver (%v)", err)
	}

	if re.Type == "ERROR" {
		status := &unversionedAPI.Status{}
		err = json.Unmarshal(re.Object, status)
		if err != nil {
			return nil, nil, fmt.Errorf("fail to decode (%s) into unversioned.Status (%v)", re.Object, err)
		}
		return nil, status, nil
	}

	ev := &Event{
		Type:   re.Type,
		Object: &spec.Function{},
	}
	err = json.Unmarshal(re.Object, ev.Object)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to unmarshal Function object from data (%s): %v", re.Object, err)
	}
	return ev, nil, nil
}
