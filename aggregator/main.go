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

package main

import (
	"k8s.io/utils/clock"

	"github.com/kubewharf/podseidon/util/cmd"
	"github.com/kubewharf/podseidon/util/component"
	healthzobserver "github.com/kubewharf/podseidon/util/healthz/observer"
	kubeobserver "github.com/kubewharf/podseidon/util/kube/observer"
	"github.com/kubewharf/podseidon/util/o11y/metrics"
	pprutil "github.com/kubewharf/podseidon/util/podprotector"
	pprutilobserver "github.com/kubewharf/podseidon/util/podprotector/observer"
	"github.com/kubewharf/podseidon/util/pprof"
	"github.com/kubewharf/podseidon/util/util"
	workerobserver "github.com/kubewharf/podseidon/util/worker/observer"

	"github.com/kubewharf/podseidon/aggregator/aggregator"
	aggregatorobserver "github.com/kubewharf/podseidon/aggregator/observer"
	"github.com/kubewharf/podseidon/aggregator/synctime"
	"github.com/kubewharf/podseidon/aggregator/updatetrigger"
)

func main() {
	cmd.Run(
		component.RequireDep(pprof.New(util.Empty{})),
		component.RequireDep(metrics.NewHttp(metrics.HttpArgs{})),
		healthzobserver.Provide,
		workerobserver.Provide,
		kubeobserver.ProvideElector,
		aggregatorobserver.Provide,
		pprutilobserver.ProvideInformer,
		component.RequireDep(aggregator.NewController(aggregator.ControllerArgs{
			Clock:          clock.RealClock{},
			SourceProvider: pprutil.RequestSingleSourceProvider("core"),
		})),
		synctime.DefaultImpls,
		component.RequireDep(updatetrigger.New(updatetrigger.Args{})),
	)
}
