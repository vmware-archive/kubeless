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

package handlers

import (
	"github.com/skippbox/kubeless/config"
	"github.com/skippbox/kubeless/pkg/event"
	"fmt"
	"log"
)

// Handler is implemented by any handler.
// The Handle method is used to process event
type Handler interface {
	Init(c *config.Config) error
	ObjectCreated(obj interface{})
	ObjectDeleted(obj interface{})
	ObjectUpdated(oldObj, newObj interface{})
}

// Map maps each event handler function to a name for easily lookup
var Map = map[string]interface{}{
	"default": &Default{},
}

// Default handler implements Handler interface,
// print each event with JSON format
type Default struct {
}

// Init initializes handler configuration
// Do nothing for default handler
func (d *Default) Init(c *config.Config) error {
	return nil
}

func (d *Default) ObjectCreated(obj interface{}) {
	notify(obj, "created")
}

func (d *Default) ObjectDeleted(obj interface{}) {
	notify(obj, "deleted")
}

func (d *Default) ObjectUpdated(oldObj, newObj interface{}) {
	notify(newObj, "updated")
}

func notify(obj interface{}, action string) {
	e := event.New(obj, action)
	attachment := prepareMessage(e)
	log.Printf(attachment)
}

func prepareMessage(e event.Event) string {
	msg := fmt.Sprintf(
		"A %s in namespace %s has been %s: %s",
		e.Kind,
		e.Namespace,
		e.Reason,
		e.Name,
	)

	return msg
}