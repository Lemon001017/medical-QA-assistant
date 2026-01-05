<template>
  <div class="register-container">
    <div class="register-card">
      <h1>医学问答助手</h1>
      <h2>注册</h2>
      <form @submit.prevent="handleRegister">
        <div class="form-group">
          <label for="username">用户名</label>
          <input
            id="username"
            v-model="username"
            type="text"
            required
            minlength="3"
            maxlength="50"
            placeholder="请输入用户名（3-50个字符）"
          />
        </div>
        <div class="form-group">
          <label for="email">邮箱</label>
          <input
            id="email"
            v-model="email"
            type="email"
            required
            placeholder="请输入邮箱地址"
          />
        </div>
        <div class="form-group">
          <label for="password">密码</label>
          <input
            id="password"
            v-model="password"
            type="password"
            required
            minlength="6"
            placeholder="请输入密码（至少6个字符）"
          />
        </div>
        <div class="form-group">
          <label for="confirmPassword">确认密码</label>
          <input
            id="confirmPassword"
            v-model="confirmPassword"
            type="password"
            required
            placeholder="请再次输入密码"
          />
        </div>
        <div v-if="error" class="error-message">{{ error }}</div>
        <button type="submit" :disabled="loading" class="submit-btn">
          {{ loading ? '注册中...' : '注册' }}
        </button>
        <div class="link-text">
          已有账号？
          <router-link to="/login">立即登录</router-link>
        </div>
      </form>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

export default {
  name: 'Register',
  setup() {
    const router = useRouter()
    const authStore = useAuthStore()
    
    const username = ref('')
    const email = ref('')
    const password = ref('')
    const confirmPassword = ref('')
    const error = ref('')
    const loading = ref(false)

    const handleRegister = async () => {
      error.value = ''

      // Validate password match
      if (password.value !== confirmPassword.value) {
        error.value = '两次输入的密码不一致'
        return
      }

      loading.value = true

      try {
        await authStore.register(username.value, email.value, password.value)
        router.push('/home')
      } catch (err) {
        error.value = err
      } finally {
        loading.value = false
      }
    }

    return {
      username,
      email,
      password,
      confirmPassword,
      error,
      loading,
      handleRegister
    }
  }
}
</script>

<style scoped>
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  padding: 20px;
}

.register-card {
  background: white;
  padding: 40px;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  width: 100%;
  max-width: 400px;
}

h1 {
  text-align: center;
  color: #2c3e50;
  margin-bottom: 10px;
  font-size: 24px;
}

h2 {
  text-align: center;
  color: #7f8c8d;
  margin-bottom: 30px;
  font-size: 20px;
  font-weight: normal;
}

.form-group {
  margin-bottom: 20px;
}

label {
  display: block;
  margin-bottom: 8px;
  color: #34495e;
  font-weight: 500;
}

input {
  width: 100%;
  padding: 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  transition: border-color 0.3s;
}

input:focus {
  outline: none;
  border-color: #3498db;
}

.error-message {
  color: #e74c3c;
  margin-bottom: 15px;
  font-size: 14px;
  text-align: center;
}

.submit-btn {
  width: 100%;
  padding: 12px;
  background-color: #3498db;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.3s;
}

.submit-btn:hover:not(:disabled) {
  background-color: #2980b9;
}

.submit-btn:disabled {
  background-color: #95a5a6;
  cursor: not-allowed;
}

.link-text {
  text-align: center;
  margin-top: 20px;
  color: #7f8c8d;
  font-size: 14px;
}

.link-text a {
  color: #3498db;
  text-decoration: none;
}

.link-text a:hover {
  text-decoration: underline;
}
</style>
