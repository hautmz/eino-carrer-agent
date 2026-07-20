<template>
  <div class="file-upload">
    <el-upload
      :auto-upload="false"
      :limit="1"
      :on-change="handleFileChange"
      :before-upload="beforeUpload"
      accept=".pdf,.docx"
    >
      <el-button :icon="Upload" size="small">上传简历</el-button>
    </el-upload>
    <div v-if="uploadedFileId" class="upload-result">
      <el-tag type="success" size="small">文件已上传 (ID: {{ uploadedFileId }})</el-tag>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Upload } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { uploadFile } from '../api/upload'

const emit = defineEmits(['uploaded'])

const uploadedFileId = ref(null)

function beforeUpload(file) {
  const maxSize = 10 * 1024 * 1024
  if (file.size > maxSize) {
    ElMessage.error('文件大小不能超过10MB')
    return false
  }
  const ext = file.name.split('.').pop().toLowerCase()
  if (!['pdf', 'docx'].includes(ext)) {
    ElMessage.error('仅支持 PDF 和 DOCX 格式')
    return false
  }
  return true
}

async function handleFileChange(file) {
  if (!beforeUpload(file)) return

  const formData = new FormData()
  formData.append('file', file.raw)

  try {
    const res = await uploadFile(formData)
    uploadedFileId.value = res.data.id
    ElMessage.success('文件上传成功')
    emit('uploaded', res.data.id)
  } catch {
    ElMessage.error('文件上传失败')
  }
}
</script>

<style scoped>
.file-upload {
  display: inline-block;
}

.upload-result {
  margin-top: 4px;
}
</style>
