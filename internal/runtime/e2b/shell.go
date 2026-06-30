package e2b

import "fmt"

func uploadError(path string, index, total int, err error) error {
	return fmt.Errorf("upload %s (%d/%d): %w", path, index, total, err)
}
