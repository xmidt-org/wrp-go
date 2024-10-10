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

	// alwaysInvalidValidatorErrorTotalName is the name of the counter for all AlwaysInvalid validation errors.
	alwaysInvalidValidatorErrorTotalName = metricPrefix + "always_invalid"

	// alwaysInvalidValidatorErrorTotalHelp is the help text for the AlwaysInvalid metric.
	alwaysInvalidValidatorErrorTotalHelp = "the total number of AlwaysInvalid validations"

	// destinationValidatorErrorTotalName is the name of the counter for all destination validation errors.
	destinationValidatorErrorTotalName = metricPrefix + "destination"

	// destinationValidatorErrorTotalHelp is the help text for the Destination metric.
	destinationValidatorErrorTotalHelp = "the total number of Destination validation errors"

	// sourceValidatorErrorTotalName is the name of the counter for all source validation errors.
	sourceValidatorErrorTotalName = metricPrefix + "source"

	// sourceValidatorErrorTotalHelp is the help text for the Source metric.
	sourceValidatorErrorTotalHelp = "the total number of Source validation errors"

	// messageTypeValidatorErrorTotalName is the name of the counter for all MessageType validation errors.
	messageTypeValidatorErrorTotalName = metricPrefix + "message_type"

	// messageTypeValidatorErrorTotalHelp is the help text for the MessageType metric.
	messageTypeValidatorErrorTotalHelp = "the total number of MessageType validation errors"

	// utf8ValidatorErrorTotalName is the name of the counter for all UTF8 validation errors.
	utf8ValidatorErrorTotalName = metricPrefix + "utf8"

	// utf8ValidatorErrorTotalHelp is the help text for the UTF8 Validator metric.
	utf8ValidatorErrorTotalHelp = "the total number of UTF8 validation errors"

	// simpleEventTypeValidatorErrorTotalName is the name of the counter for all SimpleEventType validation errors.
	simpleEventTypeValidatorErrorTotalName = metricPrefix + "simple_event_type"

	// simpleEventTypeValidatorErrorTotalHelp is the help text for the SimpleEventType Validator metric.
	simpleEventTypeValidatorErrorTotalHelp = "the total number of SimpleEventType validation errors"

	// simpleRequestResponseMessageTypeErrorTotalName is the name of the counter for all SimpleRequestResponseMessageType validation errors.
	simpleRequestResponseMessageTypeErrorTotalName = metricPrefix + "simple_request_response_message_type"

	// simpleRequestResponseMessageTypeErrorTotalHelp is the help text for the SimpleRequestResponseMessageType Validator metric.
	simpleRequestResponseMessageTypeErrorTotalHelp = "the total number of SimpleRequestResponseMessageType validation errors"

	// spansValidatorErrorTotalName is the name of the counter for all Spans validation errors.
	spansValidatorErrorTotalName = metricPrefix + "spans"

	// spansValidatorErrorTotalHelp is the help text for the Spans Validator metric.
	spansValidatorErrorTotalHelp = "the total number of Spans validation errors"

	// noneEmptySourceErrorTotalName is the name of the counter for all noneEmptySource validation errors.
	noneEmptySourceErrorTotalName = metricPrefix + "none_empty_source"

	// noneEmptySourceErrorTotalHelp is the help text for the noneEmptySource Validator metric.
	noneEmptySourceErrorTotalHelp = "the total number of None Empty Source validation errors"

	// noneEmptyDestinationErrorTotalName is the name of the counter for all noneEmptyDestination validation errors.
	noneEmptyDestinationErrorTotalName = metricPrefix + "none_empty_destination"

	// noneEmptyDestinationErrorTotalHelp is the help text for the noneEmptyDestination Validator metric.
	noneEmptyDestinationErrorTotalHelp = "the total number of None Empty Destination validation errors"
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

func newNoneEmptySourceErrorTotal(tf *touchstone.Factory, labelNames ...string) (m *prometheus.CounterVec, err error) {
	return tf.NewCounterVec(
		prometheus.CounterOpts{
			Name: noneEmptySourceErrorTotalName,
			Help: noneEmptySourceErrorTotalHelp,
		},
		labelNames...,
	)
}

func newNoneEmptyDestinationErrorTotal(tf *touchstone.Factory, labelNames ...string) (m *prometheus.CounterVec, err error) {
	return tf.NewCounterVec(
		prometheus.CounterOpts{
			Name: noneEmptyDestinationErrorTotalName,
			Help: noneEmptyDestinationErrorTotalHelp,
		},
		labelNames...,
	)
}
