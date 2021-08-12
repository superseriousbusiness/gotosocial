package exif

import (
	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v2/common"
)

// TODO(dustin): This file now exists for backwards-compatibility only.

// NewIfdMapping returns a new IfdMapping struct.
func NewIfdMapping() (ifdMapping *exifcommon.IfdMapping) {
	return exifcommon.NewIfdMapping()
}

// NewIfdMappingWithStandard retruns a new IfdMapping struct preloaded with the
// standard IFDs.
func NewIfdMappingWithStandard() (ifdMapping *exifcommon.IfdMapping) {
	return exifcommon.NewIfdMappingWithStandard()
}

// LoadStandardIfds loads the standard IFDs into the mapping.
func LoadStandardIfds(im *exifcommon.IfdMapping) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	err = exifcommon.LoadStandardIfds(im)
	log.PanicIf(err)

	return nil
}
