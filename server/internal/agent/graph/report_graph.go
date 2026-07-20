// Package graph 提供 Eino Career Agent 的 Graph 编排功能
// 包括报告 12 章节并行生成 Graph 和普通对话 Graph
package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/eino/compose"
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

// ReportGraphInput 报告 Graph 的输入
type ReportGraphInput struct {
	ConversationHistory []*schema.Message // 对话历史
	UserProfile         string            // 用户画像摘要
	ReportID            string            // 报告 ID
}

// chatModelInterface 是 ChatModel 的最小接口，避免循环依赖
type chatModelInterface interface {
	Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error)
}

// maxConcurrentSections 最大并行章节数
var maxConcurrentSections = 4

// SetMaxConcurrentSections 设置最大并行章节数
func SetMaxConcurrentSections(n int) {
	if n > 0 {
		maxConcurrentSections = n
	}
}

// NewReportGraph 创建报告并行生成 Graph
// 流程：START → extract_profile → (12章节并行) → merge → END
func NewReportGraph(chatModel chatModelInterface) (compose.Runnable[ReportGraphInput, *ReportResult], error) {
	g := compose.NewGraph[ReportGraphInput, *ReportResult]()

	// 添加用户画像提取节点
	g.AddLambdaNode("extract_profile", compose.AnyLambda(func(ctx context.Context, input ReportGraphInput) (string, error) {
		// 如果已有用户画像则直接使用
		if input.UserProfile != "" {
			return input.UserProfile, nil
		}
		// 否则从对话历史中提取
		return extractProfileFromHistory(ctx, chatModel, input.ConversationHistory)
	}))

	// 添加 12 个章节并行生成节点
	sections := prompts.ReportSections
	for i := range sections {
		section := sections[i]
		g.AddLambdaNode(section.Name, compose.AnyLambda(func(ctx context.Context, profile string) (*SectionResult, error) {
			return generateSection(ctx, chatModel, section, profile)
		}))
	}

	// 添加合并节点
	g.AddLambdaNode("merge", compose.AnyLambda(func(ctx context.Context, inputs map[string]*SectionResult) (*ReportResult, error) {
		return mergeSections(inputs)
	}))

	// 连接边：START → extract_profile
	g.AddEdge(compose.START, "extract_profile")

	// 连接边：extract_profile → 各章节（并行）
	for i := range sections {
		g.AddEdge("extract_profile", sections[i].Name)
	}

	// 连接边：各章节 → merge
	for i := range sections {
		g.AddEdge(sections[i].Name, "merge")
	}

	// 连接边：merge → END
	g.AddEdge("merge", compose.END)

	// 编译 Graph
	runnable, err := g.Compile(context.Background())
	if err != nil {
		return nil, fmt.Errorf("编译 Report Graph 失败: %w", err)
	}

	return runnable, nil
}

// GenerateReport 直接执行报告生成（非 Graph 方式，更可控）
// 使用 goroutine + semaphore 控制并行度
func GenerateReport(ctx context.Context, chatModel chatModelInterface, input ReportGraphInput, sectionTimeout int) *ReportResult {
	result := &ReportResult{
		ReportID: input.ReportID,
		Sections: make(map[string]interface{}),
	}

	// 1. 提取用户画像
	profile := input.UserProfile
	if profile == "" {
		var err error
		profile, err = extractProfileFromHistory(ctx, chatModel, input.ConversationHistory)
		if err != nil {
			result.Err = fmt.Errorf("提取用户画像失败: %w", err)
			return result
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
			sectionCtx, cancel := context.WithTimeout(ctx, time.Duration(sectionTimeout)*time.Second)
			defer cancel()

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
		result = &ReportResult{
			ReportID: input.ReportID,
			Err:      mergeErr,
		}
	}
	result.ReportID = input.ReportID

	return result
}

// extractProfileFromHistory 从对话历史中提取用户画像
func extractProfileFromHistory(ctx context.Context, chatModel chatModelInterface, history []*schema.Message) (string, error) {
	// 构造消息列表
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
func generateSection(ctx context.Context, chatModel chatModelInterface, section prompts.ReportSection, profile string) (*SectionResult, error) {
	logger.S().Infof("开始生成报告章节: %s (%s)", section.Title, section.Name)

	// 替换 Prompt 中的模板变量
	promptText := section.Prompt
	if profile != "" {
		// 简单替换 {{.user_profile}} 占位符
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
		// JSON 解析失败时，将原始文本包装为对象
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
