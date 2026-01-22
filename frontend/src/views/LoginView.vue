<template>
  <div class="auth-page">
    <div class="auth-split-layout">
      <!-- Left: Editorial/Brand Section -->
      <section class="brand-section">
        <div class="brand-content stagger-1">
          <div class="brand-icon">W</div>
          <h1 class="brand-title">Webook</h1>
          <p class="brand-tagline">Where thoughts find their home.</p>
        </div>
        <div class="brand-footer stagger-2">
          <p>© 2026 Webook Editorial</p>
        </div>
      </section>

      <!-- Right: Form Section -->
      <section class="form-section">
        <div class="form-container stagger-3">
          <div v-if="isLogin" class="auth-mode">
            <h2 class="form-title">Welcome back</h2>
            <p class="form-subtitle">Sign in to continue your writing journey.</p>
            
            <form @submit.prevent="handleLogin" class="auth-form">
              <div class="form-group">
                <label class="form-label">Email</label>
                <input 
                  v-model="loginForm.email" 
                  type="email" 
                  class="input-text"
                  placeholder="name@example.com" 
                  required 
                />
              </div>
              
              <div class="form-group">
                <label class="form-label">Password</label>
                <input 
                  v-model="loginForm.password" 
                  type="password" 
                  class="input-text"
                  placeholder="••••••••" 
                  required 
                />
              </div>
              
              <button type="submit" class="btn btn-primary btn-block" :disabled="loading">
                {{ loading ? 'Signing in...' : 'Sign In' }}
              </button>
            </form>
            
            <div class="form-footer">
              <span>New to Webook?</span>
              <a @click="isLogin = false" class="switch-link">Create an account</a>
            </div>
          </div>

          <div v-else class="auth-mode">
            <h2 class="form-title">Join Webook</h2>
            <p class="form-subtitle">Start curating your thoughts today.</p>
            
            <form @submit.prevent="handleSignup" class="auth-form">
              <div class="form-group">
                <label class="form-label">Email</label>
                <input 
                  v-model="signupForm.email" 
                  type="email" 
                  class="input-text"
                  placeholder="name@example.com" 
                  required 
                />
              </div>
              
              <div class="form-group">
                <label class="form-label">Password</label>
                <input 
                  v-model="signupForm.password" 
                  type="password" 
                  class="input-text"
                  placeholder="Create a password" 
                  required 
                />
              </div>
              
              <div class="form-group">
                <label class="form-label">Confirm Password</label>
                <input 
                  v-model="signupForm.confirmPassword" 
                  type="password" 
                  class="input-text"
                  placeholder="Confirm password" 
                  required 
                />
              </div>
              
              <button type="submit" class="btn btn-primary btn-block" :disabled="loading">
                {{ loading ? 'Creating account...' : 'Create Account' }}
              </button>
            </form>
            
            <div class="form-footer">
              <span>Already a member?</span>
              <a @click="isLogin = true" class="switch-link">Sign in instead</a>
            </div>
          </div>

          <!-- Message Toast -->
          <transition name="fade">
            <div v-if="message" :class="['message-toast', messageType]">
              {{ message }}
            </div>
          </transition>
        </div>
      </section>
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
      localStorage.setItem('token', res.data.accessToken)
      localStorage.setItem('userId', res.data.userId)
      localStorage.setItem('userEmail', loginForm.value.email)
      router.push('/posts')
    } else {
      showMessage(res.msg || 'Login failed', 'error')
    }
  } catch (e) {
    showMessage('Network error, please try again', 'error')
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
      showMessage('Account created successfully. Please log in.', 'success')
      isLogin.value = true
      signupForm.value = { email: '', password: '', confirmPassword: '' }
    } else {
      showMessage(res.msg || 'Signup failed', 'error')
    }
  } catch (e) {
    showMessage('Network error, please try again', 'error')
  }
  loading.value = false
}

function showMessage(msg, type) {
  message.value = msg
  messageType.value = type
  setTimeout(() => { message.value = '' }, 3000)
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  background: var(--color-bg-secondary);
}

.auth-split-layout {
  display: flex;
  min-height: 100vh;
}

/* ─── Brand Section (Left) ─────────────────────────────────────────────────── */
.brand-section {
  flex: 1;
  background: var(--color-bg-primary);
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: var(--space-3xl);
  position: relative;
  overflow: hidden;
  border-right: 1px solid var(--color-border);
}

.brand-section::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, var(--color-accent-primary), var(--color-accent-blue), var(--color-accent-green));
}

.brand-content {
  max-width: 480px;
  margin: 0 auto;
}

.brand-icon {
  width: 48px;
  height: 48px;
  background: var(--color-text-primary);
  color: var(--color-bg-primary);
  font-family: var(--font-display);
  font-size: 1.5rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  margin-bottom: var(--space-xl);
}

.brand-title {
  font-size: 4rem;
  font-weight: 300;
  line-height: 1.1;
  margin-bottom: var(--space-md);
  color: var(--color-text-primary);
}

.brand-tagline {
  font-family: var(--font-body);
  font-size: 1.5rem;
  color: var(--color-text-secondary);
  font-weight: 400;
}

.brand-footer {
  position: absolute;
  bottom: var(--space-xl);
  left: var(--space-3xl);
  font-size: 0.875rem;
  color: var(--color-text-muted);
}

/* ─── Form Section (Right) ─────────────────────────────────────────────────── */
.form-section {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-2xl);
  background: var(--color-bg-elevated);
}

.form-container {
  width: 100%;
  max-width: 400px;
}

.form-title {
  font-size: 2rem;
  margin-bottom: var(--space-xs);
}

.form-subtitle {
  color: var(--color-text-secondary);
  margin-bottom: var(--space-2xl);
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-lg);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.form-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-primary);
}

.btn-block {
  width: 100%;
  margin-top: var(--space-md);
  padding: 1rem;
}

.form-footer {
  margin-top: var(--space-xl);
  text-align: center;
  font-size: 0.95rem;
  color: var(--color-text-secondary);
}

.switch-link {
  color: var(--color-accent-primary);
  font-weight: 500;
  margin-left: var(--space-xs);
  cursor: pointer;
  text-decoration: underline;
  text-decoration-color: transparent;
  transition: all 0.2s;
}

.switch-link:hover {
  color: var(--color-accent-hover);
  text-decoration-color: currentColor;
}

/* ─── Toast ────────────────────────────────────────────────────────────────── */
.message-toast {
  padding: var(--space-md);
  margin-top: var(--space-lg);
  border-radius: var(--radius-sm);
  font-size: 0.9rem;
  text-align: center;
}

.message-toast.error {
  background: var(--color-error-light);
  color: var(--color-error);
}

.message-toast.success {
  background: var(--color-success-light);
  color: var(--color-success);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* ─── Responsive ───────────────────────────────────────────────────────────── */
@media (max-width: 1024px) {
  .brand-title {
    font-size: 3rem;
  }
}

@media (max-width: 768px) {
  .auth-split-layout {
    flex-direction: column;
  }

  .brand-section {
    padding: var(--space-xl);
    flex: 0 0 auto;
    border-right: none;
    border-bottom: 1px solid var(--color-border);
  }

  .brand-icon {
    display: none;
  }

  .brand-title {
    font-size: 2rem;
    margin-bottom: var(--space-xs);
  }

  .brand-tagline {
    font-size: 1.1rem;
  }

  .brand-footer {
    display: none;
  }

  .form-section {
    padding: var(--space-xl);
    align-items: flex-start;
  }
}
</style>
