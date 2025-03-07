// Copyright The OpenTelemetry Authors
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

package configmapprovider // import "go.opentelemetry.io/collector/config/configmapprovider"

import (
	"context"
	"io"

	"go.opentelemetry.io/collector/config"
)

type inMemoryMapProvider struct {
	buf io.Reader
}

// NewInMemory returns a new Provider that reads the configuration, from the provided buffer, as YAML.
func NewInMemory(buf io.Reader) Provider {
	return &inMemoryMapProvider{buf: buf}
}

func (inp *inMemoryMapProvider) Retrieve(context.Context) (Retrieved, error) {
	cfg, err := config.NewMapFromBuffer(inp.buf)
	if err != nil {
		return nil, err
	}
	return &simpleRetrieved{confMap: cfg}, nil
}

func (inp *inMemoryMapProvider) Close(context.Context) error {
	return nil
}
