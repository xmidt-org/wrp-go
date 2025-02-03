// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

func toHeaders(msg *Message, headers http.Header) {
	v := reflect.ValueOf(msg).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		tag := parseHttpTag(field.Tag.Get("http"))
		if tag == nil {
			continue
		}

		if isEmptyValue(value) {
			if tag.omitempty {
				continue
			}
			headers.Set(tag.primary, "")
		}

		switch value.Kind() {
		case reflect.String:
			headers.Set(tag.primary, value.String())
		case reflect.Ptr:
			if value.Elem().Kind() == reflect.Int64 {
				if !value.IsNil() {
					headers.Set(tag.primary, fmt.Sprintf("%d", value.Elem().Int()))
				}
			}
		case reflect.Slice:
			if value.Type().Elem().Kind() == reflect.String {
				if tag.multiline {
					for _, s := range value.Interface().([]string) {
						headers.Add(tag.primary, s)
					}
				} else {
					headers.Set(tag.primary, strings.Join(value.Interface().([]string), ","))
				}
			}
		case reflect.Map:
			if value.Type().Key().Kind() == reflect.String && value.Type().Elem().Kind() == reflect.String {
				for k, v := range value.Interface().(map[string]string) {
					headers.Add(tag.primary, k+"="+v)
				}
			}
		case reflect.Int, reflect.Int64:
			headers.Add(tag.primary, fmt.Sprintf("%d", value.Int()))
		}
	}
}

func fromHeaders(headers http.Header, msg *Message) error {
	v := reflect.ValueOf(msg).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		tag := parseHttpTag(field.Tag.Get("http"))
		if tag == nil {
			continue
		}

		switch value.Kind() {
		case reflect.TypeOf(MessageType(0)).Kind():
			v := tag.get(headers)
			if v == "" {
				continue
			}
			msgType := StringToMessageType(v)
			if msgType == LastMessageType {
				return fmt.Errorf("unknown message type: %s", v)
			}
			value.Set(reflect.ValueOf(msgType))
		case reflect.String:
			value.SetString(tag.get(headers))
		case reflect.Ptr:
			if value.Type().Elem().Kind() == reflect.Int64 {
				v := tag.get(headers)
				if v == "" {
					continue
				}
				intVal, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return err
				}
				value.Set(reflect.ValueOf(&intVal))
			}
		case reflect.Slice:
			if value.Type().Elem().Kind() == reflect.String {
				lines := tag.all(headers)
				final := make([]string, 0, len(lines))

				for _, line := range lines {
					list := strings.Split(line, ",")
					for _, item := range list {
						item = strings.TrimSpace(item)
						if item != "" {
							final = append(final, item)
						}
					}
				}
				if len(final) > 0 {
					value.Set(reflect.ValueOf(final))
				}
			}
		case reflect.Map:
			if value.Type().Key().Kind() == reflect.String && value.Type().Elem().Kind() == reflect.String {
				mapValue := make(map[string]string)
				slice := tag.all(headers)
				for _, item := range slice {
					parts := strings.SplitN(item, "=", 2)
					for i := range parts {
						parts[i] = strings.TrimSpace(parts[i])
					}

					// If there is no value, append an empty string so that the
					// key is still present
					parts = append(parts, "")

					if parts[0] != "" {
						mapValue[parts[0]] = parts[1]
					}
				}
				if len(mapValue) > 0 {
					value.Set(reflect.ValueOf(mapValue))
				}
			}
		case reflect.Int, reflect.Int64:
			v := tag.get(headers)
			if v == "" {
				continue
			}
			intVal, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return err
			}
			value.SetInt(intVal)
		}
	}

	return nil
}

func defaultAHeader(headers *http.Header, key string, value []string) {
	list := headers.Values(key)
	for _, v := range list {
		v = strings.TrimSpace(v)
		if v != "" {
			return
		}
	}

	if value == nil {
		return
	}

	headers.Set(key, value[0])
	for i := 1; i < len(value); i++ {
		headers.Add(key, value[i])
	}
}

type httpTag struct {
	primary   string
	accepted  []string
	multiline bool
	omitempty bool
}

func parseHttpTag(tag string) *httpTag {
	if tag == "" || tag == "-" {
		return nil
	}

	tagParts := strings.Split(tag, ",")

	rv := httpTag{
		primary:  http.CanonicalHeaderKey(strings.TrimSpace(tagParts[0])),
		accepted: make([]string, 0, len(tagParts)),
	}

	if rv.primary == "" {
		return nil
	}

	for _, val := range tagParts[1:] {
		val = strings.TrimSpace(val)

		switch val {
		case "omitempty":
			rv.omitempty = true
		case "multiline":
			rv.multiline = true
		default:
			rv.accepted = append(rv.accepted, val)
		}
	}

	return &rv
}

func (tag httpTag) get(headers http.Header) string {
	if v := headers.Get(tag.primary); v != "" {
		return v
	}
	for _, accepted := range tag.accepted {
		if v := headers.Get(accepted); v != "" {
			return v
		}
	}

	return ""
}

func (tag httpTag) all(headers http.Header) []string {
	if v := headers[tag.primary]; len(v) > 0 {
		return v
	}
	for _, accepted := range tag.accepted {
		if v := headers[accepted]; len(v) > 0 {
			return v
		}
	}

	return nil
}
