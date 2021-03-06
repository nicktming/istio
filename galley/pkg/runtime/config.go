//  Copyright 2018 Istio Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package runtime

import (
	"istio.io/istio/galley/pkg/meshconfig"
	"istio.io/istio/galley/pkg/runtime/resource"
)

// Config used by the runtime
type Config struct {
	// Cached mesh config
	// 缓存的网格配置
	Mesh meshconfig.Cache

	// Domain suffix to use
	DomainSuffix string

	// Schema for runtime
	Schema *resource.Schema

	// Enable Service Entry processing
	// 是否支持service entry
	SynthesizeServiceEntries bool
}
