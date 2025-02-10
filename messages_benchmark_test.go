// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"bytes"
	"testing"
)

func getTestMessage() Message {
	return Message{
		Type:            SimpleEventMessageType,
		Source:          "mac:112233445566i/service-name/ignored/12344",
		Destination:     "event:device-status/foo",
		TransactionUUID: "60dfdf5b-98c5-4e91-95fd-1fa6cb114cf5",
		ContentType:     "application/json",
		Headers:         []string{"key1:value1", "key2:value2"},
		Metadata: map[string]string{
			"/key/1": "value1",
			"/key/2": "value2",
		},
		Payload:    []byte(`{"key": "hello world"}`),
		SessionID:  "1234",
		PartnerIDs: []string{"partner1", "partner2"},
	}
}

func BenchmarkMarshalMsg(b *testing.B) {
	v := getTestMessage()
	buf := make([]byte, 0, 1024*32)

	var tmp bytes.Buffer
	err := NewEncoder(&tmp, Msgpack).Encode(&v)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.SetBytes(int64(tmp.Len()))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = v.marshalMsg(buf)
	}
}
func BenchmarkMarshalSingleMsg(b *testing.B) {
	v := getTestMessage()
	buf := make([]byte, 0, 1024*32)

	var tmp bytes.Buffer
	err := NewEncoder(&tmp, Msgpack).Encode(&v)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.SetBytes(int64(tmp.Len()))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = v.marshalMsg(buf)
	}
}

func BenchmarkUnmarshalMsg(b *testing.B) {
	v := getTestMessage()
	bts, _ := v.marshalMsg(nil)
	b.ReportAllocs()
	b.SetBytes(int64(len(bts)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := v.unmarshalMsg(bts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestMarshalEncodeDecode(t *testing.T) {
	v := getTestMessage()

	var buf bytes.Buffer
	err := NewEncoder(&buf, Msgpack).Encode(&v)
	if err != nil {
		t.Fatal(err)
	}

	err = NewDecoder(&buf, Msgpack).Decode(&v)
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkNewEncoderBytes(b *testing.B) {
	v := getTestMessage()
	buf := make([]byte, 0, 1024*32)

	var tmp bytes.Buffer
	err := NewEncoder(&tmp, Msgpack).Encode(&v)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.SetBytes(int64(tmp.Len()))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewEncoderBytes(&buf, Msgpack).Encode(&v)
	}
}

func BenchmarkNewDecoderBytes(b *testing.B) {
	v := getTestMessage()
	var buf bytes.Buffer
	err := NewEncoder(&buf, Msgpack).Encode(&v)
	if err != nil {
		b.Fatal(err)
	}

	bits := buf.Bytes()

	b.ReportAllocs()
	b.SetBytes(int64(buf.Len()))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = NewDecoderBytes(bits, Msgpack).Decode(&v)
		if err != nil {
			b.Fatal(err)
		}
	}
}
