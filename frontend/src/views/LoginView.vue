<template>
  <div class="auth-container">
    <div class="auth-card">
      <h1 class="logo">📚 Webook</h1>
      
      <!-- 登录表单 -->
      <div v-if="isLogin" class="form-section">
        <h2>登录</h2>
        <form @submit.prevent="handleLogin">
          <div class="form-group">
            <label>邮箱</label>
            <input v-model="loginForm.email" type="email" placeholder="请输入邮箱" required />
          </div>
          <div class="form-group">
            <label>密码</label>
            <input v-model="loginForm.password" type="password" placeholder="请输入密码" required />
          </div>
          <button type="submit" class="btn-primary" :disabled="loading">
            {{ loading ? '登录中...' : '登录' }}
          </button>
        </form>
        <p class="switch-text">
          还没有账号？<a @click="isLogin = false">立即注册</a>
        </p>
      </div>

      <!-- 注册表单 -->
      <div v-else class="form-section">
        <h2>注册</h2>
        <form @submit.prevent="handleSignup">
          <div class="form-group">
            <label>邮箱</label>
            <input v-model="signupForm.email" type="email" placeholder="请输入邮箱" required />
          </div>
          <div class="form-group">
            <label>密码</label>
            <input v-model="signupForm.password" type="password" placeholder="6-16位密码" required />
          </div>
          <div class="form-group">
            <label>确认密码</label>
            <input v-model="signupForm.confirmPassword" type="password" placeholder="请再次输入密码" required />
          </div>
          <button type="submit" class="btn-primary" :disabled="loading">
            {{ loading ? '注册中...' : '注册' }}
          </button>
        </form>
        <p class="switch-text">
          已有账号？<a @click="isLogin = true">立即登录</a>
        </p>
      </div>

      <p v-if="message" :class="['message', messageType]">{{ message }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { login, signup } from '../api/user'

const router = useRouter()
const isLogin = ref(true)
const loading = ref(false)
const message = ref('')
const messageType = ref('error')

const loginForm = ref({ email: '', password: '' })
const signupForm = ref({ email: '', password: '', confirmPassword: '' })

async function handleLogin() {
  loading.value = true
  message.value = ''
  try {
    const res = await login(loginForm.value.email, loginForm.value.password)
    if (res.code === 0) {
      localStorage.setItem('userId', res.data.userId)
      router.push('/profile')
    } else {
      message.value = res.msg || '登录失败'
      messageType.value = 'error'
    }
  } catch (e) {
    message.value = '网络错误'
    messageType.value = 'error'
  }
  loading.value = false
}

async function handleSignup() {
  loading.value = true
  message.value = ''
  try {
    const res = await signup(
      signupForm.value.email,
      signupForm.value.password,
      signupForm.value.confirmPassword
    )
    if (res.code === 0) {
      message.value = '注册成功，请登录'
      messageType.value = 'success'
      isLogin.value = true
      signupForm.value = { email: '', password: '', confirmPassword: '' }
    } else {
      message.value = res.msg || '注册失败'
      messageType.value = 'error'
    }
  } catch (e) {
    message.value = '网络错误'
    messageType.value = 'error'
  }
  loading.value = false
}
</script>

<style scoped>
.auth-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.auth-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 20px;
  padding: 40px;
  width: 400px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
}

.logo {
  text-align: center;
  font-size: 2rem;
  margin-bottom: 30px;
  color: #667eea;
}

h2 {
  text-align: center;
  margin-bottom: 25px;
  color: #333;
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
  border-color: #667eea;
}

.btn-primary {
  width: 100%;
  padding: 14px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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
  box-shadow: 0 10px 20px rgba(102, 126, 234, 0.4);
}

.btn-primary:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.switch-text {
  text-align: center;
  margin-top: 20px;
  color: #666;
}

.switch-text a {
  color: #667eea;
  cursor: pointer;
  font-weight: 600;
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
