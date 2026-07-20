package prompts

// ReportSection 定义报告章节的结构
type ReportSection struct {
	Name        string // 章节英文标识（对应数据库字段名）
	Title       string // 章节中文标题
	Description string // 章节简短描述
	Prompt      string // 章节生成 Prompt
}

// ReportSections 是报告的 12 个章节定义
// 顺序决定了报告展示顺序，但不影响并行生成逻辑
var ReportSections = []ReportSection{
	{
		Name:        "professional_index",
		Title:       "专业指数评分",
		Description: "8个维度的专业能力评分",
		Prompt: `你是一位专业的职业评估专家。请根据以下用户画像，对用户的8个维度进行专业指数评分（1-5分）。

用户画像：
{{.user_profile}}

请严格按以下 JSON 格式输出，不要输出任何其他内容：
{
  "self_awareness": {"score": 4, "evidence": "依据描述"},
  "achievement_events": {"score": 3, "evidence": "依据描述"},
  "skill_depth": {"score": 4, "evidence": "依据描述"},
  "skill_breadth": {"score": 3, "evidence": "依据描述"},
  "career_adaptability": {"score": 4, "evidence": "依据描述"},
  "learning_agility": {"score": 5, "evidence": "依据描述"},
  "interpersonal_influence": {"score": 3, "evidence": "依据描述"},
  "industry_insight": {"score": 2, "evidence": "依据描述"},
  "overall_score": 3.5,
  "overall_level": "中等偏上",
  "improvement_suggestions": ["建议1", "建议2"]
}`,
	},
	{
		Name:        "myself_report",
		Title:       "个人信息提取",
		Description: "从对话中提取的用户基本画像",
		Prompt: `你是一位专业的职业信息分析师。请根据以下用户画像，生成完整的个人职业画像报告。

用户画像：
{{.user_profile}}

请严格按以下 JSON 格式输出，不要输出任何其他内容：
{
  "basic_info": {
    "age_range": "年龄段",
    "education": "学历",
    "major": "专业",
    "work_years": "工作年限"
  },
  "career_stage": "当前职业阶段（如成长期/成熟期/转型期）",
  "core_competencies": ["核心能力1", "核心能力2"],
  "personality_analysis": "性格特质分析，200字以内",
  "interest_analysis": "兴趣倾向分析，200字以内",
  "value_orientation": "价值观倾向，200字以内",
  "strengths": ["优势1", "优势2"],
  "weaknesses": ["待提升1", "待提升2"],
  "summary": "200字以内的个人画像综合总结"
}`,
	},
	{
		Name:        "achievement_superiority",
		Title:       "成就与优势",
		Description: "关键成就事件与核心优势分析",
		Prompt: `你是一位专业的职业成就分析师。请根据以下用户画像，分析用户的成就事件和核心优势。

用户画像：
{{.user_profile}}

请严格按以下 JSON 格式输出，不要输出任何其他内容：
{
  "key_achievements": [
    {"title": "成就标题", "description": "成就描述", "impact": "影响和价值", "skills_used": ["所用技能"]}
  ],
  "core_strengths": [
    {"strength": "优势名称", "evidence": "支撑证据", "how_to_leverage": "如何发挥"}
  ],
  "achievement_pattern": "成就模式分析，200字以内",
  "unique_advantages": "独特竞争优势，200字以内",
  "potential_risks": ["潜在风险1", "潜在风险2"],
  "strength_development_plan": "优势发展建议，150字以内"
}`,
	},
	{
		Name:        "career_experience",
		Title:       "职业经历与成长路径",
		Description: "职业发展时间线与成长轨迹",
		Prompt: `你是一位专业的职业发展顾问。请根据以下用户画像，分析用户的职业经历和成长路径。

用户画像：
{{.user_profile}}

请严格按以下 JSON 格式输出，不要输出任何其他内容：
{
  "career_timeline": [
    {"period": "时间段", "role": "角色/职位", "organization_type": "组织类型", "key_learning": "关键收获"}
  ],
  "growth_pattern": "成长模式分析，200字以内",
  "career_transitions": [
    {"from": "起始状态", "to": "目标状态", "driver": "转型驱动力"}
  ],
  "skill_evolution": "技能演进轨迹，200字以内",
  "turning_points": ["关键转折点1", "关键转折点2"],
  "current_stage_analysis": "当前阶段分析，200字以内",
  "growth_trajectory": "未来成长轨迹预测，150字以内"
}`,
	},
	{
		Name:        "motivation_values",
		Title:       "动机与价值观评估",
		Description: "6维度动机分析+15项价值观评估",
		Prompt: `你是一位专业的职业心理学评估师。请根据以下用户画像，评估用户的职业动机和价值观。

用户画像：
{{.user_profile}}

请严格按以下 JSON 格式输出，不要输出任何其他内容：
{
  "motivation_dimensions": {
    "achievement": {"score": 4, "description": "成就动机描述"},
    "power": {"score": 3, "description": "权力动机描述"},
    "affiliation": {"score": 4, "description": "亲和动机描述"},
    "security": {"score": 3, "description": "安全动机描述"},
    "autonomy": {"score": 5, "description": "自主动机描述"},
    "growth": {"score": 4, "description": "成长动机描述"}
  },
  "values_assessment": [
    {"value": "价值观名称", "importance": "高/中/低", "manifestation": "在职业中的体现"}
  ],
  "core_values": ["核心价值观1", "核心价值观2", "核心价值观3"],
  "motivation_profile": "动机画像总结，200字以内",
  "values_career_alignment": "价值观与职业匹配度分析，200字以内",
  "conflict_areas": ["价值观冲突领域1"],
  "recommendation": "基于动机价值观的职业建议，150字以内"
}`,
	},
	{
		Name:        "skill_heatmap",
		Title:       "技能热力图与胜任力模型",
		Description: "6维度技能评分+胜任力模型",
		Prompt: `你是一位专业的技能评估专家。请根据以下用户画像，生成技能热力图和胜任力模型。

用户画像：
{{.user_profile}}

请严格按以下 JSON 格式输出，不要输出任何其他内容：
{
  "skill_dimensions": {
    "technical_skills": {"score": 4, "details": ["技能1", "技能2"]},
    "management_skills": {"score": 3, "details": ["技能1"]},
    "communication_skills": {"score": 4, "details": ["技能1", "技能2"]},
    "innovation_skills": {"score": 3, "details": ["技能1"]},
    "analytical_skills": {"score": 4, "details": ["技能1", "技能2"]},
    "leadership_skills": {"score": 3, "details": ["技能1"]}
  },
  "competency_model": {
    "core_competencies": ["核心胜任力1", "核心胜任力2"],
    "differentiating_competencies": ["差异化胜任力1"],
    "threshold_competencies": ["基本胜任力1"]
  },
  "skill_gaps": [
    {"skill": "缺失技能", "importance": "高/中/低", "development_path": "发展路径"}
  ],
  "skill_development_priority": ["优先发展1", "优先发展2"],
  "competency_assessment": "胜任力综合评估，200字以内"
}`,
	},
	{
		Name:        "interest_assessment",
		Title:       "职业兴趣评估",
		Description: "基于霍兰德理论的6维度职业兴趣评估",
		Prompt: `你是一位专业的职业兴趣评估师，擅长霍兰德职业兴趣理论。请根据以下用户画像，评估用户的职业兴趣。

用户画像：
{{.user_profile}}

请严格按以下 JSON 格式输出，不要输出任何其他内容：
{
  "holland_codes": {
    "R_realistic": {"score": 3, "description": "现实型描述"},
    "I_investigative": {"score": 4, "description": "研究型描述"},
    "A_artistic": {"score": 3, "description": "艺术型描述"},
    "S_social": {"score": 5, "description": "社会型描述"},
    "E_enterprising": {"score": 3, "description": "企业型描述"},
    "C_conventional": {"score": 2, "description": "常规型描述"}
  },
  "top_holland_code": "SEA",
  "interest_profile": "兴趣画像描述，200字以内",
  "suitable_work_environments": ["适合的工作环境1", "适合的工作环境2"],
  "unsuitable_work_environments": ["不适合的工作环境1"],
  "interest_career_alignment": "兴趣与职业匹配分析，200字以内",
  "development_suggestions": ["兴趣发展建议1", "兴趣发展建议2"]
}`,
	},
	{
		Name:        "career_recommendations",
		Title:       "基于兴趣的职业推荐",
		Description: "推荐匹配的职业方向及详细分析",
		Prompt: `你是一位专业的职业推荐顾问。请根据以下用户画像，推荐最适合的职业方向。

用户画像：
{{.user_profile}}

请严格按以下 JSON 格式输出，不要输出任何其他内容：
{
  "recommended_careers": [
    {
      "title": "推荐职位名称",
      "match_score": 85,
      "description": "职位描述，100字以内",
      "requirements": ["要求1", "要求2"],
      "prospects": "发展前景，100字以内",
      "salary_range": "薪资范围",
      "success_story": "成功案例简述，80字以内",
      "entry_path": "入行路径，80字以内"
    }
  ],
  "career_change_options": [
    {
      "direction": "转型方向",
      "feasibility": "高/中/低",
      "preparation_needed": "所需准备"
    }
  ],
  "emerging_opportunities": ["新兴机会1", "新兴机会2"],
  "overall_recommendation": "综合推荐建议，200字以内"
}`,
	},
	{
		Name:        "industry_analysis",
		Title:       "行业分析",
		Description: "当前行业现状、趋势与市场需求",
		Prompt: `你是一位专业的行业研究分析师。请根据以下用户画像，分析用户所在行业及相关行业的情况。

用户画像：
{{.user_profile}}

请严格按以下 JSON 格式输出，不要输出任何其他内容：
{
  "current_industry": {
    "name": "行业名称",
    "status": "行业现状描述，150字以内",
    "market_size": "市场规模估算",
    "growth_rate": "增长率估算",
    "key_players": ["头部企业1", "头部企业2"]
  },
  "industry_trends": [
    {"trend": "趋势描述", "impact": "对用户的影响", "timeframe": "短期/中期/长期"}
  ],
  "market_demand": {
    "hot_skills": ["热门技能1", "热门技能2"],
    "talent_gap": "人才缺口描述",
    "salary_trend": "薪资趋势描述"
  },
  "related_industries": [
    {"name": "相关行业", "connection": "关联性", "opportunity": "机会描述"}
  ],
  "risk_factors": ["风险因素1", "风险因素2"],
  "industry_advice": "行业建议，200字以内"
}`,
	},
	{
		Name:        "goal_setting",
		Title:       "目标设定",
		Description: "短期/中期/长期职业目标",
		Prompt: `你是一位专业的职业目标规划师。请根据以下用户画像，为用户设定短期、中期和长期职业目标。

用户画像：
{{.user_profile}}

请严格按以下 JSON 格式输出，不要输出任何其他内容：
{
  "short_term_goals": [
    {
      "goal": "目标描述",
      "timeframe": "1年内",
      "specific_actions": ["具体行动1", "具体行动2"],
      "success_metrics": ["衡量标准1"],
      "resources_needed": ["所需资源1"]
    }
  ],
  "medium_term_goals": [
    {
      "goal": "目标描述",
      "timeframe": "1-3年",
      "specific_actions": ["具体行动1", "具体行动2"],
      "success_metrics": ["衡量标准1"],
      "resources_needed": ["所需资源1"]
    }
  ],
  "long_term_goals": [
    {
      "goal": "目标描述",
      "timeframe": "3-5年",
      "specific_actions": ["具体行动1", "具体行动2"],
      "success_metrics": ["衡量标准1"],
      "resources_needed": ["所需资源1"]
    }
  ],
  "goal_alignment": "目标一致性分析，150字以内",
  "milestone_checkpoints": ["里程碑1", "里程碑2"],
  "goal_adjustment_triggers": ["需要调整目标的情况1"]
}`,
	},
	{
		Name:        "action_plan",
		Title:       "行动计划",
		Description: "学习/人脉/自我反思的具体行动方案",
		Prompt: `你是一位专业的职业发展行动规划师。请根据以下用户画像，制定具体的行动计划。

用户画像：
{{.user_profile}}

请严格按以下 JSON 格式输出，不要输出任何其他内容：
{
  "learning_plan": [
    {
      "area": "学习领域",
      "method": "学习方式",
      "timeline": "时间安排",
      "expected_outcome": "预期成果",
      "resources": ["推荐资源1", "推荐资源2"]
    }
  ],
  "networking_plan": [
    {
      "target": "人脉目标",
      "approach": "拓展方式",
      "timeline": "时间安排",
      "expected_value": "预期价值"
    }
  ],
  "self_reflection_plan": {
    "frequency": "反思频率（如每周/每月）",
    "methods": ["反思方法1", "反思方法2"],
    "focus_areas": ["关注领域1", "关注领域2"],
    "journaling_template": "日志模板建议"
  },
  "daily_habits": ["推荐日常习惯1", "推荐日常习惯2"],
  "weekly_tasks": ["推荐周任务1", "推荐周任务2"],
  "monthly_reviews": ["推荐月度复盘1", "推荐月度复盘2"],
  "accountability_suggestions": "自我监督建议，100字以内"
}`,
	},
	{
		Name:        "summary_outlook",
		Title:       "总结与展望",
		Description: "报告总结与未来展望",
		Prompt: `你是一位专业的职业规划总结师。请根据以下用户画像和报告各章节分析，撰写报告总结与展望。

用户画像：
{{.user_profile}}

请严格按以下 JSON 格式输出，不要输出任何其他内容：
{
  "key_findings": ["核心发现1", "核心发现2", "核心发现3"],
  "strengths_summary": "优势总结，150字以内",
  "areas_for_growth": "成长空间总结，150字以内",
  "career_direction": "推荐职业方向总结，150字以内",
  "next_steps": ["下一步行动1", "下一步行动2", "下一步行动3"],
  "encouragement": "激励寄语，100字以内，温暖有力量",
  "outlook": "未来展望，150字以内，积极正面",
  "final_blessing": "最后的祝福语，50字以内"
}`,
	},
}

// GetReportSectionNames 返回所有章节的英文标识列表
func GetReportSectionNames() []string {
	names := make([]string, len(ReportSections))
	for i, s := range ReportSections {
		names[i] = s.Name
	}
	return names
}

// GetReportSectionByName 根据英文标识查找章节定义
func GetReportSectionByName(name string) *ReportSection {
	for i := range ReportSections {
		if ReportSections[i].Name == name {
			return &ReportSections[i]
		}
	}
	return nil
}
