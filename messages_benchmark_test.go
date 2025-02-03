// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp_test

import (
	"bytes"
	"testing"

	"github.com/xmidt-org/wrp-go/v4"
)

func getTestMessage() wrp.Message {
	return wrp.Message{
		Type:            wrp.SimpleEventMessageType,
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
	err := wrp.NewEncoder(&tmp, wrp.Msgpack).Encode(&v)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.SetBytes(int64(tmp.Len()))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = v.MarshalMsg(buf)
	}
}

func BenchmarkUnmarshalMsg(b *testing.B) {
	v := getTestMessage()
	bts, _ := v.MarshalMsg(nil)
	b.ReportAllocs()
	b.SetBytes(int64(len(bts)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := v.UnmarshalMsg(bts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestMarshalEncodeDecode(t *testing.T) {
	v := getTestMessage()

	var buf bytes.Buffer
	err := wrp.NewEncoder(&buf, wrp.Msgpack).Encode(&v)
	if err != nil {
		t.Fatal(err)
	}

	err = wrp.NewDecoder(&buf, wrp.Msgpack).Decode(&v)
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkNewEncoderBytes(b *testing.B) {
	v := getTestMessage()
	buf := make([]byte, 0, 1024*32)

	var tmp bytes.Buffer
	err := wrp.NewEncoder(&tmp, wrp.Msgpack).Encode(&v)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.SetBytes(int64(tmp.Len()))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrp.NewEncoderBytes(&buf, wrp.Msgpack).Encode(&v)
	}
}

func BenchmarkNewDecoderBytes(b *testing.B) {
	v := getTestMessage()
	var buf bytes.Buffer
	err := wrp.NewEncoder(&buf, wrp.Msgpack).Encode(&v)
	if err != nil {
		b.Fatal(err)
	}

	bits := buf.Bytes()

	b.ReportAllocs()
	b.SetBytes(int64(buf.Len()))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = wrp.NewDecoderBytes(bits, wrp.Msgpack).Decode(&v)
		if err != nil {
			b.Fatal(err)
		}
	}
}
