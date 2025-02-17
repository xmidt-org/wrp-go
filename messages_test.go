// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// allFormats enumerates all of the supported formats to use in testing
	allFormats = []Format{JSON, Msgpack}
)

func TestMessageSetStatus(t *testing.T) {
	var (
		assert  = assert.New(t)
		message Message
	)

	assert.Nil(message.Status)
	assert.True(&message == message.SetStatus(72))
	assert.NotNil(message.Status)
	assert.Equal(int64(72), *message.Status)
	assert.True(&message == message.SetStatus(6))
	assert.NotNil(message.Status)
	assert.Equal(int64(6), *message.Status)
}

func TestMessageSetRequestDeliveryResponse(t *testing.T) {
	var (
		assert  = assert.New(t)
		message Message
	)

	assert.Nil(message.RequestDeliveryResponse)
	assert.True(&message == message.SetRequestDeliveryResponse(14))
	assert.NotNil(message.RequestDeliveryResponse)
	assert.Equal(int64(14), *message.RequestDeliveryResponse)
	assert.True(&message == message.SetRequestDeliveryResponse(456))
	assert.NotNil(message.RequestDeliveryResponse)
	assert.Equal(int64(456), *message.RequestDeliveryResponse)
}

type msgTest struct {
	desc    string // if there is a Source field, put it there, otherwise put it here
	msg     Message
	invalid bool
}

func int64Ptr(value int64) *int64 {
	return &value
}

var testMessages = []msgTest{
	// SimpleEventMessageType
	{
		msg: Message{
			Type:             SimpleEventMessageType,
			Source:           "mac:121234345656",
			Destination:      "dns:foobar.com/service",
			TransactionUUID:  "a unique identifier",
			QualityOfService: 24,
		},
	}, {
		msg: Message{
			Type:             SimpleEventMessageType,
			Source:           "invalid-source",
			Destination:      "dns:foobar.com/service",
			TransactionUUID:  "a unique identifier",
			QualityOfService: 24,
		},
		invalid: true,
	}, {
		msg: Message{
			Type:             SimpleEventMessageType,
			Source:           "dns:invalid-dest.com",
			Destination:      "invalid-dest",
			TransactionUUID:  "a unique identifier",
			QualityOfService: 24,
		},
		invalid: true,
	}, {
		msg: Message{
			Type:             SimpleEventMessageType,
			Source:           "dns:invalid-qos.com",
			Destination:      "dns:foobar.com/service",
			TransactionUUID:  "a unique identifier",
			QualityOfService: 109,
		},
		invalid: true,
	}, {
		msg: Message{
			Type:            SimpleEventMessageType,
			Source:          "dns:invalid-utf8.com",
			Destination:     "dns:foobar.com/service",
			TransactionUUID: string([]byte{0xbf}),
		},
		invalid: true,
	},

	// SimpleRequestResponseMessageType
	{
		msg: Message{
			Type:                    SimpleRequestResponseMessageType,
			Source:                  "dns:somewhere.comcast.net:9090/something",
			Destination:             "serial:1234/blergh",
			TransactionUUID:         "123-123-123",
			Status:                  int64Ptr(3471),
			RequestDeliveryResponse: int64Ptr(34),
		},
	}, {
		msg: Message{
			Type:            SimpleRequestResponseMessageType,
			Source:          "dns:external.com",
			Destination:     "mac:FFEEAADD44443333",
			TransactionUUID: "DEADBEEF",
			Headers:         []string{"Header1", "Header2"},
			Metadata:        map[string]string{"name": "value"},
			Payload:         []byte{1, 2, 3, 4, 0xff, 0xce},
			PartnerIDs:      []string{"foo"},
		},
	}, {
		msg: Message{
			Type:             SimpleRequestResponseMessageType,
			Source:           "invalid-source",
			Destination:      "dns:foobar.com/service",
			TransactionUUID:  "a unique identifier",
			QualityOfService: 24,
		},
		invalid: true,
	}, {
		msg: Message{
			Type:             SimpleRequestResponseMessageType,
			Source:           "dns:invalid-dest.com",
			Destination:      "invalid-dest",
			TransactionUUID:  "a unique identifier",
			QualityOfService: 24,
		},
		invalid: true,
	}, {
		msg: Message{
			Type:             SimpleRequestResponseMessageType,
			Source:           "dns:invalid-qos.com",
			Destination:      "dns:foobar.com/service",
			TransactionUUID:  "a unique identifier",
			QualityOfService: 109,
		},
		invalid: true,
	}, {
		msg: Message{
			Type:            SimpleRequestResponseMessageType,
			Source:          "dns:invalid-utf8.com",
			Destination:     "dns:foobar.com/service",
			TransactionUUID: string([]byte{0xbf}),
		},
		invalid: true,
	},

	// CRUD message types
	{
		msg: Message{
			Type:            CreateMessageType,
			Source:          "dns:wherever.webpa.comcast.net/glorious",
			Destination:     "uuid:1111-11-111111-11111",
			TransactionUUID: "123-123-123",
			Path:            "/some/where/over/the/rainbow",
			Payload:         []byte{1, 2, 3, 4, 0xff, 0xce},
			PartnerIDs:      []string{"foo", "bar"},
		},
	}, {
		msg: Message{
			Type:             CreateMessageType,
			Source:           "invalid-source",
			Destination:      "dns:foobar.com/service",
			TransactionUUID:  "a unique identifier",
			QualityOfService: 24,
		},
		invalid: true,
	}, {
		msg: Message{
			Type:             CreateMessageType,
			Source:           "dns:invalid-dest.com",
			Destination:      "invalid-dest",
			TransactionUUID:  "a unique identifier",
			QualityOfService: 24,
		},
		invalid: true,
	}, {
		msg: Message{
			Type:             CreateMessageType,
			Source:           "dns:invalid-qos.com",
			Destination:      "dns:foobar.com/service",
			TransactionUUID:  "a unique identifier",
			QualityOfService: 109,
		},
		invalid: true,
	}, {
		msg: Message{
			Type:            CreateMessageType,
			Source:          "dns:invalid-utf8.com",
			Destination:     "dns:foobar.com/service",
			TransactionUUID: string([]byte{0xbf}),
		},
		invalid: true,
	},

	//ServiceRegistrationMessageType
	{
		msg: Message{
			Type:        ServiceRegistrationMessageType,
			ServiceName: "service-name",
			URL:         "http://example.com",
		},
	}, {
		msg: Message{
			Type:        ServiceRegistrationMessageType,
			ServiceName: "invalid/service-name",
			URL:         "http://example.com",
		},
		invalid: true,
	}, {
		msg: Message{
			Type:        ServiceRegistrationMessageType,
			ServiceName: "invalid-utf8-string",
			URL:         string([]byte{0xbf}),
		},
		invalid: true,
	},
}

func TestMessage(t *testing.T) {
	for _, tc := range testMessages {
		desc := tc.msg.Source
		if desc == "" {
			desc = tc.desc
		}
		desc = fmt.Sprintf("%s %s", tc.msg.Type.FriendlyName(), desc)
		if tc.invalid {
			t.Run(fmt.Sprintf("Validate invalid: %s", desc), func(t *testing.T) {
				assert.Error(t, tc.msg.Validate())
			})
			continue
		}

		for _, format := range allFormats {
			t.Run(fmt.Sprintf("Validate valid: %s", desc), func(t *testing.T) {
				assert.NoError(t, tc.msg.Validate())
			})
			t.Run(fmt.Sprintf("Encode: %s %s", format, desc), func(t *testing.T) {
				var decoded Message
				var buffer bytes.Buffer
				var encoder = NewEncoder(&buffer, format)
				var decoder = NewDecoder(&buffer, format)

				tmp := tc.msg

				require.NoError(t, encoder.Encode(&tmp))
				require.NotZero(t, buffer.Len())

				require.NoError(t, decoder.Decode(&decoded))
				assert.Equal(t, tc.msg, decoded)
			})
		}
	}
}

func TestMessage_TrimmedPartnerIDs(t *testing.T) {
	tests := []struct {
		description string
		partners    []string
		want        []string
	}{
		{
			description: "empty partner list",
			partners:    []string{},
			want:        []string(nil),
		}, {
			description: "normal partner list",
			partners:    []string{"foo", "bar", "baz"},
			want:        []string{"foo", "bar", "baz"},
		}, {
			description: "partner list with empty strings",
			partners:    []string{"", "foo", "", "bar", "", "baz", ""},
			want:        []string{"foo", "bar", "baz"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tc.want, trimPartnerIDs(tc.partners))
		})
	}
}

func TestMessageTrucation(t *testing.T) {
	msg := Message{
		Type:             SimpleEventMessageType,
		Source:           "dns:foo.example.com",
		Destination:      "dns:bar.example.com",
		TransactionUUID:  "foo",
		ContentType:      "foo",
		Accept:           "foo",
		Headers:          []string{"foo", "bar"},
		Metadata:         map[string]string{"foo": "bar", "baz": "qux"},
		Path:             "foo",
		Payload:          []byte("foo"),
		ServiceName:      "foo",
		URL:              "foo",
		PartnerIDs:       []string{"foo", "bar"},
		SessionID:        "foo",
		QualityOfService: 1,
	}
	msg.SetRequestDeliveryResponse(42)
	msg.SetStatus(42)

	buf, err := msg.marshalMsg(nil)
	require.NoError(t, err)
	require.NotNil(t, buf)
	require.NotEmpty(t, buf)

	_, err = msg.unmarshalMsg(buf)
	require.NoError(t, err)

	for len(buf) > 0 {
		// truncate the buffer
		buf = buf[:len(buf)-1]
		_, err := msg.unmarshalMsg(buf)
		require.Error(t, err)
	}
}

/*
func TestValidateUTF8(t *testing.T) {
	msgType := reflect.TypeOf(Message{})

	// iterate over all of the fields in the Message struct and check if the
	// field is a string, []string, or map[string]string.  If it is, then we
	// want it set the string to a non-utf8 string and ensure that the
	// ValidateUTF8 method returns an error.
	for field := 0; field < msgType.NumField(); field++ {
		this := msgType.Field(field)
		fieldName := this.Name
		if fieldName == "Type" {
			continue
		}

		switch this.Type.Kind() {
		case reflect.String:
			t.Run(fieldName, func(t *testing.T) {
				msg := Message{}
				reflect.ValueOf(&msg).Elem().FieldByName(fieldName).SetString(string([]byte{0xbf}))
				err := validateUTF8(&msg)
				require.Error(t, err)
			})
		case reflect.Slice:
			if this.Type.Elem().Kind() == reflect.String {
				t.Run(fieldName, func(t *testing.T) {
					msg := Message{}
					reflect.ValueOf(&msg).Elem().FieldByName(fieldName).Set(reflect.ValueOf([]string{string([]byte{0xbf})}))
					err := validateUTF8(&msg)
					require.Error(t, err)
				})
			}
		case reflect.Map:
			if this.Type.Key().Kind() == reflect.String && this.Type.Elem().Kind() == reflect.String {
				t.Run(fieldName, func(t *testing.T) {
					msg := Message{}
					reflect.ValueOf(&msg).Elem().FieldByName(fieldName).Set(reflect.ValueOf(map[string]string{"invalid": string([]byte{0xbf})}))
					err := validateUTF8(&msg)
					require.Error(t, err)
				})
			}
		}
	}
}
*/

// TestMessageConsistency will test that the Message struct is consistent with
// the other structs that are used to represent the different message types.
func TestMessageConsistency(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	wrpType := reflect.TypeOf(Message{})

	for mt, strct := range mtToStruct {
		switch mt {
		case Invalid0MessageType, Invalid1MessageType, LastMessageType:
			require.Nil(strct)
			continue
		default:
			require.NotNil(strct)
		}

		strctName := reflect.TypeOf(strct).Name()
		for field := 0; field < reflect.TypeOf(strct).NumField(); field++ {
			this := reflect.TypeOf(strct).Field(field)
			// check if the field exists in wrpType.
			wrpField, found := wrpType.FieldByName(this.Name)
			assert.True(found, "Field %v.%v not found in wrp.Message", strctName, this.Name)

			// check if the type is the same in both structs, or if the field is
			// a pointer, check if the type is the same.
			wrpFieldType := wrpField.Type
			thisType := this.Type
			if wrpFieldType.Kind() == reflect.Ptr {
				wrpFieldType = wrpFieldType.Elem()

				// the field in wrp.Message is a pointer, but the field in the
				// other struct may not be a pointer, that's ok.  Example is
				// the Status field in Message struct vs Authorization struct.
				if this.Type.Kind() == reflect.Ptr {
					thisType = this.Type.Elem()
				}
			}
			assert.Equal(wrpFieldType, thisType,
				"Field '%v.%v' type mismatch", strctName, this.Name)
		}
	}
}

// TestExactCopy will test against all of the MessageType values to ensure that
// the specific struct can be copied exactly from a Message struct.  This is
// done by creating a new instance of the specific struct, populating the
// required fields with non-zero values, and then calling the From method on
// the specific struct with the Message struct.  The goal is to ensure that
// if fields are added or removed from the Message struct, that the specific
// struct will still be able to be copied exactly from the Message struct and
// unxpected fields will not be copied.
func TestExactCopy(t *testing.T) {
	keys := make([]MessageType, 0, len(mtToStruct))

	for msgType, specificStruct := range mtToStruct {
		if specificStruct == nil {
			continue
		}

		keys = append(keys, msgType)
	}

	// sort the keys
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for _, msgType := range keys {
		specificStruct := mtToStruct[msgType]
		if specificStruct == nil {
			continue
		}

		for i := -1; i < reflect.TypeOf(Message{}).NumField(); i++ {
			desc := fmt.Sprintf("MessageType %v valid", msgType)
			if i > 0 {
				name := reflect.TypeOf(Message{}).Field(i).Name
				desc = fmt.Sprintf("MessageType %v  field %s", msgType, name)
			}
			t.Run(desc, func(t *testing.T) {
				thing := reflect.New(reflect.TypeOf(specificStruct)).Interface()
				runTest(t, i, msgType, thing)
			})
		}
	}
}

func runTest(t *testing.T, index int, mt MessageType, goal any) {
	msg := &Message{
		Type: mt,
	}

	populateRequired(msg, goal)

	runTest := 1
	if index >= 0 {
		runTest = changeIndex(msg, goal, index)
		if runTest == 0 {
			return
		}
	}

	/*
		fmt.Println("Original:")
		pp.Println(goal)
		pp.Println(msg)
	*/

	if runTest == 1 {
		// run the test and expect it to pass
		err := goal.(converter).From(msg)
		assert.NoError(t, err)

		/*
			fmt.Println("After:")
			pp.Println(goal)
		*/

		// create a new instance of goal
		var back Message
		err = goal.(converter).To(&back)
		require.NoError(t, err)
		assert.Equal(t, msg, &back)

		buf, err := msg.marshalMsg(nil)
		require.NoError(t, err)
		require.NotNil(t, buf)
		require.NotEmpty(t, buf)

		left, err := msg.unmarshalMsg(buf)
		require.NoError(t, err)
		require.Empty(t, left)
		return
	}

	// run the test and expect it to fail
	next := reflect.New(reflect.TypeOf(goal).Elem()).Interface()
	err := next.(converter).From(msg)
	require.Error(t, err)
}

// populateRequired will populate the required fields in the msg with
// non-zero values.  This is done by looking up the field name in the goal
// struct and checking for the 'required' tag.  If the field is required,
// then we want to set the value in the msg to the value of required or 42.
func populateRequired(msg *Message, goal any) {
	goalType := reflect.TypeOf(goal).Elem()

	for i := 0; i < goalType.NumField(); i++ {
		field := goalType.Field(i)
		fieldName := field.Name
		if fieldName == "Type" {
			continue
		}
		if _, found := field.Tag.Lookup("required"); found {
			msgField := reflect.ValueOf(msg).Elem().FieldByName(fieldName)
			switch msgField.Kind() {
			case reflect.String:
				msgField.SetString("dns:required.example.com")
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				msgField.SetInt(42)
			case reflect.Ptr:
				if msgField.Type().Elem().Kind() == reflect.Int64 {
					ptrValue := reflect.New(msgField.Type().Elem())
					ptrValue.Elem().SetInt(42)
					msgField.Set(ptrValue)
				}
			case reflect.Slice:
				if msgField.Type().Elem().Kind() == reflect.String {
					msgField.Set(reflect.ValueOf([]string{"required"}))
				}
			case reflect.Map:
				if msgField.Type().Key().Kind() == reflect.String && msgField.Type().Elem().Kind() == reflect.String {
					msgField.Set(reflect.ValueOf(map[string]string{"key": "value"}))
				}
			default:
				if msgField.Type() == reflect.TypeOf(QOSValue(0)) {
					msgField.Set(reflect.ValueOf(QOSValue(42)))
				}
			}
		}
	}
}

// Check to see of the field index in the msg is required by looking up
// the name of the field in the goal struct and checking for the 'required'
// tag.  If the field is not required, then we want to set the value in
// the msg to a non-zero value, and if there is a field of the same name in
// the goal struct, we want to set that to the same non-zero value.  The return
// value of 0 is used to indicate that the test should not be run.  The return
// value of 1 is used to indicate that the test should be run and should pass.
// The return value of -1 is used to indicate that the test should be run and
// should fail.  If the field is required, then we want to set the value in the
// msg to the zero value of that type to break the test.
func changeIndex(msg *Message, goal any, index int) int {
	msgType := reflect.TypeOf(msg).Elem()
	goalType := reflect.TypeOf(goal).Elem()

	fieldName := msgType.Field(index).Name
	msgField := reflect.ValueOf(msg).Elem().FieldByName(fieldName)

	if tmp, found := goalType.FieldByName(fieldName); found {
		if _, required := tmp.Tag.Lookup("required"); required {
			msgField.Set(reflect.Zero(msgField.Type()))
			return -1
		}
	}

	if fieldName == "Type" {
		msgField.Set(reflect.Zero(msgField.Type()))
		return -1
	}

	switch msgField.Kind() {
	case reflect.String:
		msgField.SetString("non-zero")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		msgField.SetInt(1)
	case reflect.Ptr:
		if msgField.Type().Elem().Kind() == reflect.Int64 {
			ptrValue := reflect.New(msgField.Type().Elem())
			ptrValue.Elem().SetInt(1)
			msgField.Set(ptrValue)
		} else {
			ptrValue := reflect.New(msgField.Type().Elem())
			msgField.Set(ptrValue)
		}
	case reflect.Slice:
		switch msgField.Type().Elem().Kind() {
		case reflect.String:
			msgField.Set(reflect.ValueOf([]string{"non-zero"}))
		case reflect.Uint8:
			msgField.Set(reflect.ValueOf([]byte{1, 2, 3, 4, 0xff, 0xce}))
		default:
			panic("Unhandled slice type " + msgField.Type().Elem().Kind().String())
		}
	case reflect.Map:
		if msgField.Type().Key().Kind() == reflect.String && msgField.Type().Elem().Kind() == reflect.String {
			msgField.Set(reflect.ValueOf(map[string]string{"key": "value"}))
		}
	default:
		if msgField.Type() == reflect.TypeOf(QOSValue(0)) {
			msgField.Set(reflect.ValueOf(QOSValue(42)))
		} else {
			panic("Unhandled type " + msgField.Type().String())
		}
	}

	goalField := reflect.ValueOf(goal).Elem().FieldByName(fieldName)
	if goalField.IsValid() {
		switch goalField.Kind() {
		case reflect.String:
			goalField.SetString("non-zero")
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			goalField.SetInt(1)
		case reflect.Ptr:
			if goalField.Type().Elem().Kind() == reflect.Int64 {
				ptrValue := reflect.New(goalField.Type().Elem())
				ptrValue.Elem().SetInt(1)
				goalField.Set(ptrValue)
			} else {
				ptrValue := reflect.New(goalField.Type().Elem())
				goalField.Set(ptrValue)
			}
		case reflect.Slice:
			switch goalField.Type().Elem().Kind() {
			case reflect.String:
				goalField.Set(reflect.ValueOf([]string{"non-zero"}))
			case reflect.Uint8:
				goalField.Set(reflect.ValueOf([]byte{1, 2, 3, 4, 0xff, 0xce}))
			default:
				panic("Unhandled slice type " + msgField.Type().Elem().Kind().String())
			}
		case reflect.Map:
			if goalField.Type().Key().Kind() == reflect.String && goalField.Type().Elem().Kind() == reflect.String {
				goalField.Set(reflect.ValueOf(map[string]string{"key": "value"}))
			}
		default:
			if goalField.Type() == reflect.TypeOf(QOSValue(0)) {
				goalField.Set(reflect.ValueOf(QOSValue(42)))
			}
		}
		return 1
	}

	return -1
}

func TestMessage_ToAndValidate(t *testing.T) {
	tests := []struct {
		desc    string
		msg     converter
		invalid bool
	}{
		{
			desc: "SimpleEvent valid",
			msg: &SimpleEvent{
				Source:      "dns:foo.example.com",
				Destination: "dns:bar.example.com",
			},
		}, {
			desc: "SimpleRequestResponse valid",
			msg: &SimpleRequestResponse{
				Source:          "dns:foo.example.com",
				Destination:     "dns:bar.example.com",
				TransactionUUID: "foo",
			},
		}, {
			desc: "CRUD valid",
			msg: &CRUD{
				Type:            CreateMessageType,
				Source:          "dns:foo.example.com",
				Destination:     "dns:bar.example.com",
				TransactionUUID: "foo",
			},
		}, {
			desc: "ServiceRegistration valid",
			msg: &ServiceRegistration{
				ServiceName: "service-name",
				URL:         "http://example.com",
			},
		}, {
			desc: "Authorization valid",
			msg:  &Authorization{},
		}, {
			desc: "ServiceAlive valid",
			msg:  &ServiceAlive{},
		}, {
			desc: "Unknown valid",
			msg:  &Unknown{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			assert := assert.New(t)
			// test To
			var got Message
			err := tc.msg.To(&got)
			if tc.invalid {
				assert.Nil(got)
				assert.Error(err)
				return
			}

			assert.NotNil(got)
			assert.NoError(err)

			// test Validate
			assert.NoError(got.Validate())
		})
	}
}

func TestMessage_Setters(t *testing.T) {
	var ssr SimpleRequestResponse
	var se SimpleEvent
	var crud CRUD

	ssr.SetStatus(42)
	assert.Equal(t, int64(42), *ssr.Status)
	ssr.SetRequestDeliveryResponse(42)
	assert.Equal(t, int64(42), *ssr.RequestDeliveryResponse)

	se.SetRequestDeliveryResponse(42)
	assert.Equal(t, int64(42), *se.RequestDeliveryResponse)

	crud.SetStatus(42)
	assert.Equal(t, int64(42), *crud.Status)
	crud.SetRequestDeliveryResponse(42)
	assert.Equal(t, int64(42), *crud.RequestDeliveryResponse)

}
