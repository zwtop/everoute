/*
Copyright 2021 The Lynx Authors.

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

package informer

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"k8s.io/client-go/tools/cache"

	"github.com/smartxworks/lynx/plugin/tower/pkg/client"
	"github.com/smartxworks/lynx/plugin/tower/pkg/schema"
	"github.com/smartxworks/lynx/plugin/tower/third_party/forked/client-go/informer"
)

// SharedInformerFactory provides shared informers for all resources
type SharedInformerFactory interface {
	// Start initializes all requested informers
	Start(stopCh <-chan struct{})
	// WaitForCacheSync waits for all started informers' cache were synced
	WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool

	// VM return informer for &schema.VM{}
	VM() cache.SharedIndexInformer
	// Label return informer for &schema.Label{}
	Label() cache.SharedIndexInformer
}

// NewSharedInformerFactory constructs a new instance of sharedInformerFactory for all resources
func NewSharedInformerFactory(client *client.Client, defaultResync time.Duration) SharedInformerFactory {
	factory := &sharedInformerFactory{
		client:           client,
		defaultResync:    defaultResync,
		informers:        make(map[reflect.Type]cache.SharedIndexInformer),
		startedInformers: make(map[reflect.Type]bool),
		customResync:     make(map[reflect.Type]time.Duration),
	}
	return factory
}

type sharedInformerFactory struct {
	client        *client.Client
	lock          sync.Mutex
	defaultResync time.Duration
	customResync  map[reflect.Type]time.Duration

	informers map[reflect.Type]cache.SharedIndexInformer
	// startedInformers is used for tracking which informers have been started.
	// This allows Start() to be called multiple times safely.
	startedInformers map[reflect.Type]bool
}

// Start implements SharedInformerFactory.Start
func (f *sharedInformerFactory) Start(stopCh <-chan struct{}) {
	f.lock.Lock()
	defer f.lock.Unlock()

	for informerType, shareInformer := range f.informers {
		if !f.startedInformers[informerType] {
			go shareInformer.Run(stopCh)
			f.startedInformers[informerType] = true
		}
	}
}

// WaitForCacheSync implements SharedInformerFactory.WaitForCacheSync
func (f *sharedInformerFactory) WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool {
	informers := func() map[reflect.Type]cache.SharedIndexInformer {
		f.lock.Lock()
		defer f.lock.Unlock()

		informers := map[reflect.Type]cache.SharedIndexInformer{}
		for informerType, shareInformer := range f.informers {
			if f.startedInformers[informerType] {
				informers[informerType] = shareInformer
			}
		}
		return informers
	}()

	res := map[reflect.Type]bool{}
	for informType, shareInformer := range informers {
		res[informType] = cache.WaitForCacheSync(stopCh, shareInformer.HasSynced)
	}
	return res
}

// VM implements SharedInformerFactory.VM
func (f *sharedInformerFactory) VM() cache.SharedIndexInformer {
	return f.informerFor(&schema.VM{})
}

// Label implements SharedInformerFactory.Label
func (f *sharedInformerFactory) Label() cache.SharedIndexInformer {
	return f.informerFor(&schema.Label{})
}

func (f *sharedInformerFactory) informerFor(obj schema.Object) cache.SharedIndexInformer {
	f.lock.Lock()
	defer f.lock.Unlock()

	informerType := reflect.TypeOf(obj)
	shareInformer, exists := f.informers[informerType]
	if exists {
		return shareInformer
	}

	resyncPeriod, exists := f.customResync[informerType]
	if !exists {
		resyncPeriod = f.defaultResync
	}

	shareInformer = informer.NewSharedIndexInformer(NewReflectorBuilder(f.client), obj, towerObjectKey, resyncPeriod, nil)
	f.informers[informerType] = shareInformer

	return shareInformer
}

func towerObjectKey(obj interface{}) (string, error) {
	resource, ok := obj.(schema.Object)
	if !ok {
		return "", fmt.Errorf("unsupport resource type %s, object: %v", obj, obj)
	}
	return resource.GetID(), nil
}
