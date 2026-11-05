package common

var (
	FileTypeTXT string = "txt"
	FileTypeMD  string = "md"
)

// ExistsFileType 是否存在文件类型。
func ExistsFileType(kind string) bool {
	switch kind {
	case FileTypeTXT:
		return true
	case FileTypeMD:
		return true
	default:
		return false
	}
}
