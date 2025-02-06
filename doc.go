// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

/*
Package wrp defines the various WRP messages supported by WebPA and implements serialization for those messages.

Some common uses of this package include:

(1) Encoding a specific message to send to a WebPA server:

	var (
		// the infrastructure automatically fills in the correct Type field
		message = SimpleRequestResponse{
			Source: "dns:myserver.com",
			Destination: "mac:112233445566/service",
			Payload: []byte("here is a lovely little payload that the device understands"),
		}

		buffer bytes.Buffer
		encoder = NewEncoder(&buffer, Msgpack)
	)

	// To() returns a *Message, which is the generic, encodable WRP message
	out, _ := message.To()
	if err := encoder.Encode(out); err != nil {
		// deal with the error
	}

(2) Decoding any generic WRP message, perhaps sent by a client:

	// encoded may also be an io.Reader if desired
	func myHandler(encoded []byte) (message *Message, err error) {
		decoder := NewDecoderBytes(encoded, Msgpack)
		message = new(Message)
		err = decoder.Decode(message)
		return
	}

(3) Transcoding messages from one format to another:

	// assume source contains a JSON message
	func jsonToMsgpack(source io.Reader) ([]byte, error) {
		var (
			decoder = NewDecoder(source, JSON)
			buffer bytes.Buffer
			encoder = NewEncoder(&buffer, Msgpack)
		)

		// TranscodeMessage returns a *Message as its first value, which contains
		// the generic WRP message data
		if _, err := TranscodeMessage(encoder, decoder); err != nil {
			return nil, err
		}

		return buffer.Bytes(), nil
	}
*/
package wrp
