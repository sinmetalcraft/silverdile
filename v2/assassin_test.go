package silverdile

import "testing"

// GetImageServiceAlterBucket is Unit Test のために AlterBucket を返す
func GetImageServiceAlterBucket(t *testing.T, is *ImageService) string {
	return is.alterBucket
}
