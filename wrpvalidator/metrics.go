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

	// destinationValidatorErrorTotalHelp is the help text for the Destination metric.
	destinationValidatorErrorTotalHelp = "the total number of Destination metric"

	// sourceValidatorErrorTotalName is the name of the counter for all source validation.
	sourceValidatorErrorTotalName = metricPrefix + "source"

	// sourceValidatorErrorTotalHelp is the help text for the Source metric.
	sourceValidatorErrorTotalHelp = "the total number of Source metric"

	// messageTypeValidatorErrorTotalName is the name of the counter for all MessageType validation.
	messageTypeValidatorErrorTotalName = metricPrefix + "message_type"

	// messageTypeValidatorErrorTotalHelp is the help text for the MessageType metric.
	messageTypeValidatorErrorTotalHelp = "the total number of MessageType metric"

	// utf8ValidatorErrorTotalName is the name of the counter for all UTF8 validation.
	utf8ValidatorErrorTotalName = metricPrefix + "utf8"

	// utf8ValidatorErrorTotalHelp is the help text for the UTF8 Validator metric.
	utf8ValidatorErrorTotalHelp = "the total number of UTF8 Validator metric"

	// simpleEventTypeValidatorErrorTotalName is the name of the counter for all SimpleEventType validation.
	simpleEventTypeValidatorErrorTotalName = metricPrefix + "simple_event_type"

	// simpleEventTypeValidatorErrorTotalHelp is the help text for the SimpleEventType Validator metric.
	simpleEventTypeValidatorErrorTotalHelp = "the total number of SimpleEventType Validator metric"

	// simpleRequestResponseMessageTypeErrorTotalName is the name of the counter for all SimpleRequestResponseMessageType validation.
	simpleRequestResponseMessageTypeErrorTotalName = metricPrefix + "simple_request_response_message_type"

	// simpleRequestResponseMessageTypeErrorTotalHelp is the help text for the SimpleRequestResponseMessageType Validator metric.
	simpleRequestResponseMessageTypeErrorTotalHelp = "the total number of SimpleRequestResponseMessageType Validator metric"

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

func newAlwaysInvalidErrorTotal(tf *touchstone.Factory, labelNames ...string) (m *prometheus.CounterVec, err error) {
	return tf.NewCounterVec(
		prometheus.CounterOpts{
			Name: alwaysInvalidValidatorErrorTotalName,
			Help: alwaysInvalidValidatorErrorTotalHelp,
		},
		labelNames...,
	)
}

func newDestinationErrorTotal(tf *touchstone.Factory, labelNames ...string) (m *prometheus.CounterVec, err error) {
	return tf.NewCounterVec(
		prometheus.CounterOpts{
			Name: destinationValidatorErrorTotalName,
			Help: destinationValidatorErrorTotalHelp,
		},
		labelNames...,
	)
}

func newSourceErrorTotal(tf *touchstone.Factory, labelNames ...string) (m *prometheus.CounterVec, err error) {
	return tf.NewCounterVec(
		prometheus.CounterOpts{
			Name: sourceValidatorErrorTotalName,
			Help: sourceValidatorErrorTotalHelp,
		},
		labelNames...,
	)
}

func newMessageTypeErrorTotal(tf *touchstone.Factory, labelNames ...string) (m *prometheus.CounterVec, err error) {
	return tf.NewCounterVec(
		prometheus.CounterOpts{
			Name: messageTypeValidatorErrorTotalName,
			Help: messageTypeValidatorErrorTotalHelp,
		},
		labelNames...,
	)
}

func newUTF8ErrorTotal(tf *touchstone.Factory, labelNames ...string) (m *prometheus.CounterVec, err error) {
	return tf.NewCounterVec(
		prometheus.CounterOpts{
			Name: utf8ValidatorErrorTotalName,
			Help: utf8ValidatorErrorTotalHelp,
		},
		labelNames...,
	)
}

func newSimpleEventTypeErrorTotal(tf *touchstone.Factory, labelNames ...string) (m *prometheus.CounterVec, err error) {
	return tf.NewCounterVec(
		prometheus.CounterOpts{
			Name: simpleEventTypeValidatorErrorTotalName,
			Help: simpleEventTypeValidatorErrorTotalHelp,
		},
		labelNames...,
	)
}

func newSimpleRequestResponseMessageTypeErrorTotal(tf *touchstone.Factory, labelNames ...string) (m *prometheus.CounterVec, err error) {
	return tf.NewCounterVec(
		prometheus.CounterOpts{
			Name: simpleRequestResponseMessageTypeErrorTotalName,
			Help: simpleRequestResponseMessageTypeErrorTotalHelp,
		},
		labelNames...,
	)
}

func newSpansErrorTotal(tf *touchstone.Factory, labelNames ...string) (m *prometheus.CounterVec, err error) {
	return tf.NewCounterVec(
		prometheus.CounterOpts{
			Name: spansValidatorErrorTotalName,
			Help: spansValidatorErrorTotalHelp,
		},
		labelNames...,
	)
}
