<template>
  <div class="profile-page">
    <div class="profile-container stagger-1">
      <header class="page-header">
         <h1 class="page-title">Personal Profile</h1>
         <button class="btn btn-secondary btn-sm" @click="goBack">← Back to Editorial</button>
      </header>

      <div class="profile-card card">
        <div class="profile-header">
          <div class="avatar-placeholder">{{ email ? email[0].toUpperCase() : 'U' }}</div>
          <h2 class="user-email">{{ email }}</h2>
          <span class="user-id">User ID: {{ userId }}</span>
        </div>

        <div class="divider"></div>

        <div class="settings-section">
          <h3>Security Settings</h3>
          
          <form @submit.prevent="handleUpdatePassword" class="settings-form">
             <div class="form-group">
              <label class="form-label">Current Password</label>
              <input 
                v-model="passwordForm.oldPassword" 
                type="password" 
                class="input-text"
                placeholder="Enter current password" 
                required 
              />
            </div>
            
             <div class="form-group">
              <label class="form-label">New Password</label>
              <input 
                v-model="passwordForm.newPassword" 
                type="password" 
                class="input-text"
                placeholder="Enter new password" 
                required 
              />
            </div>

            <div class="form-actions">
              <button type="submit" class="btn btn-primary" :disabled="loading">
                {{ loading ? 'Updating...' : 'Update Password' }}
              </button>
            </div>

            <p v-if="message" :class="['message', messageType]">{{ message }}</p>
          </form>
        </div>

        <div class="divider"></div>

        <button @click="handleLogout" class="btn btn-secondary btn-block text-error">
          Sign Out
        </button>
      </div>
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
      message.value = res.msg || 'Failed to load profile'
    }
  } catch (e) {
    message.value = 'Network error'
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
      message.value = 'Password updated successfully'
      messageType.value = 'success'
      passwordForm.value = { oldPassword: '', newPassword: '' }
    } else {
      message.value = res.msg || 'Update failed'
      messageType.value = 'error'
    }
  } catch (e) {
    message.value = 'Network error'
    messageType.value = 'error'
  }
  loading.value = false
}

function handleLogout() {
  localStorage.removeItem('token')
  localStorage.removeItem('userId')
  router.push('/')
}

function goBack() {
  router.push('/posts')
}
</script>

<style scoped>
.profile-page {
  min-height: 100vh;
  background: var(--color-bg-primary);
  padding: var(--space-2xl) var(--space-md);
  display: flex;
  justify-content: center;
}

.profile-container {
  width: 100%;
  max-width: 500px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-xl);
}

.page-title {
  font-size: 2rem;
  margin: 0;
}

.profile-card {
  padding: var(--space-2xl);
}

.profile-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  margin-bottom: var(--space-xl);
}

.avatar-placeholder {
  width: 80px;
  height: 80px;
  background: var(--color-accent-primary);
  color: white;
  font-family: var(--font-display);
  font-size: 2.5rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  margin-bottom: var(--space-md);
  box-shadow: var(--shadow-md);
}

.user-email {
  font-size: 1.25rem;
  margin-bottom: var(--space-xs);
}

.user-id {
  font-family: var(--font-body);
  font-size: 0.9rem;
  color: var(--color-text-muted);
}

.divider {
  height: 1px;
  background: var(--color-border-light);
  margin: var(--space-xl) 0;
}

.settings-section h3 {
  font-size: 1.1rem;
  margin-bottom: var(--space-lg);
  color: var(--color-text-primary);
}

.settings-form {
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
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--color-text-secondary);
}

.btn-block {
  width: 100%;
}

.text-error {
  color: var(--color-error);
}

.message {
  padding: var(--space-sm);
  border-radius: var(--radius-sm);
  text-align: center;
  font-size: 0.9rem;
}

.message.success {
  background: var(--color-success-light);
  color: var(--color-success);
}

.message.error {
  background: var(--color-error-light);
  color: var(--color-error);
}
</style>
