<template>
  <div class="qa-container">
    <header class="header">
      <h1>医学问答</h1>
      <div class="user-info">
        <span>欢迎，{{ user?.username }}</span>
        <div class="actions">
          <router-link to="/home" class="link-btn">文档</router-link>
          <button @click="handleLogout" class="logout-btn">退出</button>
        </div>
      </div>
    </header>
    <main class="main-content">
      <section class="panel">
        <h2>提问</h2>
        <form @submit.prevent="handleAsk">
          <div class="form-group">
            <label for="question">问题</label>
            <textarea
              id="question"
              v-model="question"
              rows="6"
              required
              placeholder="请输入医学相关问题"
              :disabled="loading"
            ></textarea>
          </div>
          <div class="actions">
            <button type="submit" :disabled="loading">
              {{ loading ? '思考中...' : '发送' }}
            </button>
            <button v-if="loading" type="button" @click="handleStop" class="stop-btn">
              停止
            </button>
          </div>
        </form>
      </section>

      <section class="panel" v-if="answer || error">
        <h2>回答</h2>
        <div v-if="answer" class="answer">{{ answer }}</div>
        <p v-else class="error">{{ error }}</p>
      </section>
    </main>
  </div>
</template>

<script>
import { computed, ref, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

export default {
  name: 'QA',
  setup() {
    const router = useRouter()
    const authStore = useAuthStore()

    const user = computed(() => authStore.user)
    const question = ref('')
    const answer = ref('')
    const error = ref('')
    const loading = ref(false)
    let abortController = null

    const handleAsk = () => {
      if (!question.value.trim()) return
      loading.value = true
      answer.value = ''
      error.value = ''

      // Create abort controller for cancellation
      abortController = new AbortController()

      // Use fetch with POST for SSE (EventSource doesn't support POST)
      const baseURL = import.meta.env.VITE_API_BASE_URL || '/api/v1'
      fetch(`${baseURL}/qa/ask/stream`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authStore.token}`
        },
        body: JSON.stringify({ question: question.value }),
        signal: abortController.signal
      })
        .then(response => {
          if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`)
          }

          const reader = response.body.getReader()
          const decoder = new TextDecoder()
          let buffer = ''

          const readStream = () => {
            reader.read().then(({ done, value }) => {
              if (done) {
                loading.value = false
                abortController = null
                return
              }

              buffer += decoder.decode(value, { stream: true })
              const lines = buffer.split('\n')
              buffer = lines.pop() || ''

              for (const line of lines) {
                if (line.startsWith('data: ')) {
                  const data = line.slice(6)
                  if (data.trim()) {
                    try {
                      const parsed = JSON.parse(data)
                      if (parsed.error) {
                        error.value = parsed.error
                        loading.value = false
                        abortController = null
                        return
                      }
                      if (parsed.done) {
                        loading.value = false
                        abortController = null
                        return
                      }
                      if (parsed.chunk) {
                        answer.value += parsed.chunk
                      }
                    } catch (e) {
                      console.error('Failed to parse SSE data:', e)
                    }
                  }
                }
              }

              readStream()
            }).catch(err => {
              if (err.name !== 'AbortError') {
                error.value = err.message || '流式请求失败'
                loading.value = false
              }
              abortController = null
            })
          }

          readStream()
        })
        .catch(err => {
          if (err.name !== 'AbortError') {
            error.value = err.message || '请求失败'
            loading.value = false
          }
        })
    }

    const handleStop = () => {
      if (abortController) {
        abortController.abort()
        abortController = null
      }
      loading.value = false
    }

    const handleLogout = () => {
      handleStop()
      authStore.logout()
      router.push('/login')
    }

    onUnmounted(() => {
      handleStop()
    })

    return {
      user,
      question,
      answer,
      error,
      loading,
      handleAsk,
      handleStop,
      handleLogout
    }
  }
}
</script>

<style scoped>
.qa-container {
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
  gap: 12px;
}

.actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.logout-btn,
.link-btn,
button[type='submit'] {
  padding: 8px 14px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.logout-btn {
  background-color: #e74c3c;
  color: white;
}

.link-btn {
  background-color: #3498db;
  color: white;
  text-decoration: none;
}

.main-content {
  max-width: 1000px;
  margin: 40px auto;
  padding: 0 20px;
  display: flex;
  flex-direction: column;
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
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  color: #34495e;
}

.form-group textarea {
  padding: 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 14px;
  resize: vertical;
}

.actions button {
  background-color: #3498db;
  color: white;
}

.actions button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.stop-btn {
  background-color: #e74c3c !important;
}

.answer {
  white-space: pre-wrap;
  line-height: 1.6;
  color: #2c3e50;
  min-height: 20px;
}

.error {
  color: #e74c3c;
}
</style>
