// Package graph 提供 Eino Career Agent 的报告并行生成功能
// 使用 goroutine + semaphore 控制并行度，不使用 Eino Graph Compile 方式
package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/hautmz/eino-carrer-agent/server/internal/agent/prompts"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/logger"
)

// ReportResult 报告生成结果，包含所有章节内容
type ReportResult struct {
	ReportID string                 `json:"report_id"`
	Sections map[string]interface{} `json:"sections"` // key 为章节名，value 为 JSON 解析后的内容
	Err      error                  `json:"-"`        // 整体错误
}

// SectionResult 单个章节的生成结果
type SectionResult struct {
	SectionName string      `json:"section_name"`
	Content     interface{} `json:"content"` // JSON 解析后的内容
	RawResponse string      `json:"-"`       // LLM 原始响应
	Err         error       `json:"-"`       // 生成错误
}

// maxConcurrentSections 最大并行章节数
var maxConcurrentSections = 4

// reportTimeout 报告总超时（秒）
var reportTimeout = 360

// sectionTimeout 单章节超时（秒）
var sectionTimeout = 120

// SetMaxConcurrentSections 设置最大并行章节数
func SetMaxConcurrentSections(n int) {
	if n > 0 {
		maxConcurrentSections = n
	}
}

// SetTimeouts 设置报告生成超时时间
func SetTimeouts(reportSec, sectionSec int) {
	if reportSec > 0 {
		reportTimeout = reportSec
	}
	if sectionSec > 0 {
		sectionTimeout = sectionSec
	}
}

// GenerateReport 执行报告生成（使用 goroutine + semaphore 并行控制）
// profileContext 为用户画像摘要文本，conversationHistory 为对话历史
func GenerateReport(ctx context.Context, chatModel model.BaseChatModel, profileContext string, conversationHistory ...*schema.Message) (*ReportResult, error) {
	// 设置报告总超时
	reportCtx, cancel := context.WithTimeout(ctx, time.Duration(reportTimeout)*time.Second)
	defer cancel()

	// 1. 提取用户画像
	profile := profileContext
	if profile == "" && len(conversationHistory) > 0 {
		var err error
		profile, err = extractProfileFromHistory(reportCtx, chatModel, conversationHistory)
		if err != nil {
			return nil, fmt.Errorf("提取用户画像失败: %w", err)
		}
	}

	// 2. 并行生成 12 个章节
	sections := prompts.ReportSections
	sectionResults := make(map[string]*SectionResult)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 使用 semaphore 控制并行数
	sem := make(chan struct{}, maxConcurrentSections)

	for i := range sections {
		section := sections[i]
		wg.Add(1)

		go func() {
			defer wg.Done()

			// 获取信号量
			sem <- struct{}{}
			defer func() { <-sem }()

			// 为每个章节设置独立的超时 context
			sectionCtx, sectionCancel := context.WithTimeout(reportCtx, time.Duration(sectionTimeout)*time.Second)
			defer sectionCancel()

			sr, err := generateSection(sectionCtx, chatModel, section, profile)

			mu.Lock()
			if err != nil {
				sectionResults[section.Name] = &SectionResult{
					SectionName: section.Name,
					Err:         err,
				}
			} else {
				sectionResults[section.Name] = sr
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	// 3. 合并结果
	result, mergeErr := mergeSections(sectionResults)
	if mergeErr != nil {
		return nil, mergeErr
	}

	return result, nil
}

// extractProfileFromHistory 从对话历史中提取用户画像
func extractProfileFromHistory(ctx context.Context, chatModel model.BaseChatModel, history []*schema.Message) (string, error) {
	messages := []*schema.Message{
		{Role: schema.System, Content: prompts.ProfileExtractionPrompt},
	}
	messages = append(messages, history...)

	resp, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("调用 LLM 提取画像失败: %w", err)
	}

	return resp.Content, nil
}

// generateSection 生成单个报告章节
func generateSection(ctx context.Context, chatModel model.BaseChatModel, section prompts.ReportSection, profile string) (*SectionResult, error) {
	logger.S().Infof("开始生成报告章节: %s (%s)", section.Title, section.Name)

	// 替换 Prompt 中的模板变量
	promptText := section.Prompt
	if profile != "" {
		promptText = replaceTemplateVar(promptText, "user_profile", profile)
	}

	// 构造消息
	messages := []*schema.Message{
		{Role: schema.System, Content: promptText},
	}

	// 调用 LLM
	resp, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("生成章节 %s 失败: %w", section.Name, err)
	}

	// 解析 JSON 响应
	var content interface{}
	rawResp := resp.Content
	if err := json.Unmarshal([]byte(rawResp), &content); err != nil {
		logger.S().Warnf("章节 %s JSON 解析失败，使用原始文本: %v", section.Name, err)
		content = map[string]string{"raw_content": rawResp}
	}

	logger.S().Infof("报告章节生成完成: %s", section.Title)

	return &SectionResult{
		SectionName: section.Name,
		Content:     content,
		RawResponse: rawResp,
	}, nil
}

// mergeSections 合并所有章节结果
func mergeSections(results map[string]*SectionResult) (*ReportResult, error) {
	reportResult := &ReportResult{
		Sections: make(map[string]interface{}),
	}

	errorSections := []string{}
	for name, sr := range results {
		if sr == nil {
			errorSections = append(errorSections, name)
			continue
		}
		if sr.Err != nil {
			errorSections = append(errorSections, name)
			reportResult.Sections[name] = map[string]string{
				"error":   sr.Err.Error(),
				"status":  "failed",
				"section": name,
			}
			continue
		}
		reportResult.Sections[name] = sr.Content
	}

	if len(errorSections) > 0 {
		logger.S().Warnf("以下章节生成失败: %v", errorSections)
	}

	return reportResult, nil
}

// replaceTemplateVar 简单的模板变量替换
// 将 {{.var_name}} 替换为实际值
func replaceTemplateVar(template string, varName string, value string) string {
	placeholder := "{{." + varName + "}}"
	result := make([]byte, 0, len(template)+len(value))
	i := 0
	for i < len(template) {
		if i+len(placeholder) <= len(template) && template[i:i+len(placeholder)] == placeholder {
			result = append(result, value...)
			i += len(placeholder)
		} else {
			result = append(result, template[i])
			i++
		}
	}
	return string(result)
}
