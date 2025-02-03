// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

//go:generate stringer -type=QOSLevel -linecomment

// QOSLevel is the quality of service level associated with a WRP message.
type QOSLevel int

const (
	QOSLow      QOSLevel = iota // Low
	QOSMedium                   // Medium
	QOSHigh                     // High
	QOSCritical                 // Critical
)

// QOSValue is the quality of service value set in a WRP message.  Values of this
// type determine what QOSLevel a message has.
type QOSValue int

const (
	QOSLowValue QOSValue = iota * 25
	QOSMediumValue
	QOSHighValue
	QOSCriticalValue
)

// Level determines the QOSLevel for this value.  Negative values are assumed
// to be QOSLow.  Values higher than the highest value (99) are assumed to
// be QOSCritical.
func (qv QOSValue) Level() QOSLevel {
	switch {
	case qv < 25:
		return QOSLow

	case qv < 50:
		return QOSMedium

	case qv < 75:
		return QOSHigh

	default:
		return QOSCritical
	}
}

// Valid determines if the QOSValue is valid.  Valid values are between 0 and 99,
// inclusive.
func (qv QOSValue) Valid() bool {
	return 0 <= qv && qv <= 99
}
