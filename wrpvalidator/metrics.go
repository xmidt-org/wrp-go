// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpvalidator

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/touchstone"
)

const (
	// MetricPrefix is prepended to all metrics exposed by this package.
	metricPrefix = "wrp_validator_"

	// alwaysInvalidValidatorErrorTotalName is the name of the counter for all AlwaysInvalid validation.
	alwaysInvalidValidatorErrorTotalName = metricPrefix + "always_invalid"

	// alwaysInvalidValidatorErrorTotalHelp is the help text for the AlwaysInvalid metric.
	alwaysInvalidValidatorErrorTotalHelp = "the total number of AlwaysInvalid validations"

	// destinationValidatorErrorTotalName is the name of the counter for all destination validation.
	destinationValidatorErrorTotalName = metricPrefix + "destination"

	// destinationValidatorErrorTotalHelp is the help text for the DestinationValidator metric.
	destinationValidatorErrorTotalHelp = "the total number of DestinationValidator metric"

	// sourceValidatorErrorTotalName is the name of the counter for all source validation.
	sourceValidatorErrorTotalName = metricPrefix + "source"

	// sourceValidatorErrorTotalHelp is the help text for the SourceValidator metric.
	sourceValidatorErrorTotalHelp = "the total number of SourceValidator metric"

	// messageTypeValidatorErrorTotalName is the name of the counter for all MessageType validation.
	messageTypeValidatorErrorTotalName = metricPrefix + "message_type"

	// messageTypeValidatorErrorTotalHelp is the help text for the MessageTypeValidator metric.
	messageTypeValidatorErrorTotalHelp = "the total number of MessageTypeValidator metric"

	// utf8ValidatorErrorTotalName is the name of the counter for all UTF8 validation.
	utf8ValidatorErrorTotalName = metricPrefix + "utf8"

	// utf8ValidatorErrorTotalHelp is the help text for the UTF8 Validator metric.
	utf8ValidatorErrorTotalHelp = "the total number of UTF8 Validator metric"

	// simpleEventTypeValidatorErrorTotalName is the name of the counter for all SimpleEventType validation.
	simpleEventTypeValidatorErrorTotalName = metricPrefix + "simple_event_type"

	// simpleEventTypeValidatorErrorTotalHelp is the help text for the SimpleEventType Validator metric.
	simpleEventTypeValidatorErrorTotalHelp = "the total number of SimpleEventType Validator metric"

	// simpleRequestResponseMessageTypeValidatorErrorTotalName is the name of the counter for all SimpleRequestResponseMessageType validation.
	simpleRequestResponseMessageTypeValidatorErrorTotalName = metricPrefix + "simple_request_response_message_type"

	// simpleRequestResponseMessageTypeValidatorErrorTotalHelp is the help text for the SimpleRequestResponseMessageType Validator metric.
	simpleRequestResponseMessageTypeValidatorErrorTotalHelp = "the total number of SimpleRequestResponseMessageType Validator metric"

	// spansValidatorErrorTotalName is the name of the counter for all Spans validation.
	spansValidatorErrorTotalName = metricPrefix + "spans"

	// spansValidatorErrorTotalHelp is the help text for the Spans Validator metric.
	spansValidatorErrorTotalHelp = "the total number of Spans Validator metric"
)

// Metric label names
const (
	PartnerIDLabel   = "partner_id"
	MessageTypeLabel = "message_type"
	ClientIDLabel    = "client_id"
)

func newAlwaysInvalidValidatorErrorTotal(f *touchstone.Factory, labelNames ...string) (m *prometheus.CounterVec, err error) {
	return f.NewCounterVec(
		prometheus.CounterOpts{
			Name: alwaysInvalidValidatorErrorTotalName,
			Help: alwaysInvalidValidatorErrorTotalHelp,
		},
		labelNames...,
	)
}

func newDestinationValidatorErrorTotal(f *touchstone.Factory, labelNames ...string) (m *prometheus.CounterVec, err error) {
	return f.NewCounterVec(
		prometheus.CounterOpts{
			Name: destinationValidatorErrorTotalName,
			Help: destinationValidatorErrorTotalHelp,
		},
		labelNames...,
	)
}

func newSourceValidatorErrorTotal(f *touchstone.Factory, labelNames ...string) (m *prometheus.CounterVec, err error) {
	return f.NewCounterVec(
		prometheus.CounterOpts{
			Name: sourceValidatorErrorTotalName,
			Help: sourceValidatorErrorTotalHelp,
		},
		labelNames...,
	)
}

func newMessageTypeValidatorErrorTotal(f *touchstone.Factory, labelNames ...string) (m *prometheus.CounterVec, err error) {
	return f.NewCounterVec(
		prometheus.CounterOpts{
			Name: messageTypeValidatorErrorTotalName,
			Help: messageTypeValidatorErrorTotalHelp,
		},
		labelNames...,
	)
}

func newUTF8ValidatorErrorTotal(f *touchstone.Factory, labelNames ...string) (m *prometheus.CounterVec, err error) {
	return f.NewCounterVec(
		prometheus.CounterOpts{
			Name: utf8ValidatorErrorTotalName,
			Help: utf8ValidatorErrorTotalHelp,
		},
		labelNames...,
	)
}

func newSimpleEventTypeValidatorErrorTotal(f *touchstone.Factory, labelNames ...string) (m *prometheus.CounterVec, err error) {
	return f.NewCounterVec(
		prometheus.CounterOpts{
			Name: simpleEventTypeValidatorErrorTotalName,
			Help: simpleEventTypeValidatorErrorTotalHelp,
		},
		labelNames...,
	)
}

func newSimpleRequestResponseMessageTypeValidatorErrorTotal(f *touchstone.Factory, labelNames ...string) (m *prometheus.CounterVec, err error) {
	return f.NewCounterVec(
		prometheus.CounterOpts{
			Name: simpleRequestResponseMessageTypeValidatorErrorTotalName,
			Help: simpleRequestResponseMessageTypeValidatorErrorTotalHelp,
		},
		labelNames...,
	)
}

func newSpansValidatorErrorTotal(f *touchstone.Factory, labelNames ...string) (m *prometheus.CounterVec, err error) {
	return f.NewCounterVec(
		prometheus.CounterOpts{
			Name: spansValidatorErrorTotalName,
			Help: spansValidatorErrorTotalHelp,
		},
		labelNames...,
	)
}
