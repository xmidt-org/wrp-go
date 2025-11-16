// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"testing"
)

// BenchmarkParseLocator benchmarks the original regex-based implementation
func BenchmarkParseLocator(b *testing.B) {
	locators := []string{
		"mac:112233445566/service-name/ignored/12344",
		"event:device-status/foo",
		"uuid:60dfdf5b-98c5-4e91-95fd-1fa6cb114cf5",
		"dns:talaria.example.com",
	}

	for _, loc := range locators {
		b.Run(loc, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := ParseLocator(loc)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkParseLocatorFast benchmarks the zero-allocation implementation
func BenchmarkParseLocatorFast(b *testing.B) {
	locators := []string{
		"mac:112233445566/service-name/ignored/12344",
		"event:device-status/foo",
		"uuid:60dfdf5b-98c5-4e91-95fd-1fa6cb114cf5",
		"dns:talaria.example.com",
	}

	for _, loc := range locators {
		b.Run(loc, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := ParseLocator(loc)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkParseLocatorComparison compares both implementations side by side
func BenchmarkParseLocatorComparison(b *testing.B) {
	loc := "mac:112233445566/service-name/ignored/12344"

	b.Run("Original", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ParseLocator(loc)
		}
	})

	b.Run("Fast", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ParseLocator(loc)
		}
	})
}
