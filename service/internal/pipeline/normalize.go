package pipeline

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"service/internal/common"
	"service/internal/storage"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// normalize 归一化文档。
func normalize(origin string, kind string) error {
	switch kind {
	case common.FileTypeTXT:
		return normalizeTxt(origin)
	case common.FileTypeMD:
		return normalizeMd(origin)
	}
	return errors.New("不支持的文件类型。")
}

// normalizeTxt 归一化 txt 格式文档。
func normalizeTxt(origin string) error {
	return convertFile(origin, common.FileTypeTXT)
}

// normalizeMd 归一化 md 格式文档。
func normalizeMd(origin string) error {
	return convertFile(origin, common.FileTypeMD)
}

// convertFile 读取原始文件，解码清洗后保存为 Markdown 文档。
func convertFile(origin string, kind string) error {
	ctx := context.Background()
	// 读取原始文件。
	data, err := storage.OSS.Select(ctx, "Origin", origin+"."+kind)
	if err != nil {
		return fmt.Errorf("读取原始文件失败，详细错误信息：%w", err)
	}
	// 解码并清洗文本。
	text, err := convertText(data)
	if err != nil {
		return err
	}
	// 保存归一化结果。
	if err := storage.OSS.Upload(ctx, "Normal", origin+".md", text, "text/markdown; charset=utf-8"); err != nil {
		return fmt.Errorf("保存归一化文档失败，详细错误信息：%w", err)
	}
	return nil
}

// convertText 将原始数据转换为 UTF-8 编码的文本。
func convertText(data []byte) ([]byte, error) {
	// 去除 UTF-8 BOM。
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	// 非 UTF-8 文本按 GB18030 解码。
	if utf8.Valid(data) == false {
		var err error
		data, _, err = transform.Bytes(simplifiedchinese.GB18030.NewDecoder(), data)
		if err != nil {
			return nil, fmt.Errorf("解码文本失败，详细错误信息：%w", err)
		}
	}
	// 统一换行并剔除控制字符。
	var buffer bytes.Buffer
	buffer.Grow(len(data))
	for _, current := range data {
		switch {
		case current == '\r':
			continue
		case current == '\n' || current == '\t':
			buffer.WriteByte(current)
		case current < 0x20 || current == 0x7F:
			continue
		default:
			buffer.WriteByte(current)
		}
	}
	return buffer.Bytes(), nil
}
