// Copyright 2019 The MFA Authors
// This file is part of this library.
//
// This library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with this library. If not, see <http://www.gnu.org/licenses/>.

package accounts

import (
	"testing"
)

func TestURLParsing(t *testing.T) {
	url, err := parseURL("https://www.mfa.land")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if url.Scheme != "https" {
		t.Errorf("expected: %v, got: %v", "https", url.Scheme)
	}
	if url.Path != "mfa.org" {
		t.Errorf("expected: %v, got: %v", "mfa.org", url.Path)
	}

	_, err = parseURL("mfa.org")
	if err == nil {
		t.Error("expected err, got: nil")
	}
}

func TestURLString(t *testing.T) {
	url := URL{Scheme: "https", Path: "mfa.org"}
	if url.String() != "https://www.mfa.land" {
		t.Errorf("expected: %v, got: %v", "https://www.mfa.land", url.String())
	}

	url = URL{Scheme: "", Path: "mfa.org"}
	if url.String() != "mfa.org" {
		t.Errorf("expected: %v, got: %v", "mfa.org", url.String())
	}
}

func TestURLMarshalJSON(t *testing.T) {
	url := URL{Scheme: "https", Path: "mfa.org"}
	json, err := url.MarshalJSON()
	if err != nil {
		t.Errorf("unexpcted error: %v", err)
	}
	if string(json) != "\"https://www.mfa.land\"" {
		t.Errorf("expected: %v, got: %v", "\"https://www.mfa.land\"", string(json))
	}
}

func TestURLUnmarshalJSON(t *testing.T) {
	url := &URL{}
	err := url.UnmarshalJSON([]byte("\"https://www.mfa.land\""))
	if err != nil {
		t.Errorf("unexpcted error: %v", err)
	}
	if url.Scheme != "https" {
		t.Errorf("expected: %v, got: %v", "https", url.Scheme)
	}
	if url.Path != "mfa.org" {
		t.Errorf("expected: %v, got: %v", "https", url.Path)
	}
}

func TestURLComparison(t *testing.T) {
	tests := []struct {
		urlA   URL
		urlB   URL
		expect int
	}{
		{URL{"https", "mfa.org"}, URL{"https", "mfa.org"}, 0},
		{URL{"http", "mfa.org"}, URL{"https", "mfa.org"}, -1},
		{URL{"https", "mfa.org/a"}, URL{"https", "mfa.org"}, 1},
		{URL{"https", "abc.org"}, URL{"https", "mfa.org"}, -1},
	}

	for i, tt := range tests {
		result := tt.urlA.Cmp(tt.urlB)
		if result != tt.expect {
			t.Errorf("test %d: cmp mismatch: expected: %d, got: %d", i, tt.expect, result)
		}
	}
}
