// Package parser 提供 Eino Career Agent 的文件解析功能
// 支持 PDF 和 DOCX 格式的文本提取
package parser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ParsePDF 从 PDF 文件中提取文本内容
// 使用 pdftotext 命令行工具（来自 poppler-utils）
// 如果 pdftotext 不可用，则尝试读取 PDF 中的可提取文本
func ParsePDF(filePath string) (string, error) {
	// 验证文件存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("PDF 文件不存在: %s", filePath)
	}

	// 尝试使用 pdftotext 命令行工具（如果可用）
	if path, err := exec.LookPath("pdftotext"); err == nil {
		return parsePDFWithTool(path, filePath)
	}

	// 回退方案：直接读取文件中的可提取文本
	return parsePDFDirect(filePath)
}

// parsePDFWithTool 使用 pdftotext 工具提取文本
func parsePDFWithTool(toolPath string, filePath string) (string, error) {
	// 创建临时文件保存提取的文本
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("pdf_text_%d", len(filePath)))
	defer os.Remove(tmpFile)

	cmd := exec.Command(toolPath, "-layout", filePath, tmpFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("pdftotext 执行失败: %w, output: %s", err, string(output))
	}

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		return "", fmt.Errorf("读取提取文本失败: %w", err)
	}

	return strings.TrimSpace(string(content)), nil
}

// parsePDFDirect 直接从 PDF 文件中提取可读文本
// 这是一个简化的实现，读取 PDF 中的文本流
func parsePDFDirect(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取 PDF 文件失败: %w", err)
	}

	// 简单提取：查找 PDF 中括号内的文本（BT...ET 文本块）
	// 这不是完整的 PDF 解析，但对于简单的文本 PDF 通常有效
	var texts []string
	content := string(data)
	inText := false

	// 查找 BT (Begin Text) 和 ET (End Text) 标记之间的文本
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "BT") {
			inText = true
			continue
		}
		if strings.HasPrefix(line, "ET") {
			inText = false
			continue
		}
		if inText {
			// 提取括号内的文本
			start := strings.Index(line, "(")
			end := strings.LastIndex(line, ")")
			if start >= 0 && end > start {
				text := line[start+1 : end]
				// 过滤掉明显的非文本内容
				if len(text) > 0 && isPrintableText(text) {
					texts = append(texts, text)
				}
			}
		}
	}

	if len(texts) == 0 {
		return "", fmt.Errorf("无法从 PDF 中提取文本，建议安装 pdftotext 工具以获得更好的解析效果")
	}

	return strings.Join(texts, "\n"), nil
}

// isPrintableText 检查文本是否为可打印文本（而非二进制数据）
func isPrintableText(s string) bool {
	printable := 0
	for _, r := range s {
		if r >= 32 && r < 127 || r > 127 {
			printable++
		}
	}
	if len(s) == 0 {
		return false
	}
	return float64(printable)/float64(len(s)) > 0.8
}
