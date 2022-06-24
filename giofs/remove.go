package giofs

import "fmt"

// Remove removes the named file or (empty) directory using `gio remove` command.
func (fsys FS) Remove(name string) error {
	_, err := gio("remove", fsys.uri+name)
	if err != nil {
		return fmt.Errorf("remove: %w", err)
	}
	return nil
}
