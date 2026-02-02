package discovery

import "context"

// SingleModuleFinder returns only the provided path.
type SingleModuleFinder struct{}

// NewSingleModuleFinder creates a new [SingleModuleFinder].
func NewSingleModuleFinder() *SingleModuleFinder {
	return &SingleModuleFinder{}
}

// FindModules returns the rootPath as the only module.
func (f *SingleModuleFinder) FindModules(ctx context.Context, rootPath string) ([]string, error) {
	return []string{rootPath}, nil
}
