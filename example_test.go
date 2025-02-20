// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/xmidt-org/wrp-go/v5"
)

func Example_encode() {
	msg := wrp.SimpleEvent{
		Source:      "self:",
		Destination: "event:status",
		Payload:     []byte("hello"),
	}

	var buf bytes.Buffer

	if err := wrp.Msgpack.Encoder(&buf).Encode(&msg); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Encoded message:")
	fmt.Println(hex.Dump(buf.Bytes()))

	// Output: Encoded message:
	// 00000000  85 a8 6d 73 67 5f 74 79  70 65 04 a6 73 6f 75 72  |..msg_type..sour|
	// 00000010  63 65 a5 73 65 6c 66 3a  a4 64 65 73 74 ac 65 76  |ce.self:.dest.ev|
	// 00000020  65 6e 74 3a 73 74 61 74  75 73 a7 70 61 79 6c 6f  |ent:status.paylo|
	// 00000030  61 64 c4 05 68 65 6c 6c  6f a3 71 6f 73 00        |ad..hello.qos.|
}

func Example_decode() {
	bytes := []byte(`{"msg_type":4,"source":"self:","dest":"event:status","payload":"aGVsbG8="}`)

	// Decode the message into the exact expected type, or return an error. This
	// is useful when you know the exact type of message you are expecting, and
	// you want to ensure that the message is of that type.
	var exact wrp.SimpleEvent
	if err := wrp.JSON.DecoderBytes(bytes).Decode(&exact); err != nil {
		log.Fatal(err)
	}

	// Decode the message into a general type, which can be used to decode any
	// message.  This is useful when you don't know the exact type of message
	// you are expecting, or when you want to handle multiple types of messages
	// in a single function.
	var general wrp.Message
	if err := wrp.JSON.DecoderBytes(bytes).Decode(&general); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Decoded messages:")
	fmt.Printf("Source:\n   exact: %s\n   general: %s\n", exact.Source, general.Source)
	fmt.Printf("Dest:\n   exact: %s\n   general: %s\n", exact.Destination, general.Destination)
	fmt.Printf("Payload:\n   exact: %s\n   general: %s\n", string(exact.Payload), string(general.Payload))

	// Output: Decoded messages:
	// Source:
	//    exact: self:
	//    general: self:
	// Dest:
	//    exact: event:status
	//    general: event:status
	// Payload:
	//    exact: hello
	//    general: hello
}

func Example_transcode() {
	// Encode a message using the JSON format
	msg := wrp.SimpleEvent{
		Source:      "self:",
		Destination: "event:status",
		Payload:     []byte("hello"),
	}

	var srcBuf bytes.Buffer
	if err := wrp.JSON.Encoder(&srcBuf).Encode(&msg); err != nil {
		log.Fatal(err)
	}
	srcDecoder := wrp.JSON.Decoder(&srcBuf)

	var dstBuf bytes.Buffer
	dstEncoder := wrp.Msgpack.Encoder(&dstBuf)

	// Transcode the message from JSON to Msgpack
	_, err := wrp.TranscodeMessage(dstEncoder, srcDecoder)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Encoded message:")
	fmt.Println(hex.Dump(dstBuf.Bytes()))

	// Output: Encoded message:
	// 00000000  85 a8 6d 73 67 5f 74 79  70 65 04 a6 73 6f 75 72  |..msg_type..sour|
	// 00000010  63 65 a5 73 65 6c 66 3a  a4 64 65 73 74 ac 65 76  |ce.self:.dest.ev|
	// 00000020  65 6e 74 3a 73 74 61 74  75 73 a7 70 61 79 6c 6f  |ent:status.paylo|
	// 00000030  61 64 c4 05 68 65 6c 6c  6f a3 71 6f 73 00        |ad..hello.qos.|
}
