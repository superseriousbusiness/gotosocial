package exif

import (
	"fmt"

	"github.com/dsoprea/go-logging"
	"gopkg.in/yaml.v2"
)

const (
	// IFD1

	ThumbnailOffsetTagId = 0x0201
	ThumbnailSizeTagId   = 0x0202

	// Exif

	TagVersionId = 0x0000

	TagLatitudeId     = 0x0002
	TagLatitudeRefId  = 0x0001
	TagLongitudeId    = 0x0004
	TagLongitudeRefId = 0x0003

	TagTimestampId = 0x0007
	TagDatestampId = 0x001d

	TagAltitudeId    = 0x0006
	TagAltitudeRefId = 0x0005
)

var (
	// tagsWithoutAlignment is a tag-lookup for tags whose value size won't
	// necessarily be a multiple of its tag-type.
	tagsWithoutAlignment = map[uint16]struct{}{
		// The thumbnail offset is stored as a long, but its data is a binary
		// blob (not a slice of longs).
		ThumbnailOffsetTagId: {},
	}
)

var (
	tagsLogger = log.NewLogger("exif.tags")
)

// File structures.

type encodedTag struct {
	// id is signed, here, because YAML doesn't have enough information to
	// support unsigned.
	Id       int    `yaml:"id"`
	Name     string `yaml:"name"`
	TypeName string `yaml:"type_name"`
}

// Indexing structures.

type IndexedTag struct {
	Id      uint16
	Name    string
	IfdPath string
	Type    TagTypePrimitive
}

func (it *IndexedTag) String() string {
	return fmt.Sprintf("TAG<ID=(0x%04x) NAME=[%s] IFD=[%s]>", it.Id, it.Name, it.IfdPath)
}

func (it *IndexedTag) IsName(ifdPath, name string) bool {
	return it.Name == name && it.IfdPath == ifdPath
}

func (it *IndexedTag) Is(ifdPath string, id uint16) bool {
	return it.Id == id && it.IfdPath == ifdPath
}

type TagIndex struct {
	tagsByIfd  map[string]map[uint16]*IndexedTag
	tagsByIfdR map[string]map[string]*IndexedTag
}

func NewTagIndex() *TagIndex {
	ti := new(TagIndex)

	ti.tagsByIfd = make(map[string]map[uint16]*IndexedTag)
	ti.tagsByIfdR = make(map[string]map[string]*IndexedTag)

	return ti
}

func (ti *TagIndex) Add(it *IndexedTag) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// Store by ID.

	family, found := ti.tagsByIfd[it.IfdPath]
	if found == false {
		family = make(map[uint16]*IndexedTag)
		ti.tagsByIfd[it.IfdPath] = family
	}

	if _, found := family[it.Id]; found == true {
		log.Panicf("tag-ID defined more than once for IFD [%s]: (%02x)", it.IfdPath, it.Id)
	}

	family[it.Id] = it

	// Store by name.

	familyR, found := ti.tagsByIfdR[it.IfdPath]
	if found == false {
		familyR = make(map[string]*IndexedTag)
		ti.tagsByIfdR[it.IfdPath] = familyR
	}

	if _, found := familyR[it.Name]; found == true {
		log.Panicf("tag-name defined more than once for IFD [%s]: (%s)", it.IfdPath, it.Name)
	}

	familyR[it.Name] = it

	return nil
}

// Get returns information about the non-IFD tag.
func (ti *TagIndex) Get(ifdPath string, id uint16) (it *IndexedTag, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if len(ti.tagsByIfd) == 0 {
		err := LoadStandardTags(ti)
		log.PanicIf(err)
	}

	family, found := ti.tagsByIfd[ifdPath]
	if found == false {
		log.Panic(ErrTagNotFound)
	}

	it, found = family[id]
	if found == false {
		log.Panic(ErrTagNotFound)
	}

	return it, nil
}

// Get returns information about the non-IFD tag.
func (ti *TagIndex) GetWithName(ifdPath string, name string) (it *IndexedTag, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if len(ti.tagsByIfdR) == 0 {
		err := LoadStandardTags(ti)
		log.PanicIf(err)
	}

	it, found := ti.tagsByIfdR[ifdPath][name]
	if found != true {
		log.Panic(ErrTagNotFound)
	}

	return it, nil
}

// LoadStandardTags registers the tags that all devices/applications should
// support.
func LoadStandardTags(ti *TagIndex) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// Read static data.

	encodedIfds := make(map[string][]encodedTag)

	err = yaml.Unmarshal([]byte(tagsYaml), encodedIfds)
	log.PanicIf(err)

	// Load structure.

	count := 0
	for ifdPath, tags := range encodedIfds {
		for _, tagInfo := range tags {
			tagId := uint16(tagInfo.Id)
			tagName := tagInfo.Name
			tagTypeName := tagInfo.TypeName

			// TODO(dustin): !! Non-standard types, but found in real data. Ignore for right now.
			if tagTypeName == "SSHORT" || tagTypeName == "FLOAT" || tagTypeName == "DOUBLE" {
				continue
			}

			tagTypeId, found := TypeNamesR[tagTypeName]
			if found == false {
				log.Panicf("type [%s] for [%s] not valid", tagTypeName, tagName)
				continue
			}

			it := &IndexedTag{
				IfdPath: ifdPath,
				Id:      tagId,
				Name:    tagName,
				Type:    tagTypeId,
			}

			err = ti.Add(it)
			log.PanicIf(err)

			count++
		}
	}

	tagsLogger.Debugf(nil, "(%d) tags loaded.", count)

	return nil
}
