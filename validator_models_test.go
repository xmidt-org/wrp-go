// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp_test

import (
	"reflect"
	"strconv"

	"github.com/xmidt-org/wrp-go/v4"
)

// This file defines the different message types so tests can be constructed
// from the models.
//
//   - Match the names of the fields in the models to the names of the fields in
//     Message.  The type is gnored, so use int for all fields.
//   - If the field is not present in the model, it is not allowed.
//   - If the `required:""` tag is present, the field is required, otherwise it
//     is optional.
//   - If the `valid:""` tag is present, the provided value is valid.  Otherwise,
//     any value is valid.
type simpleRequestResponse struct {
	Type                    int `required:"" valid:"3"`
	Source                  int `required:"" valid:"self:/service1/ignored"`
	Destination             int `required:"" valid:"mac:112233445566/service2/ignored"`
	TransactionUUID         int `required:"" valid:"546514d4-9cb6-41c9-88ca-ccd4c130c525"`
	ContentType             int
	Accept                  int
	Status                  int
	RequestDeliveryResponse int
	Headers                 int
	Metadata                int
	Payload                 int
	PartnerIDs              int
	SessionID               int
	QualityOfService        int
}

type simpleEvent struct {
	Type                    int `required:"" valid:"4"`
	Source                  int `required:"" valid:"self:/service1/ignored"`
	Destination             int `required:"" valid:"event:device-status/mac:112233445566/online"`
	TransactionUUID         int `            valid:"546514d4-9cb6-41c9-88ca-ccd4c130c525"`
	ContentType             int
	RequestDeliveryResponse int
	Headers                 int
	Metadata                int
	Payload                 int
	PartnerIDs              int
	SessionID               int
	QualityOfService        int
}

type crud struct {
	Type                    int `required:"" valid:"5"`
	Source                  int `required:"" valid:"self:/service1/ignored"`
	Destination             int `required:"" valid:"mac:112233445566/service2/ignored"`
	TransactionUUID         int `required:"" valid:"546514d4-9cb6-41c9-88ca-ccd4c130c525"`
	ContentType             int
	Accept                  int
	Status                  int
	RequestDeliveryResponse int
	Headers                 int
	Metadata                int
	Payload                 int
	PartnerIDs              int
	Path                    int `            valid:"/path/to/resource"`
	SessionID               int
	QualityOfService        int
}

type authorization struct {
	Type   int `required:"" valid:"2"`
	Status int `required:"" valid:"200"`
}

type serviceRegistration struct {
	Type        int `required:"" valid:"9"`
	ServiceName int `required:"" valid:"service1"`
	URL         int `required:"" valid:"tcp://127.0.0.1:9999"`
}

type serviceAlive struct {
	Type int `required:"" valid:"10"`
}

type unknown struct {
	Type int `required:"" valid:"11"`
}

const (
	randomString = "__random__"
)

func populateRequired(model reflect.Type, msg *wrp.Message) {
	// Set the required fields to their valid values
	for j := 0; j < model.NumField(); j++ {
		field := model.Field(j)
		_, required := field.Tag.Lookup("required")
		if !required {
			continue
		}

		valid, found := field.Tag.Lookup("valid")
		if !found {
			panic("required field must have a valid tag")
		}

		updateMessage(msg, field.Name, valid)
	}
}

func generateManyValidTestCases[T any](model T) []wrp.Message {
	var testCases []wrp.Message

	modelType := reflect.TypeOf(model)

	for i := 0; i < modelType.NumField(); i++ {
		msg := wrp.Message{}

		populateRequired(modelType, &msg)

		for j := 0; j < i; j++ {
			field := modelType.Field(j)
			_, required := field.Tag.Lookup("required")
			if required {
				continue
			}

			updateMessage(&msg, field.Name, randomString)
		}

		testCases = append(testCases, msg)
	}

	return testCases
}

func generateDisallowedFieldsTestCases[T any](model T, what *[]string) []wrp.Message {
	var testCases []wrp.Message
	action := make([]string, 0)

	modelType := reflect.TypeOf(model)
	msgVal := reflect.ValueOf(&wrp.Message{}).Elem()

	for i := 0; i < msgVal.NumField(); i++ {
		msg := wrp.Message{}
		populateRequired(modelType, &msg)

		// Look at the Message type and fill in fields that are not present in
		// the model.
		fn := msgVal.Type().Field(i).Name
		_, found := modelType.FieldByName(fn)
		if found {
			// This field is present in the model, so skip it.
			continue
		}

		updateMessage(&msg, fn, randomString)

		testCases = append(testCases, msg)
		action = append(action, fn)
	}

	*what = action

	return testCases
}

func updateMessage(msg *wrp.Message, fieldName, value string) {
	msgValue := reflect.ValueOf(msg).Elem()

	field := msgValue.FieldByName(fieldName)
	if !field.IsValid() {
		return
	}

	switch field.Kind() {
	case reflect.Int, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			// This lets us use whatever string we want for the value and still
			// have it be valid.
			intValue = 42
		}
		field.SetInt(intValue)
	case reflect.Ptr:
		if field.Type().Elem().Kind() == reflect.Int64 {
			intValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				intValue = 42
			}
			ptrValue := reflect.New(field.Type().Elem())
			ptrValue.Elem().SetInt(intValue)
			field.Set(ptrValue)
		}
	case reflect.String:
		field.SetString(value)
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.Uint8 {
			field.SetBytes([]byte(value))
		} else if field.Type().Elem().Kind() == reflect.String {
			field.Set(reflect.ValueOf([]string{value}))
		}
	case reflect.Map:
		if field.Type().Key().Kind() == reflect.String && field.Type().Elem().Kind() == reflect.String {
			field.Set(reflect.ValueOf(map[string]string{"key": value}))
		}
	}
}
