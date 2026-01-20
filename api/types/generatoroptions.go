// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"encoding/json"
	"fmt"
)

// GeneratorOptions modify behavior of all ConfigMap and Secret generators.
type GeneratorOptions struct {
	// Labels to add to all generated resources.
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	// Annotations to add to all generated resources.
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`

	// DisableNameSuffixHash if true disables the default behavior of adding a
	// suffix to the names of generated resources that is a hash of the
	// resource contents.
	DisableNameSuffixHash bool `json:"disableNameSuffixHash,omitempty" yaml:"disableNameSuffixHash,omitempty"`

	// Immutable if true add to all generated resources.
	Immutable bool `json:"immutable,omitempty" yaml:"immutable,omitempty"`

	// ValueMerge maps data keys to merge strategies for content-level merging.
	// Only keys listed here get content-merged during behavior:merge; all others
	// use normal key-override. Strategy must be explicitly specified (kv or yaml).
	ValueMerge map[string]ValueMergeStrategy `json:"valueMerge,omitempty" yaml:"valueMerge,omitempty"`
}

// ValueMergeStrategy specifies how a data key's value should be content-merged.
type ValueMergeStrategy string

const (
	ValueMergeStrategyKV   ValueMergeStrategy = "kv"
	ValueMergeStrategyYAML ValueMergeStrategy = "yaml"
)

func (s *ValueMergeStrategy) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	switch raw {
	case "kv":
		*s = ValueMergeStrategyKV
	case "yaml":
		*s = ValueMergeStrategyYAML
	default:
		return fmt.Errorf("invalid value merge strategy: %q, must be \"kv\" or \"yaml\"", raw)
	}
	return nil
}

// MergeGlobalOptionsIntoLocal merges two instances of GeneratorOptions.
// Values in the first 'local' argument cannot be overridden by the second
// 'global' argument, except in the case of booleans.
//
// With booleans, there's no way to distinguish an 'intentional'
// false from 'default' false.  So the rule is, if the global value
// of the value of a boolean is true, i.e. disable, it trumps the
// local value.  If the global value is false, then the local value is
// respected.  Bottom line: a local false cannot override a global true.
//
// boolean fields are always a bad idea; should always use enums instead.
func MergeGlobalOptionsIntoLocal(
	localOpts *GeneratorOptions,
	globalOpts *GeneratorOptions) *GeneratorOptions {
	if globalOpts == nil {
		return localOpts
	}
	if localOpts == nil {
		localOpts = &GeneratorOptions{}
	}
	overrideMap(&localOpts.Labels, globalOpts.Labels)
	overrideMap(&localOpts.Annotations, globalOpts.Annotations)
	overrideMap(&localOpts.ValueMerge, globalOpts.ValueMerge)
	if globalOpts.DisableNameSuffixHash {
		localOpts.DisableNameSuffixHash = true
	}
	if globalOpts.Immutable {
		localOpts.Immutable = true
	}
	return localOpts
}

func overrideMap[V any](localMap *map[string]V, globalMap map[string]V) {
	if *localMap == nil {
		if globalMap != nil {
			*localMap = CopyMap(globalMap)
		}
		return
	}
	for k, v := range globalMap {
		_, ok := (*localMap)[k]
		if !ok {
			(*localMap)[k] = v
		}
	}
}

// CopyMap copies a map.
func CopyMap[V any](in map[string]V) map[string]V {
	out := make(map[string]V)
	for k, v := range in {
		out[k] = v
	}
	return out
}
