package exif

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dsoprea/go-logging"
)

const (
	// IFD names. The paths that we referred to the IFDs with are comprised of
	// these.

	IfdStandard = "IFD"
	IfdExif     = "Exif"
	IfdGps      = "GPSInfo"
	IfdIop      = "Iop"

	// Tag IDs for child IFDs.

	IfdExifId = 0x8769
	IfdGpsId  = 0x8825
	IfdIopId  = 0xA005

	// Just a placeholder.

	IfdRootId = 0x0000

	// The paths of the standard IFDs expressed in the standard IFD-mappings
	// and as the group-names in the tag data.

	IfdPathStandard        = "IFD"
	IfdPathStandardExif    = "IFD/Exif"
	IfdPathStandardExifIop = "IFD/Exif/Iop"
	IfdPathStandardGps     = "IFD/GPSInfo"
)

var (
	ifdLogger = log.NewLogger("exif.ifd")
)

var (
	ErrChildIfdNotMapped = errors.New("no child-IFD for that tag-ID under parent")
)

// type IfdIdentity struct {
// 	ParentIfdName string
// 	IfdName       string
// }

// func (ii IfdIdentity) String() string {
// 	return fmt.Sprintf("IfdIdentity<PARENT-NAME=[%s] NAME=[%s]>", ii.ParentIfdName, ii.IfdName)
// }

type MappedIfd struct {
	ParentTagId uint16
	Placement   []uint16
	Path        []string

	Name     string
	TagId    uint16
	Children map[uint16]*MappedIfd
}

func (mi *MappedIfd) String() string {
	pathPhrase := mi.PathPhrase()
	return fmt.Sprintf("MappedIfd<(0x%04X) [%s] PATH=[%s]>", mi.TagId, mi.Name, pathPhrase)
}

func (mi *MappedIfd) PathPhrase() string {
	return strings.Join(mi.Path, "/")
}

// IfdMapping describes all of the IFDs that we currently recognize.
type IfdMapping struct {
	rootNode *MappedIfd
}

func NewIfdMapping() (ifdMapping *IfdMapping) {
	rootNode := &MappedIfd{
		Path:     make([]string, 0),
		Children: make(map[uint16]*MappedIfd),
	}

	return &IfdMapping{
		rootNode: rootNode,
	}
}

func NewIfdMappingWithStandard() (ifdMapping *IfdMapping) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	im := NewIfdMapping()

	err := LoadStandardIfds(im)
	log.PanicIf(err)

	return im
}

func (im *IfdMapping) Get(parentPlacement []uint16) (childIfd *MappedIfd, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	ptr := im.rootNode
	for _, tagId := range parentPlacement {
		if descendantPtr, found := ptr.Children[tagId]; found == false {
			log.Panicf("ifd child with tag-ID (%04x) not registered: [%s]", tagId, ptr.PathPhrase())
		} else {
			ptr = descendantPtr
		}
	}

	return ptr, nil
}

func (im *IfdMapping) GetWithPath(pathPhrase string) (mi *MappedIfd, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if pathPhrase == "" {
		log.Panicf("path-phrase is empty")
	}

	path := strings.Split(pathPhrase, "/")
	ptr := im.rootNode

	for _, name := range path {
		var hit *MappedIfd
		for _, mi := range ptr.Children {
			if mi.Name == name {
				hit = mi
				break
			}
		}

		if hit == nil {
			log.Panicf("ifd child with name [%s] not registered: [%s]", name, ptr.PathPhrase())
		}

		ptr = hit
	}

	return ptr, nil
}

// GetChild is a convenience function to get the child path for a given parent
// placement and child tag-ID.
func (im *IfdMapping) GetChild(parentPathPhrase string, tagId uint16) (mi *MappedIfd, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	mi, err = im.GetWithPath(parentPathPhrase)
	log.PanicIf(err)

	for _, childMi := range mi.Children {
		if childMi.TagId == tagId {
			return childMi, nil
		}
	}

	// Whether or not an IFD is defined in data, such an IFD is not registered
	// and would be unknown.
	log.Panic(ErrChildIfdNotMapped)
	return nil, nil
}

type IfdTagIdAndIndex struct {
	Name  string
	TagId uint16
	Index int
}

func (itii IfdTagIdAndIndex) String() string {
	return fmt.Sprintf("IfdTagIdAndIndex<NAME=[%s] ID=(%04x) INDEX=(%d)>", itii.Name, itii.TagId, itii.Index)
}

// ResolvePath takes a list of names, which can also be suffixed with indices
// (to identify the second, third, etc.. sibling IFD) and returns a list of
// tag-IDs and those indices.
//
// Example:
//
// - IFD/Exif/Iop
// - IFD0/Exif/Iop
//
// This is the only call that supports adding the numeric indices.
func (im *IfdMapping) ResolvePath(pathPhrase string) (lineage []IfdTagIdAndIndex, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	pathPhrase = strings.TrimSpace(pathPhrase)

	if pathPhrase == "" {
		log.Panicf("can not resolve empty path-phrase")
	}

	path := strings.Split(pathPhrase, "/")
	lineage = make([]IfdTagIdAndIndex, len(path))

	ptr := im.rootNode
	empty := IfdTagIdAndIndex{}
	for i, name := range path {
		indexByte := name[len(name)-1]
		index := 0
		if indexByte >= '0' && indexByte <= '9' {
			index = int(indexByte - '0')
			name = name[:len(name)-1]
		}

		itii := IfdTagIdAndIndex{}
		for _, mi := range ptr.Children {
			if mi.Name != name {
				continue
			}

			itii.Name = name
			itii.TagId = mi.TagId
			itii.Index = index

			ptr = mi

			break
		}

		if itii == empty {
			log.Panicf("ifd child with name [%s] not registered: [%s]", name, pathPhrase)
		}

		lineage[i] = itii
	}

	return lineage, nil
}

func (im *IfdMapping) FqPathPhraseFromLineage(lineage []IfdTagIdAndIndex) (fqPathPhrase string) {
	fqPathParts := make([]string, len(lineage))
	for i, itii := range lineage {
		if itii.Index > 0 {
			fqPathParts[i] = fmt.Sprintf("%s%d", itii.Name, itii.Index)
		} else {
			fqPathParts[i] = itii.Name
		}
	}

	return strings.Join(fqPathParts, "/")
}

func (im *IfdMapping) PathPhraseFromLineage(lineage []IfdTagIdAndIndex) (pathPhrase string) {
	pathParts := make([]string, len(lineage))
	for i, itii := range lineage {
		pathParts[i] = itii.Name
	}

	return strings.Join(pathParts, "/")
}

// StripPathPhraseIndices returns a non-fully-qualified path-phrase (no
// indices).
func (im *IfdMapping) StripPathPhraseIndices(pathPhrase string) (strippedPathPhrase string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	lineage, err := im.ResolvePath(pathPhrase)
	log.PanicIf(err)

	strippedPathPhrase = im.PathPhraseFromLineage(lineage)
	return strippedPathPhrase, nil
}

// Add puts the given IFD at the given position of the tree. The position of the
// tree is referred to as the placement and is represented by a set of tag-IDs,
// where the leftmost is the root tag and the tags going to the right are
// progressive descendants.
func (im *IfdMapping) Add(parentPlacement []uint16, tagId uint16, name string) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// TODO(dustin): !! It would be nicer to provide a list of names in the placement rather than tag-IDs.

	ptr, err := im.Get(parentPlacement)
	log.PanicIf(err)

	path := make([]string, len(parentPlacement)+1)
	if len(parentPlacement) > 0 {
		copy(path, ptr.Path)
	}

	path[len(path)-1] = name

	placement := make([]uint16, len(parentPlacement)+1)
	if len(placement) > 0 {
		copy(placement, ptr.Placement)
	}

	placement[len(placement)-1] = tagId

	childIfd := &MappedIfd{
		ParentTagId: ptr.TagId,
		Path:        path,
		Placement:   placement,
		Name:        name,
		TagId:       tagId,
		Children:    make(map[uint16]*MappedIfd),
	}

	if _, found := ptr.Children[tagId]; found == true {
		log.Panicf("child IFD with tag-ID (%04x) already registered under IFD [%s] with tag-ID (%04x)", tagId, ptr.Name, ptr.TagId)
	}

	ptr.Children[tagId] = childIfd

	return nil
}

func (im *IfdMapping) dumpLineages(stack []*MappedIfd, input []string) (output []string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	currentIfd := stack[len(stack)-1]

	output = input
	for _, childIfd := range currentIfd.Children {
		stackCopy := make([]*MappedIfd, len(stack)+1)

		copy(stackCopy, stack)
		stackCopy[len(stack)] = childIfd

		// Add to output, but don't include the obligatory root node.
		parts := make([]string, len(stackCopy)-1)
		for i, mi := range stackCopy[1:] {
			parts[i] = mi.Name
		}

		output = append(output, strings.Join(parts, "/"))

		output, err = im.dumpLineages(stackCopy, output)
		log.PanicIf(err)
	}

	return output, nil
}

func (im *IfdMapping) DumpLineages() (output []string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	stack := []*MappedIfd{im.rootNode}
	output = make([]string, 0)

	output, err = im.dumpLineages(stack, output)
	log.PanicIf(err)

	return output, nil
}

func LoadStandardIfds(im *IfdMapping) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	err = im.Add([]uint16{}, IfdRootId, IfdStandard)
	log.PanicIf(err)

	err = im.Add([]uint16{IfdRootId}, IfdExifId, IfdExif)
	log.PanicIf(err)

	err = im.Add([]uint16{IfdRootId, IfdExifId}, IfdIopId, IfdIop)
	log.PanicIf(err)

	err = im.Add([]uint16{IfdRootId}, IfdGpsId, IfdGps)
	log.PanicIf(err)

	return nil
}
