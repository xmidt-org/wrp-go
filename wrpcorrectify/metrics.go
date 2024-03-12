// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpcorrectify

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/touchstone"
	"github.com/xmidt-org/wrp-go/v3"
)

// CounterMetric provides a counter metric that can be used to track the number
// of times a specific error occurs.
type CounterMetric struct {
	Factory *touchstone.Factory
	Name    string
	Help    string
	Labels  []string
}

type metricPairs string

var mPairs = metricPairs("pairs")

// Option takes an Option and returns a new Option that increments the counter
// metric if the original Option returns an error.
func (cm CounterMetric) Option(opt Option) Option {
	metric, err := cm.Factory.NewCounterVec(
		prometheus.CounterOpts{
			Name: cm.Name,
			Help: cm.Help,
		},
		cm.Labels...,
	)

	if err != nil {
		return ErrorOption(err)
	}

	return OptionFunc(func(ctx context.Context, m *wrp.Message) error {
		err := opt.Correctify(ctx, m)
		if err != nil {
			pairs := ctx.Value(mPairs).(prometheus.Labels)
			metric.With(pairs).Add(1.0)
		}
		return err
	})
}

// WithMetricLabels returns a new context with the Prometheus labels added so
// that the labels can be used to increment the counter metric.
func WithMetricLabels(ctx context.Context, l prometheus.Labels) context.Context {
	return context.WithValue(ctx, mPairs, l)
}
