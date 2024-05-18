package disk

// KeyTransform defines a method of converting store keys to storage paths (and vice-versa)
type KeyTransform interface {

	// KeyToPath converts a supplied key to storage path
	KeyToPath(string) string

	// PathToKey converts a supplied storage path to key
	PathToKey(string) string
}
