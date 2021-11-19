package gensupport

import (
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/mobiledgex/edge-cloud/protogen"
)

// Handle references between protobuffer objects (Messages).
// For example, suppose object AppInst refers to object App.
// I.e., App must exist before AppInst (an instance of that App)
// can be created. So "AppInst" is the RefByObj, and "App"
// is the RefToObj.
type RefData struct {
	RefBys   map[string]*RefByGroup // key is RefByObj type
	RefTos   map[string]*RefToGroup // key is RefToObj type
	Trackers map[string]*RefTracker // key is Tracker type
}

// RefBy is the object that is referring to another object.
// RefBy -> points to RefTo.
type RefByObj struct {
	Type        string
	TypeDesc    *generator.Descriptor
	KeyType     string
	KeyDesc     *generator.Descriptor
	GenerateCud bool
}

// RefTo is the object that another object is referring to.
// RefTo <- is pointed to by RefBy.
type RefToObj struct {
	Type        string
	TypeDesc    *generator.Descriptor
	KeyType     string
	KeyDesc     *generator.Descriptor
	GenerateCud bool
}

// Particular field in the RefByObj that is making reference.
type RefByField struct {
	Desc     *descriptor.FieldDescriptorProto
	HierName string
	InKey    bool
}

// All references to a particular refTo object
type RefToGroup struct {
	To  RefToObj
	Bys []RefByObjField
}

// All references made by a particular refBy object
type RefByGroup struct {
	By  RefByObj
	Tos []RefByFieldTo
}

// ByField and ByObj
type RefByObjField struct {
	By    RefByObj
	Field RefByField
}

// ByField and ToObj
type RefByFieldTo struct {
	Field RefByField
	To    RefToObj
}

// RefTracker is an object that tracks and records references in the database
type RefTracker struct {
	Type     string // object tracking refs
	TypeDesc *generator.Descriptor
	KeyType  string
	KeyDesc  *generator.Descriptor
	To       RefToObj        // object it's tracking refs for
	Bys      []RefByObjField // objects that have references
}

func (s *PluginSupport) GatherRefData(g *generator.Generator) *RefData {
	files := s.GetGeneratorFiles(g)

	refData := RefData{}
	refData.RefTos = make(map[string]*RefToGroup)
	refData.RefBys = make(map[string]*RefByGroup)
	refData.Trackers = make(map[string]*RefTracker)

	for _, file := range files {
		if !s.GenFile(*file.Name) {
			continue
		}
		for _, message := range file.Messages() {
			if GetIgnoreRefersTo(message.DescriptorProto) {
				continue
			}
			if GetTracksRefersTo(message.DescriptorProto) {
				refData.BuildTrackers(g, file, message)
				continue
			}
			inKey := false
			refData.BuildRefData(g, file, message, []string{}, message, []*generator.Descriptor{}, inKey)
		}
	}
	return &refData
}

func (s *RefData) BuildRefData(g *generator.Generator, file *generator.FileDescriptor, top *generator.Descriptor, parents []string, message *generator.Descriptor, visited []*generator.Descriptor, inKey bool) {
	if WasVisited(message, visited) {
		return
	}
	for _, field := range message.DescriptorProto.Field {
		name := generator.CamelCase(*field.Name)
		hierName := strings.Join(append(parents, name), ".")

		refersTo := GetRefersTo(field)
		if refersTo != "" {
			byObj := RefByObj{
				Type:        *top.Name,
				TypeDesc:    top,
				GenerateCud: GetGenerateCud(top.DescriptorProto),
			}
			if byObj.GenerateCud {
				keyDesc := GetDescKey(g, top)
				byObj.KeyType = *keyDesc.Name
				byObj.KeyDesc = keyDesc
			}
			byField := RefByField{
				Desc:     field,
				HierName: hierName,
				InKey:    inKey,
			}
			toObj := RefToObj{
				Type:     refersTo,
				TypeDesc: GetPackageDesc(g, file, refersTo),
			}
			toObj.GenerateCud = GetGenerateCud(toObj.TypeDesc.DescriptorProto)
			if toObj.GenerateCud {
				toObj.KeyDesc = GetDescKey(g, toObj.TypeDesc)
				toObj.KeyType = *toObj.KeyDesc.Name
			}
			// allow lookup by refTo
			toGroup, found := s.RefTos[toObj.Type]
			if !found {
				toGroup = &RefToGroup{}
				toGroup.To = toObj
				s.RefTos[toObj.Type] = toGroup
			}
			by := RefByObjField{
				By:    byObj,
				Field: byField,
			}
			toGroup.Bys = append(toGroup.Bys, by)

			// allow lookup by refBy
			byGroup, found := s.RefBys[byObj.Type]
			if !found {
				byGroup = &RefByGroup{}
				byGroup.By = byObj
				s.RefBys[byObj.Type] = byGroup
			}
			byFieldTo := RefByFieldTo{
				Field: byField,
				To:    toObj,
			}
			byGroup.Tos = append(byGroup.Tos, byFieldTo)
			continue
		}
		if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			subInKey := inKey
			if len(parents) == 0 && GetMessageKey(message.DescriptorProto) == field {
				subInKey = true
			}
			subDesc := GetDesc(g, field.GetTypeName())
			s.BuildRefData(g, file, top, append(parents, name), subDesc, append(visited, message), subInKey)
		}
	}
}

func (s *RefData) BuildTrackers(g *generator.Generator, file *generator.FileDescriptor, desc *generator.Descriptor) {
	tracker := RefTracker{}
	tracker.Type = *desc.Name
	tracker.TypeDesc = desc
	tracker.KeyDesc = GetDescKey(g, desc)
	tracker.KeyType = *tracker.KeyDesc.Name

	for _, field := range desc.DescriptorProto.Field {
		if toObj := GetRefersTo(field); toObj != "" {
			if tracker.To.Type != "" {
				g.Fail(`Should only be one "refers_to" field in ` + *desc.Name)
			}
			tracker.To.Type = toObj
			tracker.To.TypeDesc = GetPackageDesc(g, file, toObj)
		}
		if byObj := GetTracksRefsBy(field); byObj != "" {
			refBy := RefByObjField{}
			refBy.By.Type = byObj
			refBy.By.TypeDesc = GetPackageDesc(g, file, byObj)
			refBy.Field.Desc = field
			refBy.Field.HierName = generator.CamelCase(*field.Name)
			tracker.Bys = append(tracker.Bys, refBy)
		}
	}
	if tracker.To.Type == "" {
		g.Fail(*desc.Name + ` missing "refers_to" field`)
	}
	if len(tracker.Bys) == 0 {
		g.Fail(*desc.Name + ` missing "tracks_refs_by" fields`)
	}
	s.Trackers[tracker.Type] = &tracker
}

const DeletePrepareName = "delete_prepare"

// GetDeletePrepareField gets the hierarchical camel case name of the
// delete prepare field in the descriptor, or blank if not found.
func GetDeletePrepareField(g *generator.Generator, desc *generator.Descriptor) string {
	return findDeletePrepareField(g, desc, []string{}, []*generator.Descriptor{})
}

func findDeletePrepareField(g *generator.Generator, desc *generator.Descriptor, parents []string, visited []*generator.Descriptor) string {
	if WasVisited(desc, visited) {
		return ""
	}
	for _, field := range desc.DescriptorProto.Field {
		name := generator.CamelCase(*field.Name)
		if *field.Name == DeletePrepareName && *field.Type == descriptor.FieldDescriptorProto_TYPE_BOOL {
			return strings.Join(append(parents, name), ".")
		}
		if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			subDesc := GetDesc(g, field.GetTypeName())
			dp := findDeletePrepareField(g, subDesc, append(parents, name), append(visited, desc))
			if dp != "" {
				return dp
			}
		}
	}
	return ""
}

func GetRefersTo(field *descriptor.FieldDescriptorProto) string {
	return GetStringExtension(field.Options, protogen.E_RefersTo, "")
}

func GetTracksRefsBy(field *descriptor.FieldDescriptorProto) string {
	return GetStringExtension(field.Options, protogen.E_TracksRefsBy, "")
}

func GetGenerateCud(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateCud, false)
}

func GetIgnoreRefersTo(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_IgnoreRefersTo, false)
}

func GetTracksRefersTo(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_TracksRefersTo, false)
}
