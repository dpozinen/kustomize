// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	. "sigs.k8s.io/kustomize/api/types"
)

func TestMergeGlobalOptionsIntoLocal(t *testing.T) {
	tests := []struct {
		name     string
		local    *GeneratorOptions
		global   *GeneratorOptions
		expected *GeneratorOptions
	}{
		{
			name:     "everything nil",
			local:    nil,
			global:   nil,
			expected: nil,
		},
		{
			name: "nil global",
			local: &GeneratorOptions{
				Labels:      map[string]string{"pet": "dog"},
				Annotations: map[string]string{"fruit": "apple"},
				ValueMerge: map[string]ValueMergeStrategy{
					"app.properties": ValueMergeStrategyKV,
				},
			},
			global: nil,
			expected: &GeneratorOptions{
				Labels:      map[string]string{"pet": "dog"},
				Annotations: map[string]string{"fruit": "apple"},
				ValueMerge: map[string]ValueMergeStrategy{
					"app.properties": ValueMergeStrategyKV,
				},
			},
		},
		{
			name:  "nil local",
			local: nil,
			global: &GeneratorOptions{
				Labels:      map[string]string{"pet": "dog"},
				Annotations: map[string]string{"fruit": "apple"},
				ValueMerge: map[string]ValueMergeStrategy{
					"config.yaml": ValueMergeStrategyYAML,
				},
			},
			expected: &GeneratorOptions{
				Labels:      map[string]string{"pet": "dog"},
				Annotations: map[string]string{"fruit": "apple"},
				ValueMerge: map[string]ValueMergeStrategy{
					"config.yaml": ValueMergeStrategyYAML,
				},
			},
		},
		{
			name: "global doesn't damage local",
			local: &GeneratorOptions{
				Labels:      map[string]string{"pet": "dog"},
				Annotations: map[string]string{"fruit": "apple"},
				ValueMerge: map[string]ValueMergeStrategy{
					"app.properties": ValueMergeStrategyKV,
				},
			},
			global: &GeneratorOptions{
				Labels:      map[string]string{"pet": "cat", "simpson": "homer"},
				Annotations: map[string]string{"fruit": "peach", "tesla": "Y"},
				ValueMerge: map[string]ValueMergeStrategy{
					"app.properties": ValueMergeStrategyYAML,
					"config.yaml":    ValueMergeStrategyYAML,
				},
			},
			expected: &GeneratorOptions{
				Labels:      map[string]string{"pet": "dog", "simpson": "homer"},
				Annotations: map[string]string{"fruit": "apple", "tesla": "Y"},
				ValueMerge: map[string]ValueMergeStrategy{
					"app.properties": ValueMergeStrategyKV,
					"config.yaml":    ValueMergeStrategyYAML,
				},
			},
		},
		{
			name: "global disable trumps local",
			local: &GeneratorOptions{
				DisableNameSuffixHash: false,
				Immutable:             false,
			},
			global: &GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
			expected: &GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
		},
		{
			name: "local disable works",
			local: &GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
			global: &GeneratorOptions{
				DisableNameSuffixHash: false,
				Immutable:             false,
			},
			expected: &GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
		},
		{
			name: "everyone wants disable",
			local: &GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
			global: &GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
			expected: &GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
		},
		{
			name:  "nil local valueMerge takes global",
			local: &GeneratorOptions{},
			global: &GeneratorOptions{
				ValueMerge: map[string]ValueMergeStrategy{
					"config.yaml": ValueMergeStrategyYAML,
				},
			},
			expected: &GeneratorOptions{
				ValueMerge: map[string]ValueMergeStrategy{
					"config.yaml": ValueMergeStrategyYAML,
				},
			},
		},
	}
	for _, tc := range tests {
		actual := MergeGlobalOptionsIntoLocal(tc.local, tc.global)
		if !reflect.DeepEqual(tc.expected, actual) {
			t.Fatalf("%s: Expected '%v', got '%v'",
				tc.name, tc.expected, *actual)
		}
	}
}

func TestGeneratorOptions_ValueMerge_JSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *GeneratorOptions
		expectError bool
	}{
		{
			name:  "with valueMerge kv and yaml strategies",
			input: `{"valueMerge":{"app.properties":"kv","config.yaml":"yaml"}}`,
			expected: &GeneratorOptions{
				ValueMerge: map[string]ValueMergeStrategy{
					"app.properties": ValueMergeStrategyKV,
					"config.yaml":    ValueMergeStrategyYAML,
				},
			},
		},
		{
			name:     "without valueMerge",
			input:    `{}`,
			expected: &GeneratorOptions{},
		},
		{
			name:        "with invalid strategy",
			input:       `{"valueMerge":{"app.properties":"invalid"}}`,
			expectError: true,
		},
		{
			name:        "with empty strategy value",
			input:       `{"valueMerge":{"app.properties":""}}`,
			expectError: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var opts GeneratorOptions
			err := json.Unmarshal([]byte(tc.input), &opts)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, &opts)
			}
		})
	}
}
