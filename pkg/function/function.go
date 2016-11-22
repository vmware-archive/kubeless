package function

import (
	"github.com/Sirupsen/logrus"
	"github.com/skippbox/kubeless/pkg/utils"
	"github.com/skippbox/kubeless/pkg/spec"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"sync"
)

const (
	defaultVersion                        = "1.0"
	eventDeleteFunction functionEventType = "Delete"
	eventModifyFunction functionEventType = "Modify"
)

type functionEventType string

type functionEvent struct {
	typ  functionEventType
	spec spec.FunctionSpec
}

type Function struct {
	logger    *logrus.Entry
	kclient   *client.Client
	status    *Status
	Spec      *spec.FunctionSpec
	Name      string
	Namespace string
	eventCh   chan *functionEvent
	stopCh    chan struct{}
}

func New(c *client.Client, name, ns string, spec *spec.FunctionSpec, stopC <-chan struct{}, wg *sync.WaitGroup) error {
	return new(c, name, ns, spec, stopC, wg)
}

func Delete(c *client.Client, name, ns string, stopC <-chan struct{}, wg *sync.WaitGroup) error {
	return delete(c, name, ns, stopC, wg)
}

func new(kclient *client.Client, name, ns string, spec *spec.FunctionSpec, stopC <-chan struct{}, wg *sync.WaitGroup) error {
	if len(spec.Version) == 0 {
		// TODO: set version in spec in apiserver
		spec.Version = defaultVersion
	}
	f := &Function{
		logger:    logrus.WithField("pkg", "function").WithField("function-name", name),
		kclient:   kclient,
		Name:      name,
		Namespace: ns,
		eventCh:   make(chan *functionEvent, 100),
		stopCh:    make(chan struct{}),
		Spec:      spec,
		status:    &Status{},
	}

	//TODO: create deployment & svc
	err := utils.CreateK8sResources(f.Namespace, f.Name, f.Spec, kclient)
	if err != nil {
		return err
	}

	wg.Add(1)

	return nil
}

func delete(kclient *client.Client, name, ns string, stopC <-chan struct{}, wg *sync.WaitGroup) error {
	err := utils.DeleteK8sResources(ns, name, kclient)
	wg.Add(1)
	return err
}

func (c *Function) send(ev *functionEvent) {
	select {
	case c.eventCh <- ev:
	case <-c.stopCh:
	default:
		panic("TODO: too many events queued...")
	}
}

func (c *Function) Update(spec *spec.FunctionSpec) {
	anyInterestedChange := false
	if len(spec.Version) == 0 {
		spec.Version = defaultVersion
	}
	if spec.Version != c.Spec.Version {
		anyInterestedChange = true
	}
	if anyInterestedChange {
		c.send(&functionEvent{
			typ:  eventModifyFunction,
			spec: *spec,
		})
	}
}
