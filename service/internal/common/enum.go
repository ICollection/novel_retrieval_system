package common

const (
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

const (
	// MissionStatus01 待启动。
	MissionStatus01 string = "01"
	// MissionStatus02 已启动。
	MissionStatus02 string = "02"
	// MissionStatus03 解析中。
	MissionStatus03 string = "03"
	// MissionStatus04 切分中。
	MissionStatus04 string = "04"
	// MissionStatus05 向量中。
	MissionStatus05 string = "05"
	// MissionStatus06 已完成。
	MissionStatus06 string = "06"
	// MissionStatus07 已失败。
	MissionStatus07 string = "07"
	// MissionStatus08 已取消。
	MissionStatus08 string = "08"
)

// 用户权限。
const (
	// PermissionManageUser 管理系统用户。
	PermissionManageUser string = "ManageUser"
	// PermissionManageOrigin 管理原始文件。
	PermissionManageOrigin string = "ManageOrigin"
	// PermissionManageMission 管理分析任务。
	PermissionManageMission string = "ManageMission"
)
