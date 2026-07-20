<template>
  <div class="report-view">
    <el-collapse v-model="expandedSections">
      <el-collapse-item
        v-for="section in sections"
        :key="section.key"
        :title="section.title"
        :name="section.key"
      >
        <div v-if="section.content" class="section-content" v-html="renderMarkdown(section.content)" />
        <div v-else class="section-empty">暂无数据</div>
      </el-collapse-item>
    </el-collapse>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import MarkdownIt from 'markdown-it'

const md = new MarkdownIt({ html: false, breaks: true })

const props = defineProps({
  report: { type: Object, required: true },
})

const sectionDefs = [
  { key: 'professional_index', title: '专业指数评分' },
  { key: 'myself_report', title: '个人信息提取' },
  { key: 'achievement_superiority', title: '成就与优势' },
  { key: 'career_experience', title: '职业经历与成长路径' },
  { key: 'motivation_values', title: '动机与价值观评估' },
  { key: 'skill_heatmap', title: '技能热力图' },
  { key: 'interest_assessment', title: '职业兴趣评估' },
  { key: 'career_recommendations', title: '职业推荐' },
  { key: 'industry_analysis', title: '行业分析' },
  { key: 'goal_setting', title: '目标设定' },
  { key: 'action_plan', title: '行动计划' },
  { key: 'summary_outlook', title: '总结与展望' },
]

const sections = computed(() =>
  sectionDefs.map((s) => ({
    ...s,
    content: props.report[s.key] || '',
  }))
)

const expandedSections = ref(sectionDefs.slice(0, 3).map((s) => s.key))

function renderMarkdown(text) {
  if (!text) return ''
  try {
    let display = text
    try {
      const parsed = JSON.parse(text)
      display = JSON.stringify(parsed, null, 2)
    } catch {}
    return md.render(display)
  } catch {
    return text
  }
}
</script>

<style scoped>
.report-view {
  padding: 12px;
}

.section-content {
  font-size: 14px;
  line-height: 1.8;
}

.section-content :deep(pre) {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 12px;
  border-radius: 6px;
  overflow-x: auto;
  font-size: 13px;
}

.section-empty {
  color: #999;
  font-style: italic;
}
</style>
