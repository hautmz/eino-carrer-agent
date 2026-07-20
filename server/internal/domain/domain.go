// Package domain 定义 Eino Career Agent 的领域模型（实体）
// 这些模型对应 SQLite 数据库表结构，使用 GORM 作为 ORM
package domain

import (
	"time"
)

// User 用户实体，对应 users 表
type User struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`           // 用户 ID，自增主键
	Username     string    `gorm:"uniqueIndex;size:50;not null" json:"username"`  // 用户名，唯一索引
	PasswordHash string    `gorm:"size:255;not null" json:"-"`                    // 密码哈希（bcrypt），不输出到 JSON
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`              // 创建时间
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`              // 更新时间
}

// TableName 指定 User 对应的数据库表名
func (User) TableName() string {
	return "users"
}

// Conversation 对话实体，对应 conversations 表
type Conversation struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`         // 对话 ID，UUID 格式
	UserID    int64     `gorm:"index;not null" json:"user_id"`        // 用户 ID，外键
	Title     string    `gorm:"size:200" json:"title"`                // 对话标题
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`    // 创建时间
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`    // 更新时间

	// 关联关系
	Messages []Message `gorm:"foreignKey:ConversationID" json:"messages,omitempty"` // 对话中的消息列表
}

// TableName 指定 Conversation 对应的数据库表名
func (Conversation) TableName() string {
	return "conversations"
}

// Message 消息实体，对应 messages 表
type Message struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`            // 消息 ID
	ConversationID string    `gorm:"index;size:36;not null" json:"conversation_id"`  // 对话 ID
	Role           string    `gorm:"size:20;not null" json:"role"`                   // 角色: user/assistant/tool
	Content        string    `gorm:"type:text;not null" json:"content"`              // 消息内容
	ToolName       string    `gorm:"size:50" json:"tool_name,omitempty"`             // Tool 名称
	FileID         *int64    `gorm:"index" json:"file_id,omitempty"`                 // 关联文件 ID
	CreatedAt      time.Time `gorm:"autoCreateTime;index" json:"created_at"`         // 创建时间
}

// TableName 指定 Message 对应的数据库表名
func (Message) TableName() string {
	return "messages"
}

// Report 报告实体，对应 reports 表
// 存储职业规划报告的 12 个章节内容，每个章节为 JSON 格式
type Report struct {
	ID                      string    `gorm:"primaryKey;size:36" json:"id"`                        // 报告 ID
	ConversationID          string    `gorm:"index;size:36;not null" json:"conversation_id"`       // 来源对话 ID
	UserID                  int64     `gorm:"index;not null" json:"user_id"`                       // 用户 ID
	ProfessionalIndex       string    `gorm:"type:text" json:"professional_index,omitempty"`       // 专业指数评分 JSON
	MyselfReport            string    `gorm:"type:text" json:"myself_report,omitempty"`            // 个人信息提取 JSON
	AchievementSuperiority  string    `gorm:"type:text" json:"achievement_superiority,omitempty"`  // 成就与优势 JSON
	CareerExperience        string    `gorm:"type:text" json:"career_experience,omitempty"`        // 职业经历与成长路径 JSON
	MotivationValues        string    `gorm:"type:text" json:"motivation_values,omitempty"`        // 动机与价值观评估 JSON
	SkillHeatmap            string    `gorm:"type:text" json:"skill_heatmap,omitempty"`            // 技能热力图 JSON
	InterestAssessment      string    `gorm:"type:text" json:"interest_assessment,omitempty"`      // 职业兴趣评估 JSON
	CareerRecommendations   string    `gorm:"type:text" json:"career_recommendations,omitempty"`   // 职业推荐 JSON
	IndustryAnalysis        string    `gorm:"type:text" json:"industry_analysis,omitempty"`        // 行业分析 JSON
	GoalSetting             string    `gorm:"type:text" json:"goal_setting,omitempty"`             // 目标设定 JSON
	ActionPlan              string    `gorm:"type:text" json:"action_plan,omitempty"`              // 行动计划 JSON
	SummaryOutlook          string    `gorm:"type:text" json:"summary_outlook,omitempty"`          // 总结与展望 JSON
	Status                  string    `gorm:"size:20;default:generating" json:"status"`            // 状态: generating/completed/failed
	CreatedAt               time.Time `gorm:"autoCreateTime" json:"created_at"`                    // 创建时间
}

// TableName 指定 Report 对应的数据库表名
func (Report) TableName() string {
	return "reports"
}

// UploadedFile 上传文件实体，对应 uploaded_files 表
type UploadedFile struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`          // 文件 ID
	UserID        int64     `gorm:"index;not null" json:"user_id"`               // 用户 ID
	Filename      string    `gorm:"size:255;not null" json:"filename"`            // 原始文件名
	FilePath      string    `gorm:"size:500;not null" json:"file_path"`           // 服务器存储路径
	FileType      string    `gorm:"size:10;not null" json:"file_type"`            // 文件类型: pdf/docx
	FileSize      int64     `gorm:"not null" json:"file_size"`                    // 文件大小（字节）
	ParsedContent string    `gorm:"type:text" json:"parsed_content,omitempty"`    // 解析后的文本内容
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`            // 创建时间
}

// TableName 指定 UploadedFile 对应的数据库表名
func (UploadedFile) TableName() string {
	return "uploaded_files"
}
