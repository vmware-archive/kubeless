// Copyright 2016-2018 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nuid

import (
	"bytes"
	"testing"
)

func TestDigits(t *testing.T) {
	if len(digits) != base {
		t.Fatalf("digits length does not match base modulo")
	}
}

func TestGlobalNUIDInit(t *testing.T) {
	if globalNUID == nil {
		t.Fatalf("Expected g to be non-nil\n")
	}
	if globalNUID.pre == nil || len(globalNUID.pre) != preLen {
		t.Fatalf("Expected prefix to be initialized\n")
	}
	if globalNUID.seq == 0 {
		t.Fatalf("Expected seq to be non-zero\n")
	}
}

func TestNUIDRollover(t *testing.T) {
	globalNUID.seq = maxSeq
	// copy
	oldPre := append([]byte{}, globalNUID.pre...)
	Next()
	if bytes.Equal(globalNUID.pre, oldPre) {
		t.Fatalf("Expected new pre, got the old one\n")
	}
}

func TestGUIDLen(t *testing.T) {
	nuid := Next()
	if len(nuid) != totalLen {
		t.Fatalf("Expected len of %d, got %d\n", totalLen, len(nuid))
	}
}

func TestProperPrefix(t *testing.T) {
	min := byte(255)
	max := byte(0)
	for i := 0; i < len(digits); i++ {
		if digits[i] < min {
			min = digits[i]
		}
		if digits[i] > max {
			max = digits[i]
		}
	}
	total := 100000
	for i := 0; i < total; i++ {
		n := New()
		for j := 0; j < preLen; j++ {
			if n.pre[j] < min || n.pre[j] > max {
				t.Fatalf("Iter %d. Valid range for bytes prefix: [%d..%d]\nIncorrect prefix at pos %d: %v (%s)",
					i, min, max, j, n.pre, string(n.pre))
			}
		}
	}
}

func BenchmarkNUIDSpeed(b *testing.B) {
	n := New()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		n.Next()
	}
}

func BenchmarkGlobalNUIDSpeed(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Next()
	}
}
