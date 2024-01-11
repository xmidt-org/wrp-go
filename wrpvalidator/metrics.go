// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpvalidator

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/touchstone"
)

const (
	// MetricPrefix is prepended to all metrics exposed by this package.
	metricPrefix = "wrp_"

	// utf8ValidatorName is the name of the counter for all UTF8 validation.
	utf8ValidatorName = metricPrefix + "utf8_validator"

	// utf8ValidatorHelp is the help text for the UTF8 Validator metric.
	utf8ValidatorHelp = "the total number of UTF8 Validator metric"
)

func NewUTF8ValidatorErrorTotal(f *touchstone.Factory, labelNames ...string) (m *prometheus.CounterVec, err error) {
	return f.NewCounterVec(
		prometheus.CounterOpts{
			Name: utf8ValidatorName,
			Help: utf8ValidatorHelp,
		},
		labelNames...
	)
}
