/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package api

import "fmt"

// ApiFile represents the apiFile.txt
type ApiFile struct {
	Name     string
	Contents string
	number   int
	group    string
	version  string
	kind     string
}

// ApiFileOptions is a way to set configurable options for the API file
type ApiFileOptions func(af *ApiFile)

// WithNumber sets the number to be used in the resulting ApiFile
func WithNumber(number int) ApiFileOptions {
	return func(af *ApiFile) {
		af.number = number
	}
}

// WithGroup sets the group value
func WithGroup(group string) ApiFileOptions {
	return func(af *ApiFile) {
		af.group = group
	}
}

// WithVersion sets the version value
func WithVersion(version string) ApiFileOptions {
	return func(af *ApiFile) {
		af.version = version
	}
}

// WithKind sets the kind value
func WithKind(kind string) ApiFileOptions {
	return func(af *ApiFile) {
		af.kind = kind
	}
}

// NewApiFile returns a new ApiFile with
func NewApiFile(opts ...ApiFileOptions) *ApiFile {
	apiFile := &ApiFile{
		Name: "apiFile.txt",
	}

	for _, opt := range opts {
		opt(apiFile)
	}

	apiFile.Contents = fmt.Sprintf(apiFileTemplate,
		apiFile.number, apiFile.group, apiFile.version, apiFile.kind)

	return apiFile
}

const apiFileTemplate = `A simple text file created with the create api subcommand
NUMBER: %d
GROUP: %s
VERSION: %s
KIND: %s`
