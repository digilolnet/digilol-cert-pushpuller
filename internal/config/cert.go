// Copyright 2025 Laurynas Četyrkinas <laurynas@digilol.net>
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

package config

import "strings"

// ExtractCertName extracts the certificate name from a filename
// Returns certificate name and true if the file is a certificate or key file
// Returns empty string and false otherwise
func ExtractCertName(fileName string) (string, bool) {
	if strings.HasSuffix(fileName, ".crt") && !strings.HasSuffix(fileName, ".issuer.crt") {
		return strings.TrimSuffix(fileName, ".crt"), true
	}
	if strings.HasSuffix(fileName, ".key") {
		return strings.TrimSuffix(fileName, ".key"), true
	}
	return "", false
}
