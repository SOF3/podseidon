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

package observer

import (
	"context"
	"time"

	"github.com/kubewharf/podseidon/util/component"
	"github.com/kubewharf/podseidon/util/errors"
	"github.com/kubewharf/podseidon/util/o11y"
	"github.com/kubewharf/podseidon/util/o11y/metrics"
	"github.com/kubewharf/podseidon/util/util"
)

func ProvideMetrics() component.Declared[Observer] {
	return o11y.Provide(
		metrics.MakeObserverDeps,
		func(deps metrics.ObserverDeps) Observer {
			type aggregatorReconcileTags struct {
				Error string
			}

			type aggregatorEnqueueTags struct {
				Kind string
			}

			type aggregatorEnqueueCtxKey struct{}

			type aggregatorEnqueueCtxValue struct {
				kind  string
				start time.Time
			}

			type reconcileStartTime struct{}

			reconcileHandle := metrics.Register(
				deps.Registry(),
				"aggregator_reconcile",
				"Duration of an aggregator reconcile run for a PodProtector.",
				metrics.FunctionDurationHistogram(),
				metrics.NewReflectTags[aggregatorReconcileTags](),
			)

			enqueueHandle := metrics.Register(
				deps.Registry(),
				"aggregator_enqueue",
				"Duration of an aggregator pod informer event handler run.",
				metrics.FunctionDurationHistogram(),
				metrics.NewReflectTags[aggregatorEnqueueTags](),
			)

			indexErrorHandle := metrics.Register(
				deps.Registry(),
				"aggregator_index_error",
				"Number of error events during pod informer event handling.",
				metrics.IntCounter(),
				metrics.NewReflectTags[aggregatorReconcileTags](),
			)

			nextEventPoolCurrentSizeHandle := metrics.Register(
				deps.Registry(),
				"aggregator_next_event_pool_current_size",
				"Current size of the aggregator NextEventPool (during sample).",
				metrics.IntGauge(),
				metrics.NewReflectTags[util.Empty](),
			).With(util.Empty{})
			nextEventPoolCurrentLatencyHandle := metrics.Register(
				deps.Registry(),
				"aggregator_next_event_pool_current_latency",
				"Time since the oldest current object in the NextEventPool (during sample).",
				metrics.DurationGauge(),
				metrics.NewReflectTags[util.Empty](),
			).With(util.Empty{})
			nextEventPoolDrainSize := metrics.Register(
				deps.Registry(),
				"aggregator_next_event_pool_drain_count",
				"Number of items in NextEventPool each time it is drained.",
				metrics.ExponentialIntHistogram(16),
				metrics.NewReflectTags[util.Empty](),
			)
			nextEventPoolDrainPeriod := metrics.Register(
				deps.Registry(),
				"aggregator_next_event_pool_drain_period",
				"Period between NextEventPool drains.",
				metrics.AsyncLatencyDurationHistogram(),
				metrics.NewReflectTags[util.Empty](),
			)
			nextEventPoolDrainObjectLatency := metrics.Register(
				deps.Registry(),
				"aggregator_next_event_pool_drain_object_latency",
				"Time since the oldest object in NextEventPool during each drain.",
				metrics.AsyncLatencyDurationHistogram(),
				metrics.NewReflectTags[util.Empty](),
			)

			return Observer{
				StartReconcile: func(ctx context.Context, _ StartReconcile) (context.Context, context.CancelFunc) {
					ctx = context.WithValue(ctx, reconcileStartTime{}, time.Now())
					return ctx, util.NoOp
				},
				EndReconcile: func(ctx context.Context, arg EndReconcile) {
					duration := time.Since(ctx.Value(reconcileStartTime{}).(time.Time))

					reconcileHandle.Emit(duration, aggregatorReconcileTags{
						Error: errors.SerializeTags(arg.Err),
					})
				},
				StartEnqueue: func(ctx context.Context, arg StartEnqueue) (context.Context, context.CancelFunc) {
					ctx = context.WithValue(
						ctx,
						aggregatorEnqueueCtxKey{},
						aggregatorEnqueueCtxValue{
							kind:  arg.Kind,
							start: time.Now(),
						},
					)

					return ctx, util.NoOp
				},
				EndEnqueue: func(ctx context.Context, _ EndEnqueue) {
					ctxValue := ctx.Value(aggregatorEnqueueCtxKey{}).(aggregatorEnqueueCtxValue)
					duration := time.Since(ctxValue.start)

					enqueueHandle.Emit(duration, aggregatorEnqueueTags{
						Kind: ctxValue.kind,
					})
				},
				EnqueueError: func(_ context.Context, arg EnqueueError) {
					indexErrorHandle.Emit(1, aggregatorReconcileTags{
						Error: errors.SerializeTags(arg.Err),
					})
				},
				Aggregated: func(_ context.Context, _ Aggregated) {
					// TODO add metrics when more fields are available
				},
				NextEventPoolCurrentSize: func(ctx context.Context, _ util.Empty, getter func() int) {
					metrics.Repeating(ctx, deps, nextEventPoolCurrentSizeHandle, getter)
				},
				NextEventPoolCurrentLatency: func(ctx context.Context, _ util.Empty, getter func() time.Duration) {
					metrics.Repeating(
						ctx,
						deps,
						nextEventPoolCurrentLatencyHandle,
						getter,
					)
				},
				NextEventPoolSingleDrain: func(_ context.Context, arg NextEventPoolSingleDrain) {
					nextEventPoolDrainSize.Emit(arg.Size, util.Empty{})
					nextEventPoolDrainObjectLatency.Emit(arg.ObjectLatency, util.Empty{})
					if timeSinceLastDrain, present := arg.TimeSinceLastDrain.Get(); present {
						nextEventPoolDrainPeriod.Emit(timeSinceLastDrain, util.Empty{})
					}
				},
			}
		},
	)
}
