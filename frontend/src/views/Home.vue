<template>
  <div class="home-container">
    <header class="header">
      <h1>医学问答助手</h1>
      <div class="user-info">
        <span>欢迎，{{ user?.username }}</span>
        <router-link to="/qa" class="link-btn">问答</router-link>
        <button @click="handleLogout" class="logout-btn">退出</button>
      </div>
    </header>
    <main class="main-content">
      <div class="layout">
        <section class="panel form-panel">
          <h2>上传/创建文档</h2>
          <form @submit.prevent="handleCreate">
            <div class="form-group">
              <label for="title">标题</label>
              <input
                id="title"
                v-model="form.title"
                type="text"
                required
                maxlength="255"
                placeholder="请输入文档标题"
              />
            </div>
            <div class="form-group">
              <label for="content">内容</label>
              <textarea
                id="content"
                v-model="form.content"
                rows="6"
                required
                placeholder="输入或粘贴文档内容"
              ></textarea>
            </div>
            <div class="actions">
              <button type="submit" :disabled="creating">
                {{ creating ? '提交中...' : '创建文档' }}
              </button>
            </div>
          </form>

          <div class="divider"></div>

          <div class="upload-section">
            <h3>上传文件</h3>
            <div class="form-group">
              <label for="uploadTitle">标题（可选，默认使用文件名）</label>
              <input
                id="uploadTitle"
                v-model="uploadTitle"
                type="text"
                maxlength="255"
                placeholder="请输入标题"
              />
            </div>
            <div class="form-group">
              <label for="file">文件</label>
              <input id="file" type="file" @change="onFileChange" />
              <p class="hint-text">支持文本类文件，将以纯文本存储。</p>
            </div>
            <div class="actions">
              <button type="button" @click="handleUpload" :disabled="uploading">
                {{ uploading ? '上传中...' : '上传文档' }}
              </button>
              <span v-if="error" class="error">{{ error }}</span>
              <span v-if="success" class="success">{{ success }}</span>
            </div>
          </div>
        </section>

        <section class="panel list-panel">
          <div class="list-header">
            <h2>我的文档</h2>
            <button class="refresh-btn" @click="fetchDocs" :disabled="loading">
              {{ loading ? '刷新中...' : '刷新' }}
            </button>
          </div>
          <div v-if="loading" class="hint">加载中...</div>
          <div v-else-if="docs.length === 0" class="hint">暂无文档</div>
          <ul v-else class="doc-list">
            <li v-for="doc in docs" :key="doc.id" class="doc-item">
              <div class="doc-meta">
                <h3>{{ doc.title }}</h3>
                <p class="content-preview">{{ doc.content }}</p>
                <p class="status">状态：{{ doc.status }}</p>
                <p class="timestamp">更新：{{ formatDate(doc.updated_at || doc.updatedAt) }}</p>
              </div>
              <button class="delete-btn" @click="handleDelete(doc.id)" :disabled="deletingId === doc.id">
                {{ deletingId === doc.id ? '删除中...' : '删除' }}
              </button>
            </li>
          </ul>
        </section>
      </div>
    </main>
  </div>
</template>

<script>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { fetchDocuments, createDocument, deleteDocument, uploadDocument } from '../api'

export default {
  name: 'Home',
  setup() {
    const router = useRouter()
    const authStore = useAuthStore()
    
    const user = computed(() => authStore.user)
    const docs = ref([])
    const loading = ref(false)
    const creating = ref(false)
    const uploading = ref(false)
    const deletingId = ref(null)
    const error = ref('')
    const success = ref('')
    const form = reactive({
      title: '',
      content: ''
    })
    const uploadTitle = ref('')
    const fileRef = ref(null)

    const resetMessages = () => {
      error.value = ''
      success.value = ''
    }

    const fetchDocs = async () => {
      resetMessages()
      loading.value = true
      try {
        const { data } = await fetchDocuments()
        docs.value = data || []
      } catch (err) {
        error.value = err?.response?.data?.error || '获取文档失败'
      } finally {
        loading.value = false
      }
    }

    const handleCreate = async () => {
      resetMessages()
      creating.value = true
      try {
        await createDocument({ title: form.title, content: form.content })
        success.value = '创建成功'
        form.title = ''
        form.content = ''
        await fetchDocs()
      } catch (err) {
        error.value = err?.response?.data?.error || '创建文档失败'
      } finally {
        creating.value = false
      }
    }

    const onFileChange = (event) => {
      const [file] = event.target.files || []
      fileRef.value = file || null
    }

    const handleUpload = async () => {
      resetMessages()
      if (!fileRef.value) {
        error.value = '请选择文件'
        return
      }
      uploading.value = true
      try {
        const formData = new FormData()
        if (uploadTitle.value) {
          formData.append('title', uploadTitle.value)
        }
        formData.append('file', fileRef.value)
        await uploadDocument(formData)
        success.value = '上传成功'
        uploadTitle.value = ''
        fileRef.value = null
        await fetchDocs()
      } catch (err) {
        error.value = err?.response?.data?.error || '上传失败'
      } finally {
        uploading.value = false
      }
    }

    const handleDelete = async (id) => {
      resetMessages()
      deletingId.value = id
      try {
        await deleteDocument(id)
        success.value = '已删除'
        docs.value = docs.value.filter((d) => d.id !== id)
      } catch (err) {
        error.value = err?.response?.data?.error || '删除失败'
      } finally {
        deletingId.value = null
      }
    }

    const formatDate = (value) => {
      if (!value) return ''
      return new Date(value).toLocaleString()
    }

    const handleLogout = () => {
      authStore.logout()
      router.push('/login')
    }

    onMounted(fetchDocs)

    return {
      user,
      docs,
      loading,
      creating,
      uploading,
      deletingId,
      form,
      uploadTitle,
      error,
      success,
      fetchDocs,
      handleCreate,
      handleDelete,
      onFileChange,
      handleUpload,
      formatDate,
      handleLogout
    }
  }
}
</script>

<style scoped>
.home-container {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.header {
  background: white;
  padding: 20px 40px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header h1 {
  color: #2c3e50;
  font-size: 24px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 20px;
}

.user-info span {
  color: #34495e;
}

.logout-btn {
  padding: 8px 16px;
  background-color: #e74c3c;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.3s;
}

.logout-btn:hover {
  background-color: #c0392b;
}

.main-content {
  max-width: 1200px;
  margin: 40px auto;
  padding: 0 20px;
}

.layout {
  display: grid;
  grid-template-columns: 1fr 1.2fr;
  gap: 20px;
}

.panel {
  background: white;
  padding: 24px;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.panel h2 {
  color: #2c3e50;
  margin-bottom: 16px;
}

.form-group {
  margin-bottom: 16px;
  display: flex;
  flex-direction: column;
}

.form-group label {
  margin-bottom: 8px;
  color: #34495e;
}

.form-group input,
.form-group textarea {
  padding: 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 14px;
  transition: border-color 0.2s;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #3498db;
}

.actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.actions button {
  padding: 10px 18px;
  background-color: #3498db;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.actions button:disabled {
  background-color: #95a5a6;
  cursor: not-allowed;
}

.actions button:hover:not(:disabled) {
  background-color: #2980b9;
}

.error {
  color: #e74c3c;
  font-size: 13px;
}

.success {
  color: #2ecc71;
  font-size: 13px;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.refresh-btn {
  padding: 8px 14px;
  background: #3498db;
  color: #fff;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

.doc-list {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.doc-item {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 16px;
  border: 1px solid #eee;
  border-radius: 8px;
}

.doc-meta h3 {
  margin-bottom: 8px;
  color: #2c3e50;
}

.content-preview {
  color: #7f8c8d;
  margin-bottom: 6px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.status,
.timestamp {
  color: #95a5a6;
  font-size: 13px;
}

.delete-btn {
  background: #e74c3c;
  color: white;
  border: none;
  border-radius: 6px;
  padding: 8px 12px;
  cursor: pointer;
  min-width: 88px;
}

.delete-btn:disabled {
  background: #c0392b;
  opacity: 0.8;
  cursor: not-allowed;
}

.hint {
  color: #7f8c8d;
}

.hint-text {
  color: #95a5a6;
  font-size: 13px;
  margin-top: 6px;
}

.divider {
  height: 1px;
  background: #f0f0f0;
  margin: 16px 0 12px;
}

.upload-section h3 {
  margin-bottom: 12px;
  color: #2c3e50;
}

@media (max-width: 960px) {
  .layout {
    grid-template-columns: 1fr;
  }
}
</style>
