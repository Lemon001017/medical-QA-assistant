<template>
  <div class="home-container">
    <header class="header">
      <h1>医学问答助手</h1>
      <div class="user-info">
        <span>欢迎，{{ user?.username }}</span>
        <button @click="handleLogout" class="logout-btn">退出</button>
      </div>
    </header>
    <main class="main-content">
      <div class="welcome-card">
        <h2>欢迎使用医学问答助手</h2>
        <p>您已成功登录！</p>
        <p>更多功能即将推出...</p>
      </div>
    </main>
  </div>
</template>

<script>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

export default {
  name: 'Home',
  setup() {
    const router = useRouter()
    const authStore = useAuthStore()
    
    const user = computed(() => authStore.user)

    const handleLogout = () => {
      authStore.logout()
      router.push('/login')
    }

    return {
      user,
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

.welcome-card {
  background: white;
  padding: 40px;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  text-align: center;
}

.welcome-card h2 {
  color: #2c3e50;
  margin-bottom: 20px;
}

.welcome-card p {
  color: #7f8c8d;
  margin-bottom: 10px;
  font-size: 16px;
}
</style>
