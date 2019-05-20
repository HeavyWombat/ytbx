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
	"fmt"

	yaml "gopkg.in/yaml.v2"
)

// Grab get the value from the provided YAML tree using a path to traverse through the tree structure
func Grab(obj interface{}, pathString string) (interface{}, error) {
	path, err := ParsePathString(pathString, obj)
	if err != nil {
		return nil, err
	}

	return grabByPath(obj, path)
}

func grabByPath(obj interface{}, path Path) (interface{}, error) {
	pointer, pointerPath := obj, Path{DocumentIdx: path.DocumentIdx}

	for _, element := range path.PathElements {
		switch {
		// Key/Value Map, where the element name is the key for the map
		case element.isMapElement():
			if !isMapSlice(pointer) {
				return nil, fmt.Errorf("failed to traverse tree, expected a %s but found type %s at %s", typeMap, GetType(pointer), pointerPath.ToGoPatchStyle())
			}

			entry, err := getValueByKey(pointer.(yaml.MapSlice), element.Name)
			if err != nil {
				return nil, err
			}

			pointer = entry

		// Complex List, where each list entry is a Key/Value map and the entry is identified by name using an indentifier (e.g. name, key, or id)
		case element.isComplexListElement():
			complexList, ok := castAsComplexList(pointer)
			if !ok {
				return nil, fmt.Errorf("failed to traverse tree, expected a %s but found type %s at %s", typeComplexList, GetType(pointer), pointerPath.ToGoPatchStyle())
			}

			entry, err := getEntryByIdentifierAndName(complexList, element.Key, element.Name)
			if err != nil {
				return nil, err
			}

			pointer = entry

		// Simple List (identified by index)
		case element.isSimpleListElement():
			if !isList(pointer) {
				return nil, fmt.Errorf("failed to traverse tree, expected a %s but found type %s at %s", typeSimpleList, GetType(pointer), pointerPath.ToGoPatchStyle())
			}

			list := pointer.([]interface{})
			if element.Idx < 0 || element.Idx >= len(list) {
				return nil, fmt.Errorf("failed to traverse tree, provided %s index %d is not in range: 0..%d", typeSimpleList, element.Idx, len(list)-1)
			}

			pointer = list[element.Idx]

		default:
			return nil, fmt.Errorf("failed to traverse tree, the provided path %s seems to be invalid", path)
		}

		// Update the path that the current pointer has (only used in error case to point to the right position)
		pointerPath.PathElements = append(pointerPath.PathElements, element)
	}

	return pointer, nil
}
