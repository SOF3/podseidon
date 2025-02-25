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

package kube

import (
	"context"
	"flag"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/ptr"

	podseidonv1a1 "github.com/kubewharf/podseidon/apis/v1alpha1"
	podseidonclient "github.com/kubewharf/podseidon/client/clientset/versioned"
	podseidonfakeclient "github.com/kubewharf/podseidon/client/clientset/versioned/fake"

	"github.com/kubewharf/podseidon/util/component"
	"github.com/kubewharf/podseidon/util/errors"
	utilflag "github.com/kubewharf/podseidon/util/flag"
	"github.com/kubewharf/podseidon/util/iter"
	kubemetrics "github.com/kubewharf/podseidon/util/kube/metrics"
)

// Api type for `NewClient`.
type Client struct {
	targetNs *string
	state    *ClientState
}

// Request a Kubernetes client config through command line.
var NewClient = component.Declare(
	func(args ClientArgs) string { return fmt.Sprintf("%s-kube", args.ClusterName) },
	func(args ClientArgs, fs *flag.FlagSet) ClientOptions {
		return ClientOptions{
			kubeconfigPath: fs.String(
				"config",
				"",
				fmt.Sprintf("path to %s cluster kubeconfig file", args.ClusterName),
			),
			masterUrl: fs.String(
				"master",
				"",
				fmt.Sprintf("URL to %s cluster apiserver", args.ClusterName),
			),
			TargetNamespace: fs.String(
				"target-namespace",
				metav1.NamespaceAll,
				fmt.Sprintf(
					"namespaces accessible in %s cluster (empty to match all)",
					args.ClusterName,
				),
			),
			impersonateUsername: fs.String(
				"impersonate-username",
				"",
				"username to impersonate as",
			),
			impersonateUid: fs.String("impersonate-uid", "", "UID to impersonate as"),
			impersonateGroups: utilflag.StringSlice(
				fs,
				"impersonate-groups",
				[]string{},
				"comma-separated user groups to impersonate as",
			),
			qps:       fs.Float64("qps", float64(rest.DefaultQPS), "client QPS (for each clientset)"),
			burst:     fs.Int("burst", rest.DefaultBurst, "client burst (for each clientset)"),
			userAgent: fs.String("user-agent", "podseidon", "user agent for the kube client"),
		}
	},
	func(_ ClientArgs, requests *component.DepRequests) ClientDeps {
		component.DepPtr(requests, kubemetrics.NewComp(kubemetrics.CompArgs{}))
		return ClientDeps{}
	},
	func(_ context.Context, _ ClientArgs, options ClientOptions, _ ClientDeps) (*ClientState, error) {
		restConfig, err := clientcmd.BuildConfigFromFlags(
			*options.masterUrl,
			*options.kubeconfigPath,
		)
		if err != nil {
			return nil, errors.TagWrapf("BuildConfig", err, "build rest config from arguments")
		}

		if *options.impersonateUsername != "" {
			restConfig.Impersonate.UserName = *options.impersonateUsername
		}
		if *options.impersonateUid != "" {
			restConfig.Impersonate.UID = *options.impersonateUid
		}
		if len(*options.impersonateGroups) > 0 {
			restConfig.Impersonate.Groups = *options.impersonateGroups
		}

		restConfig.QPS = float32(*options.qps)
		restConfig.Burst = *options.burst

		// Accept protobuf for better pod list performance
		restConfig.AcceptContentTypes = "application/vnd.kubernetes.protobuf,application/json"

		rest.AddUserAgent(restConfig, *options.userAgent)

		kubeClientSet, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			return nil, errors.TagWrapf(
				"NewClientSet",
				err,
				"create new kubernetes client set from config",
			)
		}

		podseidonClientSet, err := podseidonclient.NewForConfig(restConfig)
		if err != nil {
			return nil, errors.TagWrapf(
				"NewClientSet",
				err,
				"create new podseidon client set from config",
			)
		}

		return &ClientState{
			restConfig:         restConfig,
			kubeClientSet:      kubeClientSet,
			podseidonClientSet: podseidonClientSet,
		}, nil
	},
	component.Lifecycle[ClientArgs, ClientOptions, ClientDeps, ClientState]{
		Start:        nil,
		Join:         nil,
		HealthChecks: nil,
	},
	func(d *component.Data[ClientArgs, ClientOptions, ClientDeps, ClientState]) *Client {
		return &Client{targetNs: d.Options.TargetNamespace, state: d.State}
	},
)

type ClusterName string

type ClientArgs struct {
	ClusterName ClusterName
}

type ClientOptions struct {
	kubeconfigPath  *string
	masterUrl       *string
	TargetNamespace *string

	impersonateUsername *string
	impersonateUid      *string
	impersonateGroups   *[]string

	qps   *float64
	burst *int

	userAgent *string
}

type ClientDeps struct{}

type ClientState struct {
	restConfig *rest.Config

	kubeClientSet      kubernetes.Interface
	podseidonClientSet podseidonclient.Interface
}

// This method should only be used for constructing extension client sets for a cluster.
// Not supported by `MockClient`.
func (client *Client) RestConfig() *rest.Config {
	return client.state.restConfig
}

func (client *Client) NativeClientSet() kubernetes.Interface {
	return client.state.kubeClientSet
}

func (client *Client) PodseidonClientSet() podseidonclient.Interface {
	return client.state.podseidonClientSet
}

func (client *Client) TargetNamespace() string {
	return *client.targetNs
}

// Create a mock client with the given objects in the cluster.
func MockClient(objects ...runtime.Object) *Client {
	podseidonObjects := iter.FromSlice(objects).Filter(func(obj runtime.Object) bool {
		return obj.GetObjectKind().
			GroupVersionKind().
			GroupVersion() ==
			podseidonv1a1.SchemeGroupVersion
	}).CollectSlice()
	nativeObjects := iter.FromSlice(objects).Filter(func(obj runtime.Object) bool {
		return obj.GetObjectKind().
			GroupVersionKind().
			GroupVersion() !=
			podseidonv1a1.SchemeGroupVersion
	}).CollectSlice()

	return &Client{
		targetNs: ptr.To(metav1.NamespaceAll),
		state: &ClientState{
			restConfig:         new(rest.Config),
			kubeClientSet:      kubernetesfake.NewSimpleClientset(nativeObjects...),
			podseidonClientSet: podseidonfakeclient.NewSimpleClientset(podseidonObjects...),
		},
	}
}
