package storage

// KeyTransform defines a method of converting store keys to storage paths (and vice-versa)
type KeyTransform interface {
	// KeyToPath converts a supplied key to storage path
	KeyToPath(string) string

	// PathToKey converts a supplied storage path to key
	PathToKey(string) string
}

type nopKeyTransform struct{}

// NopTransform returns a nop key transform (i.e. key = path)
func NopTransform() KeyTransform {
	return &nopKeyTransform{}
}

func (t *nopKeyTransform) KeyToPath(key string) string {
	return key
}

func (t *nopKeyTransform) PathToKey(path string) string {
	return path
}
