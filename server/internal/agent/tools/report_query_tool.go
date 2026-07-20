package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/logger"
	"github.com/hautmz/eino-carrer-agent/server/internal/repository"
)

// ReportQueryToolInput 报告查询 Tool 的输入参数
type ReportQueryToolInput struct {
	ReportID string `json:"report_id,omitempty" jsonschema:"description=报告ID，不提供则返回用户的报告列表"` // 报告 ID（可选）
}

// ReportQueryToolOutput 报告查询 Tool 的输出
type ReportQueryToolOutput struct {
	ReportID string      `json:"report_id,omitempty"` // 报告 ID
	Status   string      `json:"status"`              // 状态
	Summary  string      `json:"summary"`             // 报告摘要
	Reports  []ReportItem `json:"reports,omitempty"`   // 报告列表（不指定 report_id 时返回）
	Message  string      `json:"message"`             // 提示消息
}

// ReportItem 报告列表项
type ReportItem struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// NewReportQueryTool 创建报告查询 Tool
// 当用户想查看已生成的职业报告时，Agent 调用此 Tool
func NewReportQueryTool(reportRepo repository.ReportRepo, userID int64) tool.BaseTool {
	return utils.NewTool(
		&schema.ToolInfo{
			Name: "query_career_report",
			Desc: "当用户想查看已生成的职业规划报告时调用此工具。如果用户提供了报告ID，返回该报告的详细信息；如果不提供报告ID，返回用户的报告列表。",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"report_id": {
					Type: "string",
					Desc: "可选的报告ID。如果不提供，则返回用户的报告列表",
				},
			}),
		},
		func(ctx context.Context, input *ReportQueryToolInput) (*ReportQueryToolOutput, error) {
			logger.S().Infof("报告查询 Tool 被调用，报告ID: %s, 用户ID: %d", input.ReportID, userID)

			// 如果提供了报告 ID，返回单份报告详情
			if input.ReportID != "" {
				report, err := reportRepo.GetByID(ctx, input.ReportID)
				if err != nil {
					return &ReportQueryToolOutput{
						ReportID: input.ReportID,
						Status:   "not_found",
						Message:  fmt.Sprintf("未找到报告: %s", input.ReportID),
					}, nil
				}

				// 生成报告摘要（取前几个章节的概要信息）
				summary := generateReportSummary(report)

				return &ReportQueryToolOutput{
					ReportID: report.ID,
					Status:   report.Status,
					Summary:  summary,
					Message:  "报告查询成功",
				}, nil
			}

			// 未提供报告 ID，返回用户的报告列表
			reports, _, err := reportRepo.ListByUserID(ctx, userID, 0, 20)
			if err != nil {
				return nil, fmt.Errorf("查询报告列表失败: %w", err)
			}

			items := make([]ReportItem, 0, len(reports))
			for _, r := range reports {
				items = append(items, ReportItem{
					ID:        r.ID,
					Status:    r.Status,
					CreatedAt: r.CreatedAt.Format("2006-01-02 15:04:05"),
				})
			}

			message := fmt.Sprintf("共找到 %d 份报告", len(items))
			if len(items) == 0 {
				message = "暂无已生成的报告"
			}

			return &ReportQueryToolOutput{
				Status:  "success",
				Reports: items,
				Message: message,
			}, nil
		},
	)
}

// generateReportSummary 生成报告摘要文本
func generateReportSummary(report interface{}) string {
	// 简单实现：将报告转为 JSON 摘要
	bytes, err := json.Marshal(report)
	if err != nil {
		return "报告内容解析失败"
	}
	// 截取前 500 字符作为摘要
	summary := string(bytes)
	if len(summary) > 500 {
		summary = summary[:500] + "..."
	}
	return summary
}
