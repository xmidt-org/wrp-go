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
		Source:          "mac:112233445566/service-name/ignored/12344",
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

func getTestMessageWithPayloadSize(size int) Message {
	msg := getTestMessage()
	msg.Payload = make([]byte, size)
	for i := 0; i < size; i++ {
		msg.Payload[i] = byte('A' + (i % 26))
	}
	return msg
}

func getSimpleRequestResponse() Message {
	status := int64(200)
	rdr := int64(1)
	return Message{
		Type:                    SimpleRequestResponseMessageType,
		Source:                  "mac:112233445566/service",
		Destination:             "mac:aabbccddeeff/service",
		TransactionUUID:         "60dfdf5b-98c5-4e91-95fd-1fa6cb114cf5",
		ContentType:             "application/json",
		Status:                  &status,
		RequestDeliveryResponse: &rdr,
		Headers:                 []string{"X-Header:value"},
		Payload:                 []byte(`{"response": "data"}`),
	}
}

func getServiceRegistration() Message {
	return Message{
		Type:        ServiceRegistrationMessageType,
		ServiceName: "test-service",
		URL:         "https://example.com/service",
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

// BenchmarkEncode tests encoding performance for different formats and message types
func BenchmarkEncode(b *testing.B) {
	testCases := []struct {
		name   string
		format Format
		msg    func() Message
	}{
		{"Msgpack/SimpleEvent", Msgpack, getTestMessage},
		{"Msgpack/SimpleRequestResponse", Msgpack, getSimpleRequestResponse},
		{"Msgpack/ServiceRegistration", Msgpack, getServiceRegistration},
		{"JSON/SimpleEvent", JSON, getTestMessage},
		{"JSON/SimpleRequestResponse", JSON, getSimpleRequestResponse},
		{"JSON/ServiceRegistration", JSON, getServiceRegistration},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			msg := tc.msg()
			buf := make([]byte, 0, 1024*32)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf = buf[:0]
				err := NewEncoderBytes(&buf, tc.format).Encode(&msg)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.SetBytes(int64(len(buf)))
		})
	}
}

// BenchmarkDecode tests decoding performance for different formats and message types
func BenchmarkDecode(b *testing.B) {
	testCases := []struct {
		name   string
		format Format
		msg    func() Message
	}{
		{"Msgpack/SimpleEvent", Msgpack, getTestMessage},
		{"Msgpack/SimpleRequestResponse", Msgpack, getSimpleRequestResponse},
		{"Msgpack/ServiceRegistration", Msgpack, getServiceRegistration},
		{"JSON/SimpleEvent", JSON, getTestMessage},
		{"JSON/SimpleRequestResponse", JSON, getSimpleRequestResponse},
		{"JSON/ServiceRegistration", JSON, getServiceRegistration},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			msg := tc.msg()
			encoded := MustEncode(&msg, tc.format)

			b.ReportAllocs()
			b.SetBytes(int64(len(encoded)))
			b.ResetTimer()

			var decoded Message
			for i := 0; i < b.N; i++ {
				err := NewDecoderBytes(encoded, tc.format).Decode(&decoded)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkEncodeWithWriter tests encoding performance using io.Writer
func BenchmarkEncodeWithWriter(b *testing.B) {
	testCases := []struct {
		name   string
		format Format
	}{
		{"Msgpack", Msgpack},
		{"JSON", JSON},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			msg := getTestMessage()
			var buf bytes.Buffer

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf.Reset()
				err := NewEncoder(&buf, tc.format).Encode(&msg)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.SetBytes(int64(buf.Len()))
		})
	}
}

// BenchmarkDecodeWithReader tests decoding performance using io.Reader
func BenchmarkDecodeWithReader(b *testing.B) {
	testCases := []struct {
		name   string
		format Format
	}{
		{"Msgpack", Msgpack},
		{"JSON", JSON},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			msg := getTestMessage()
			encoded := MustEncode(&msg, tc.format)

			b.ReportAllocs()
			b.SetBytes(int64(len(encoded)))
			b.ResetTimer()

			var decoded Message
			for i := 0; i < b.N; i++ {
				reader := bytes.NewReader(encoded)
				err := NewDecoder(reader, tc.format).Decode(&decoded)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkRoundTrip tests full encode/decode cycle
func BenchmarkRoundTrip(b *testing.B) {
	testCases := []struct {
		name   string
		format Format
	}{
		{"Msgpack", Msgpack},
		{"JSON", JSON},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			msg := getTestMessage()
			buf := make([]byte, 0, 1024*32)

			b.ReportAllocs()
			b.ResetTimer()

			var decoded Message
			for i := 0; i < b.N; i++ {
				buf = buf[:0]
				// Encode
				err := NewEncoderBytes(&buf, tc.format).Encode(&msg)
				if err != nil {
					b.Fatal(err)
				}
				// Decode
				err = NewDecoderBytes(buf, tc.format).Decode(&decoded)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkPayloadSize tests encoding/decoding performance with different payload sizes
func BenchmarkPayloadSize(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"Small_100B", 100},
		{"Medium_1KB", 1024},
		{"Large_10KB", 10 * 1024},
		{"XLarge_100KB", 100 * 1024},
	}

	for _, size := range sizes {
		b.Run("Encode/Msgpack/"+size.name, func(b *testing.B) {
			msg := getTestMessageWithPayloadSize(size.size)
			buf := make([]byte, 0, 1024*128)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf = buf[:0]
				err := NewEncoderBytes(&buf, Msgpack).Encode(&msg)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.SetBytes(int64(len(buf)))
		})

		b.Run("Encode/JSON/"+size.name, func(b *testing.B) {
			msg := getTestMessageWithPayloadSize(size.size)
			buf := make([]byte, 0, 1024*128)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf = buf[:0]
				err := NewEncoderBytes(&buf, JSON).Encode(&msg)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.SetBytes(int64(len(buf)))
		})

		b.Run("Decode/Msgpack/"+size.name, func(b *testing.B) {
			msg := getTestMessageWithPayloadSize(size.size)
			encoded := MustEncode(&msg, Msgpack)

			b.ReportAllocs()
			b.SetBytes(int64(len(encoded)))
			b.ResetTimer()

			var decoded Message
			for i := 0; i < b.N; i++ {
				err := NewDecoderBytes(encoded, Msgpack).Decode(&decoded)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("Decode/JSON/"+size.name, func(b *testing.B) {
			msg := getTestMessageWithPayloadSize(size.size)
			encoded := MustEncode(&msg, JSON)

			b.ReportAllocs()
			b.SetBytes(int64(len(encoded)))
			b.ResetTimer()

			var decoded Message
			for i := 0; i < b.N; i++ {
				err := NewDecoderBytes(encoded, JSON).Decode(&decoded)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkMustEncode tests the MustEncode helper function
func BenchmarkMustEncode(b *testing.B) {
	testCases := []struct {
		name   string
		format Format
	}{
		{"Msgpack", Msgpack},
		{"JSON", JSON},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			msg := getTestMessage()

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = MustEncode(&msg, tc.format)
			}
		})
	}
}

// BenchmarkEncodeParallel tests parallel encoding performance
func BenchmarkEncodeParallel(b *testing.B) {
	testCases := []struct {
		name   string
		format Format
	}{
		{"Msgpack", Msgpack},
		{"JSON", JSON},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				msg := getTestMessage()
				buf := make([]byte, 0, 1024*32)
				for pb.Next() {
					buf = buf[:0]
					err := NewEncoderBytes(&buf, tc.format).Encode(&msg)
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}

// BenchmarkDecodeParallel tests parallel decoding performance
func BenchmarkDecodeParallel(b *testing.B) {
	testCases := []struct {
		name   string
		format Format
	}{
		{"Msgpack", Msgpack},
		{"JSON", JSON},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			msg := getTestMessage()
			encoded := MustEncode(&msg, tc.format)

			b.ReportAllocs()
			b.SetBytes(int64(len(encoded)))
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				var decoded Message
				for pb.Next() {
					err := NewDecoderBytes(encoded, tc.format).Decode(&decoded)
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}

// BenchmarkEncodeLowAlloc tests encoding with minimal allocations by reusing encoder
func BenchmarkEncodeLowAlloc(b *testing.B) {
	testCases := []struct {
		name   string
		format Format
	}{
		{"Msgpack", Msgpack},
		{"JSON", JSON},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			msg := getTestMessage()
			buf := bytes.Buffer{}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf.Reset()
				encoder := NewEncoder(&buf, tc.format)
				err := encoder.Encode(&msg)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.SetBytes(int64(buf.Len()))
		})
	}
}

// BenchmarkEncodeMarshalMsgDirect tests direct marshalMsg call (zero-alloc for msgpack)
func BenchmarkEncodeMarshalMsgDirect(b *testing.B) {
	msg := getTestMessage()
	buf := make([]byte, 0, 1024*32)

	// Get expected size for bandwidth reporting
	tmp, _ := msg.marshalMsg(nil)

	b.ReportAllocs()
	b.SetBytes(int64(len(tmp)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf = buf[:0]
		_, _ = msg.marshalMsg(buf)
	}
}

// BenchmarkEncodeOneEncoder tests encoding by reusing a single encoder instance
func BenchmarkEncodeOneEncoder(b *testing.B) {
	testCases := []struct {
		name   string
		format Format
	}{
		{"Msgpack", Msgpack},
		{"JSON", JSON},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			msg := getTestMessage()
			var buf bytes.Buffer

			// Create encoder once, outside the loop
			encoder := NewEncoder(&buf, tc.format)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf.Reset()
				err := encoder.Encode(&msg)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.SetBytes(int64(buf.Len()))
		})
	}
}

// BenchmarkEncodeValidation compares encoding with and without validation
func BenchmarkEncodeValidation(b *testing.B) {
	msg := getTestMessage()

	b.Run("NoValidation/Direct", func(b *testing.B) {
		buf := make([]byte, 0, 1024*32)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf = buf[:0]
			_, err := msg.marshalMsg(buf)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("NoValidation/API", func(b *testing.B) {
		buf := make([]byte, 0, 1024*32)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf = buf[:0]
			err := NewEncoderBytes(&buf, Msgpack).Encode(&msg, NoStandardValidation())
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("WithValidation/API", func(b *testing.B) {
		buf := make([]byte, 0, 1024*32)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf = buf[:0]
			err := NewEncoderBytes(&buf, Msgpack).Encode(&msg)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkDecodeValidation compares decoding with and without validation
func BenchmarkDecodeValidation(b *testing.B) {
	msg := getTestMessage()
	encoded, _ := msg.marshalMsg(nil)

	b.Run("NoValidation/Direct", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(encoded)))
		b.ResetTimer()

		var decoded Message
		for i := 0; i < b.N; i++ {
			_, err := decoded.unmarshalMsg(encoded)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("NoValidation/API", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(encoded)))
		b.ResetTimer()

		var decoded Message
		for i := 0; i < b.N; i++ {
			err := NewDecoderBytes(encoded, Msgpack).Decode(&decoded, NoStandardValidation())
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("WithValidation/API", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(encoded)))
		b.ResetTimer()

		var decoded Message
		for i := 0; i < b.N; i++ {
			err := NewDecoderBytes(encoded, Msgpack).Decode(&decoded)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkValidationOverhead measures just the validation cost
func BenchmarkValidationOverhead(b *testing.B) {
	msg := getTestMessage()

	b.Run("Encode", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := validate(&msg)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Decode", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := validate(&msg)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkMessageEncodeMsgpack benchmarks the new EncodeMsgpack method
func BenchmarkMessageEncodeMsgpack(b *testing.B) {
	msg := getTestMessage()

	b.Run("WithCapacity", func(b *testing.B) {
		buf := make([]byte, 0, 1024)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf = buf[:0]
			_, err := msg.EncodeMsgpack(buf)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("WithoutCapacity", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := make([]byte, 0)
			_, err := msg.EncodeMsgpack(buf)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("NilBuffer", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := msg.EncodeMsgpack(nil)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkMessageDecodeMsgpack benchmarks the new DecodeMsgpack method
func BenchmarkMessageDecodeMsgpack(b *testing.B) {
	msg := getTestMessage()
	encoded, _ := msg.EncodeMsgpack(nil)

	b.ReportAllocs()
	b.SetBytes(int64(len(encoded)))
	b.ResetTimer()

	var decoded Message
	for i := 0; i < b.N; i++ {
		_, err := decoded.DecodeMsgpack(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMessageRoundTrip benchmarks the full encode/decode cycle using new methods
func BenchmarkMessageRoundTrip(b *testing.B) {
	msg := getTestMessage()
	buf := make([]byte, 0, 1024)

	b.ReportAllocs()
	b.ResetTimer()

	var decoded Message
	for i := 0; i < b.N; i++ {
		buf = buf[:0]
		// Encode
		encoded, err := msg.EncodeMsgpack(buf)
		if err != nil {
			b.Fatal(err)
		}
		// Decode
		_, err = decoded.DecodeMsgpack(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}
