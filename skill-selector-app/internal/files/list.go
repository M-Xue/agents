package files

// Entry describes a single file or directory found during a listing.
type Entry struct{}

// List returns all files and directories beneath root.
func List(root string) ([]Entry, error) {
	panic("not implemented")
}
