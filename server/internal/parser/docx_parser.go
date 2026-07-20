package parser

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

// ParseDOCX 从 DOCX 文件中提取文本内容
// DOCX 文件本质是 ZIP 压缩包，其中 word/document.xml 包含文档正文
func ParseDOCX(filePath string) (string, error) {
	// 验证文件存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("DOCX 文件不存在: %s", filePath)
	}

	// 打开 DOCX（ZIP 格式）
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return "", fmt.Errorf("打开 DOCX 文件失败: %w", err)
	}
	defer reader.Close()

	// 查找 word/document.xml
	var docFile *zip.File
	for _, f := range reader.File {
		if f.Name == "word/document.xml" {
			docFile = f
			break
		}
	}

	if docFile == nil {
		return "", fmt.Errorf("DOCX 文件中未找到 word/document.xml")
	}

	// 读取并解析 document.xml
	rc, err := docFile.Open()
	if err != nil {
		return "", fmt.Errorf("打开 document.xml 失败: %w", err)
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return "", fmt.Errorf("读取 document.xml 失败: %w", err)
	}

	// 从 XML 中提取文本
	text, err := extractTextFromXML(content)
	if err != nil {
		return "", fmt.Errorf("解析 document.xml 失败: %w", err)
	}

	return strings.TrimSpace(text), nil
}

// extractTextFromXML 从 DOCX 的 XML 内容中提取纯文本
// 解析 w:t 标签中的文本内容，并用换行分隔段落
func extractTextFromXML(data []byte) (string, error) {
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	var texts []string
	var currentParagraph []string
	inParagraph := false
	inText := false

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			// 忽略解析错误，返回已提取的文本
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			// 检测段落开始 <w:p>
			if t.Name.Local == "p" && strings.Contains(t.Name.Space, "wordprocessingml") {
				inParagraph = true
				currentParagraph = nil
			}
			// 检测文本开始 <w:t>
			if t.Name.Local == "t" && strings.Contains(t.Name.Space, "wordprocessingml") {
				inText = true
			}

		case xml.CharData:
			if inText {
				text := strings.TrimSpace(string(t))
				if text != "" {
					currentParagraph = append(currentParagraph, text)
				}
			}

		case xml.EndElement:
			// 段落结束
			if t.Name.Local == "p" && strings.Contains(t.Name.Space, "wordprocessingml") {
				if len(currentParagraph) > 0 {
					texts = append(texts, strings.Join(currentParagraph, ""))
				}
				inParagraph = false
			}
			// 文本结束
			if t.Name.Local == "t" && strings.Contains(t.Name.Space, "wordprocessingml") {
				inText = false
			}
		}
	}

	return strings.Join(texts, "\n"), nil
}
