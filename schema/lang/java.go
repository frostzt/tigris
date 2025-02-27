// Copyright 2022-2023 Tigris Data, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//nolint:dupl
package schema

import (
	"github.com/pkg/errors"
	"github.com/tigrisdata/tigris/templates"
)

type JSONToJava struct{}

func getJavaStringType(format string) string {
	switch format {
	case formatDateTime:
		return "Date"
	case formatByte:
		return "byte[]"
	case formatUUID:
		return "UUID"
	default:
		return "String"
	}
}

func (c *JSONToJava) GetType(tp string, format string) (string, error) {
	var resType string

	switch tp {
	case typeString:
		return getJavaStringType(format), nil
	case typeInteger:
		switch format {
		case formatInt32:
			resType = "int"
		default:
			resType = "long"
		}
	case typeNumber:
		resType = "double"
	case typeBoolean:
		resType = "boolean"
	}

	if resType == "" {
		return "", errors.Wrapf(ErrUnsupportedType, "type=%s, format=%s", tp, format)
	}

	return resType, nil
}

func (*JSONToJava) GetObjectTemplate() string {
	return templates.SchemaJavaObject
}
