package model

// SourceInfo is the top-level struct containing all extracted information
// about the app source code, used to generate main.go.
import (
	"unicode"

	"github.com/revel/cmd/utils"
)

type SourceInfo struct {
	// StructSpecs lists type info for all structs found under the code paths.
	// They may be queried to determine which ones (transitively) embed certain types.
	StructSpecs []*TypeInfo
	// ValidationKeys provides a two-level lookup.  The keys are:
	// 1. The fully-qualified function name,
	//    e.g. "github.com/revel/examples/chat/app/controllers.(*Application).Action"
	// 2. Within that func's file, the line number of the (overall) expression statement.
	//    e.g. the line returned from runtime.Caller()
	// The result of the lookup the name of variable being validated.
	ValidationKeys map[string]map[int]string
	// A list of import paths.
	// Revel notices files with an init() function and imports that package.
	InitImportPaths []string

	// controllerSpecs lists type info for all structs found under
	// app/controllers/... that embed (directly or indirectly) revel.Controller
	controllerSpecs []*TypeInfo
	// testSuites list the types that constitute the set of application tests.
	testSuites []*TypeInfo
}

// TypesThatEmbed returns all types that (directly or indirectly) embed the
// target type, which must be a fully qualified type name,
// e.g. "github.com/revel/revel.Controller"
func (s *SourceInfo) TypesThatEmbed(targetType, packageFilter string) (filtered []*TypeInfo) {
	// Do a search in the "embedded type graph", starting with the target type.
	var (
		nodeQueue = []string{targetType}
		processed []string
	)
	for len(nodeQueue) > 0 {
		typeSimpleName := nodeQueue[0]
		nodeQueue = nodeQueue[1:]
		processed = append(processed, typeSimpleName)

		// Look through all known structs.
		for _, spec := range s.StructSpecs {
			// If this one has been processed or is already in nodeQueue, then skip it.
			if utils.ContainsString(processed, spec.String()) ||
				utils.ContainsString(nodeQueue, spec.String()) {
				continue
			}

			// Look through the embedded types to see if the current type is among them.
			for _, embeddedType := range spec.EmbeddedTypes {

				// If so, add this type's simple name to the nodeQueue, and its spec to
				// the filtered list.
				if typeSimpleName == embeddedType.String() {
					nodeQueue = append(nodeQueue, spec.String())
					filtered = append(filtered, spec)
					break
				}
			}
		}
	}

	// Strip out any specifications that contain a lower case
	for exit := false; !exit; exit = true {
		for i, filteredItem := range filtered {
			if unicode.IsLower([]rune(filteredItem.StructName)[0]) {
				utils.Logger.Info("Debug: Skipping adding spec for unexported type",
					"type", filteredItem.StructName,
					"package", filteredItem.ImportPath)
				filtered = append(filtered[:i], filtered[i+1:]...)
				exit = false
				break
			}
		}
	}

	return
}

// ControllerSpecs returns the all the controllers that embeds
// `revel.Controller`
func (s *SourceInfo) ControllerSpecs() []*TypeInfo {
	if s.controllerSpecs == nil {
		s.controllerSpecs = s.TypesThatEmbed(RevelImportPath+".Controller", "controllers")
	}
	return s.controllerSpecs
}

// TestSuites returns the all the Application tests that embeds
// `testing.TestSuite`
func (s *SourceInfo) TestSuites() []*TypeInfo {
	if s.testSuites == nil {
		s.testSuites = s.TypesThatEmbed(RevelImportPath+"/testing.TestSuite", "testsuite")
	}
	return s.testSuites
}
