// Copyright 2024 The Podseidon Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/listers"
	"k8s.io/client-go/tools/cache"

	v1alpha1 "github.com/kubewharf/podseidon/apis/v1alpha1"
)

// PodProtectorLister helps list PodProtectors.
// All objects returned here must be treated as read-only.
type PodProtectorLister interface {
	// List lists all PodProtectors in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.PodProtector, err error)
	// PodProtectors returns an object that can list and get PodProtectors.
	PodProtectors(namespace string) PodProtectorNamespaceLister
	PodProtectorListerExpansion
}

// podProtectorLister implements the PodProtectorLister interface.
type podProtectorLister struct {
	listers.ResourceIndexer[*v1alpha1.PodProtector]
}

// NewPodProtectorLister returns a new PodProtectorLister.
func NewPodProtectorLister(indexer cache.Indexer) PodProtectorLister {
	return &podProtectorLister{
		listers.New[*v1alpha1.PodProtector](indexer, v1alpha1.Resource("podprotector")),
	}
}

// PodProtectors returns an object that can list and get PodProtectors.
func (s *podProtectorLister) PodProtectors(namespace string) PodProtectorNamespaceLister {
	return podProtectorNamespaceLister{
		listers.NewNamespaced[*v1alpha1.PodProtector](s.ResourceIndexer, namespace),
	}
}

// PodProtectorNamespaceLister helps list and get PodProtectors.
// All objects returned here must be treated as read-only.
type PodProtectorNamespaceLister interface {
	// List lists all PodProtectors in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.PodProtector, err error)
	// Get retrieves the PodProtector from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.PodProtector, error)
	PodProtectorNamespaceListerExpansion
}

// podProtectorNamespaceLister implements the PodProtectorNamespaceLister
// interface.
type podProtectorNamespaceLister struct {
	listers.ResourceIndexer[*v1alpha1.PodProtector]
}
