<template>
  <div class="profile-container">
    <div class="profile-card">
      <h1>👤 用户信息</h1>
      
      <div class="info-section">
        <div class="info-item">
          <span class="label">用户ID</span>
          <span class="value">{{ userId }}</span>
        </div>
        <div class="info-item">
          <span class="label">邮箱</span>
          <span class="value">{{ email }}</span>
        </div>
      </div>

      <hr />

      <h2>🔐 修改密码</h2>
      <form @submit.prevent="handleUpdatePassword">
        <div class="form-group">
          <label>旧密码</label>
          <input v-model="passwordForm.oldPassword" type="password" placeholder="请输入旧密码" required />
        </div>
        <div class="form-group">
          <label>新密码</label>
          <input v-model="passwordForm.newPassword" type="password" placeholder="6-16位新密码" required />
        </div>
        <button type="submit" class="btn-primary" :disabled="loading">
          {{ loading ? '修改中...' : '修改密码' }}
        </button>
      </form>

      <p v-if="message" :class="['message', messageType]">{{ message }}</p>

      <button @click="handleLogout" class="btn-logout">退出登录</button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getProfile, updatePassword } from '../api/user'

const router = useRouter()
const userId = ref(localStorage.getItem('userId'))
const email = ref('')
const loading = ref(false)
const message = ref('')
const messageType = ref('error')
const passwordForm = ref({ oldPassword: '', newPassword: '' })

onMounted(async () => {
  if (!userId.value) {
    router.push('/')
    return
  }
  try {
    const res = await getProfile(userId.value)
    if (res.code === 0) {
      email.value = res.data.email
    } else {
      message.value = res.msg || '获取信息失败'
    }
  } catch (e) {
    message.value = '网络错误'
  }
})

async function handleUpdatePassword() {
  loading.value = true
  message.value = ''
  try {
    const res = await updatePassword(
      userId.value,
      passwordForm.value.oldPassword,
      passwordForm.value.newPassword
    )
    if (res.code === 0) {
      message.value = '密码修改成功'
      messageType.value = 'success'
      passwordForm.value = { oldPassword: '', newPassword: '' }
    } else {
      message.value = res.msg || '修改失败'
      messageType.value = 'error'
    }
  } catch (e) {
    message.value = '网络错误'
    messageType.value = 'error'
  }
  loading.value = false
}

function handleLogout() {
  localStorage.removeItem('userId')
  router.push('/')
}
</script>

<style scoped>
.profile-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
}

.profile-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 20px;
  padding: 40px;
  width: 450px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
}

h1 {
  text-align: center;
  color: #11998e;
  margin-bottom: 30px;
}

h2 {
  color: #333;
  margin: 20px 0;
}

hr {
  border: none;
  border-top: 1px solid #eee;
  margin: 25px 0;
}

.info-section {
  background: #f8f9fa;
  border-radius: 12px;
  padding: 20px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  padding: 10px 0;
}

.info-item .label {
  color: #666;
}

.info-item .value {
  color: #333;
  font-weight: 600;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  color: #555;
  font-weight: 500;
}

.form-group input {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #e0e0e0;
  border-radius: 10px;
  font-size: 14px;
  transition: border-color 0.3s;
  box-sizing: border-box;
}

.form-group input:focus {
  outline: none;
  border-color: #11998e;
}

.btn-primary {
  width: 100%;
  padding: 14px;
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
  color: white;
  border: none;
  border-radius: 10px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 10px 20px rgba(17, 153, 142, 0.4);
}

.btn-primary:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.btn-logout {
  width: 100%;
  padding: 14px;
  background: #f5f5f5;
  color: #666;
  border: none;
  border-radius: 10px;
  font-size: 16px;
  cursor: pointer;
  margin-top: 15px;
  transition: background 0.3s;
}

.btn-logout:hover {
  background: #eee;
}

.message {
  text-align: center;
  padding: 10px;
  border-radius: 8px;
  margin-top: 15px;
}

.message.error {
  background: #fee;
  color: #c00;
}

.message.success {
  background: #efe;
  color: #060;
}
</style>
