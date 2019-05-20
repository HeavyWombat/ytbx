// Copyright © 2018 The Homeport Team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package ytbx

import (
	"sort"

	ordered "github.com/virtuald/go-ordered-json"
	yaml "gopkg.in/yaml.v2"
)

// mapSlicify makes sure that each occurrence of a map in the provided structure
// is changed to a YAML MapSlice.
//
// Please note: In case the input data were decoded by the default standard JSON
// parser, there will be no preservation of the order of keys, because JSON does
// not support such thing as an order of keys. Therfore, the keys are sorted to
// have a consistent and testable output structure.
//
// This function supports `OrderedObjects` from the JSON library fork
// `github.com/virtuald/go-ordered-json` and will translate this structure into
// the compatible YAML structure.
func mapSlicify(obj interface{}) interface{} {
	switch tobj := obj.(type) {
	case ordered.OrderedObject:
		result := make(yaml.MapSlice, 0, len(tobj))
		for _, member := range tobj {
			result = append(result, yaml.MapItem{Key: member.Key, Value: mapSlicify(member.Value)})
		}

		return result

	case map[string]interface{}:
		return mapToYamlSlice(tobj)

	case []interface{}:
		result := make([]interface{}, len(tobj))
		for idx, entry := range tobj {
			result[idx] = mapSlicify(entry)
		}

		return result

	case []map[string]interface{}:
		result := make([]yaml.MapSlice, len(tobj))
		for idx, entry := range tobj {
			result[idx] = mapToYamlSlice(entry)
		}

		return result

	default:
		return obj
	}
}

func mapToYamlSlice(input map[string]interface{}) yaml.MapSlice {
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	result := make(yaml.MapSlice, 0, len(input))
	for _, key := range keys {
		result = append(result, yaml.MapItem{Key: key, Value: mapSlicify(input[key])})
	}

	return result
}

func castAsComplexList(obj interface{}) ([]yaml.MapSlice, bool) {
	switch tobj := obj.(type) {
	case []yaml.MapSlice:
		return tobj, true

	case []interface{}:
		if IsComplexSlice(tobj) {
			result := make([]yaml.MapSlice, len(tobj))
			for idx, entry := range tobj {
				switch x := entry.(type) {
				case yaml.MapSlice:
					result[idx] = x

				case map[string]interface{}:
					result[idx] = mapToYamlSlice(x)
				}
			}

			return result, true
		}
	}

	return nil, false
}
