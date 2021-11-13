package nowish

//nolint
type noCopy struct{}

//nolint
func (*noCopy) Lock() {}

//nolint
func (*noCopy) Unlock() {}
