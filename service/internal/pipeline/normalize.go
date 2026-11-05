package pipeline

import "errors"

func normalize(id string, kind string) error {
	switch kind {
	case "txt":
		return normalizeTxt(id)
	}
	return errors.New("不支持的文件类型。")
}

func normalizeTxt(id string) error {
	return nil
}
