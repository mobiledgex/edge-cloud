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
