// Copyright 2024 The Podseidon Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

// Returns an arbitrary map entry and true,
// or zero KV and false if the map is empty.
func GetArbitraryMapEntry[K comparable, V any](map_ map[K]V) (K, V, bool) {
	for k, v := range map_ {
		return k, v, true
	}

	return Zero[K](), Zero[V](), false
}
