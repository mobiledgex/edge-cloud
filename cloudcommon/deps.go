// Copyright 2022 MobiledgeX, Inc
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

package cloudcommon

// Manage dependencies.
// Go modules isn't perfect. There are some packages that are dependencies
// that are not captured in go.mod. They are handled when building a module
// from within the same project, but for infra we load the platform plugin
// module into code built from a separate project (edge-cloud). This can
// cause version conflicts.
// Here we explicity import some packages to force the same dependent package
// versions.

// Import go-openapi/strfmt to avoid gopkg.in/yaml.v3 version conflict
import _ "github.com/go-openapi/errors"
import _ "github.com/go-openapi/strfmt"
import _ "github.com/go-openapi/swag"
import _ "github.com/go-openapi/validate"
