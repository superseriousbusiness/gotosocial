package mimetypes

import "path"

// PreferredExts defines preferred file
// extensions for input mime types (as there
// can be multiple extensions per mime type).
var PreferredExts = map[string]string{
	MimeTypes["mp3"]:  "mp3",  // audio/mpeg
	MimeTypes["mpeg"]: "mpeg", // video/mpeg
}

// GetForFilename returns mimetype for given filename.
func GetForFilename(filename string) (string, bool) {
	ext := path.Ext(filename)
	if len(ext) < 1 {
		return "", false
	}
	mime, ok := MimeTypes[ext[1:]]
	return mime, ok
}

// GetFileExt returns the file extension to use for mimetype. Relying first upon
// the 'PreferredExts' map. It simply returns the first match there may multiple.
func GetFileExt(mimeType string) (string, bool) {
	ext, ok := PreferredExts[mimeType]
	if ok {
		return ext, true
	}
	for ext, mime := range MimeTypes {
		if mime == mimeType {
			return ext, true
		}
	}
	return "", false
}

// GetFileExts returns known file extensions used for mimetype.
func GetFileExts(mimeType string) []string {
	var exts []string
	for ext, mime := range MimeTypes {
		if mime == mimeType {
			exts = append(exts, ext)
		}
	}
	return exts
}
