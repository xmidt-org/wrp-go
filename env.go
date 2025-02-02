// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func toEnvMap(msg any) map[string]string {
	v := reflect.ValueOf(msg).Elem()
	t := v.Type()
	envVars := make(map[string]string)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		tag := field.Tag.Get("env")
		if tag == "" || tag == "-" {
			continue
		}

		tagParts := strings.Split(tag, ",")
		key := tagParts[0]
		options := tagParts[1:]

		if key == "" {
			continue
		}

		if isEmptyValue(value) {
			if contains(options, "omitempty") {
				continue
			}
			envVars[key] = ""
		}

		switch value.Kind() {
		case reflect.String:
			envVars[key] = value.String()
		case reflect.Ptr:
			if value.Elem().Kind() == reflect.Int64 {
				if !value.IsNil() {
					envVars[key] = fmt.Sprintf("%d", value.Elem().Int())
				}
			}
		case reflect.Slice:
			if value.Type().Elem().Kind() == reflect.String {
				if contains(options, "multiline") {
					for i, s := range value.Interface().([]string) {
						envVars[fmt.Sprintf("%s_%03d", key, i)] = s
					}
				} else {
					envVars[key] = strings.Join(value.Interface().([]string), ",")
				}
			} else if value.Type().Elem().Kind() == reflect.Uint8 {
				envVars[key] = base64.StdEncoding.EncodeToString(value.Interface().([]byte))
			}
		case reflect.Map:
			if value.Type().Key().Kind() == reflect.String && value.Type().Elem().Kind() == reflect.String {
				for k, v := range value.Interface().(map[string]string) {
					safe := sanitizeEnvVarName(k)
					envVars[fmt.Sprintf("%s_%s", key, safe)] = k + "=" + v
				}
			}
		case reflect.Int, reflect.Int64:
			envVars[key] = fmt.Sprintf("%d", value.Int())
		}
	}

	return envVars
}

func fromEnvMap(envVars []string, msg any) error {
	v := reflect.ValueOf(msg).Elem()
	t := v.Type()
	envMap := make(map[string]string)
	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		tag := field.Tag.Get("env")
		if tag == "" || tag == "-" {
			continue
		}

		tagParts := strings.Split(tag, ",")
		key := tagParts[0]
		options := tagParts[1:]

		if key == "" {
			key = field.Name
		}

		switch value.Kind() {
		case reflect.String:
			value.SetString(envMap[key])
		case reflect.Ptr:
			if value.Type().Elem().Kind() == reflect.Int64 {
				if v, exists := envMap[key]; !exists || v == "" {
					continue
				}
				intVal, err := strconv.ParseInt(envMap[key], 10, 64)
				if err != nil {
					return err
				}
				value.Set(reflect.ValueOf(&intVal))
			}
		case reflect.Slice:
			if value.Type().Elem().Kind() == reflect.String {
				if contains(options, "multiline") {
					var slice []string
					for i := 0; ; i++ {
						multiKey := fmt.Sprintf("%s_%03d", key, i)
						if multiValue, ok := envMap[multiKey]; ok {
							slice = append(slice, multiValue)
						} else {
							break
						}
					}
					if len(slice) > 0 {
						value.Set(reflect.ValueOf(slice))
					}
				} else {
					if _, exists := envMap[key]; !exists {
						continue
					}
					list := strings.Split(envMap[key], ",")
					if len(list) > 0 && list[0] != "" {
						value.Set(reflect.ValueOf(list))
					}
				}
			} else if value.Type().Elem().Kind() == reflect.Uint8 {
				if _, exists := envMap[key]; !exists {
					continue
				}
				decoded, err := base64.StdEncoding.DecodeString(envMap[key])
				if err != nil {
					return err
				}
				if len(decoded) > 0 {
					value.Set(reflect.ValueOf(decoded))
				}
			}
		case reflect.Map:
			if value.Type().Key().Kind() == reflect.String && value.Type().Elem().Kind() == reflect.String {
				mapValue := make(map[string]string)
				for k, v := range envMap {
					if strings.HasPrefix(k, key) {
						list := strings.SplitN(v, "=", 2)
						for i := range list {
							list[i] = strings.TrimSpace(list[i])
						}
						// If there is no value, append an empty string
						list = append(list, "")
						mapValue[list[0]] = list[1]
					}
				}
				if len(mapValue) > 0 {
					value.Set(reflect.ValueOf(mapValue))
				}
			}
		case reflect.Int, reflect.Int64:
			if v, exists := envMap[key]; !exists || v == "" {
				continue
			}
			intVal, err := strconv.ParseInt(envMap[key], 10, 64)
			if err != nil {
				return err
			}
			value.SetInt(intVal)
		}
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	}
	return false
}

// Allowed environment variable characters.
var envVarNameRegexp = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func sanitizeEnvVarName(name string) string {
	name = envVarNameRegexp.ReplaceAllString(name, "_")
	// Chomp off leading underscores.
	return strings.TrimLeft(name, "_")
}
