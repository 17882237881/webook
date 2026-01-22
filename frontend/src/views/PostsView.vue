<template>
  <div class="posts-page">
    <!-- Top Navigation -->
    <header class="page-header stagger-1">
      <div class="header-left">
        <div class="brand-mark">W</div>
        <h1 class="page-title">Editorial</h1>
      </div>
      <div class="header-right">
        <span class="user-greeting">Hello, {{ userEmail }}</span>
        <div class="header-actions">
          <button class="btn btn-secondary btn-sm" @click="goToProfile">Profile</button>
          <button class="btn btn-secondary btn-sm" @click="logout">Sign Out</button>
        </div>
      </div>
    </header>

    <div class="main-layout">
      <!-- Sidebar / Post List -->
      <aside class="sidebar stagger-2">
        <div class="sidebar-header">
          <h3>Your Stories</h3>
          <button class="btn btn-primary btn-sm" @click="createNewPost">New Story</button>
        </div>
        
        <div class="post-list">
          <div v-if="myPosts.length === 0" class="empty-state-small">
            No stories yet.
          </div>
          <div 
            v-for="post in myPosts" 
            :key="post.id" 
            class="post-item"
            :class="{ active: currentPost?.id === post.id }"
            @click="selectPost(post)"
          >
            <div class="post-item-status" :class="post.status === 1 ? 'published' : 'draft'"></div>
            <div class="post-item-content">
              <h4 class="post-item-title">{{ post.title || 'Untitled Story' }}</h4>
              <span class="post-item-meta">{{ post.status === 1 ? 'Published' : 'Draft' }} · {{ formatTime(post.utime) }}</span>
            </div>
          </div>
        </div>
      </aside>

      <!-- Editor Area -->
      <main class="editor-area stagger-3">
        <div v-if="currentPost" class="editor-container">
          <div class="editor-header">
            <div class="editor-status">
              <span class="status-dot" :class="currentPost.status === 1 ? 'published' : 'draft'"></span>
              {{ currentPost.status === 1 ? 'Published' : 'Draft' }}
            </div>
            
            <div class="editor-actions">
               <span v-if="message" :class="['action-message', messageType]">{{ message }}</span>
              <button class="btn btn-secondary" @click="saveDraft" :disabled="saving">
                {{ saving ? 'Saving...' : 'Save Draft' }}
              </button>
              <button class="btn btn-primary" @click="publish" :disabled="publishing">
                {{ publishing ? 'Publishing...' : 'Publish' }}
              </button>
              <button class="btn btn-secondary text-error" @click="deleteCurrentPost" v-if="currentPost.id">
                Delete
              </button>
            </div>
          </div>

          <input 
            v-model="currentPost.title" 
            class="editor-title" 
            placeholder="Title..."
          />
          
          <textarea 
            v-model="currentPost.content" 
            class="editor-content" 
            placeholder="Tell your story..."
          ></textarea>
        </div>
        
        <div v-else class="empty-state">
          <div class="empty-icon">✎</div>
          <h3>Ready to write?</h3>
          <p>Select a story from the sidebar or start a new one.</p>
          <button class="btn btn-primary" @click="createNewPost">Create New Story</button>
        </div>
      </main>

      <!-- Published Feed (Right Sidebar/Drawer style for this layout) -->
      <aside class="published-feed stagger-4">
        <h3 class="feed-title">Latest on Webook</h3>
        <div class="feed-list">
           <div v-if="publishedPosts.length === 0" class="empty-state-small">
            No published stories yet.
          </div>
          <div v-for="post in publishedPosts" :key="post.id" class="feed-item card">
            <h4 class="feed-item-title">{{ post.title }}</h4>
            <p class="feed-item-excerpt">{{ post.content.substring(0, 80) }}...</p>
            <div class="feed-item-meta">
              <span>By User {{ post.authorId }}</span>
            </div>
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { savePost, publishPost, getMyPosts, getPublishedPosts, deletePost, getDraft } from '../api/post.js'

const router = useRouter()

const userEmail = ref(localStorage.getItem('userEmail') || 'User')
const myPosts = ref([])
const publishedPosts = ref([])
const currentPost = ref(null)
const saving = ref(false)
const publishing = ref(false)
const message = ref('')
const messageType = ref('success')

// Load data
onMounted(async () => {
  await loadMyPosts()
  await loadPublishedPosts()
})

async function loadMyPosts() {
  try {
    const res = await getMyPosts()
    if (res.code === 0) {
      myPosts.value = res.data.posts || []
    }
  } catch (e) {
    console.error('Failed to load posts', e)
  }
}

async function loadPublishedPosts() {
  try {
    const res = await getPublishedPosts()
    if (res.code === 0) {
      publishedPosts.value = res.data.posts || []
    }
  } catch (e) {
    console.error('Failed to load published posts', e)
  }
}

function createNewPost() {
  currentPost.value = { id: 0, title: '', content: '', status: 0 }
  message.value = ''
}

async function selectPost(post) {
  try {
    const res = await getDraft(post.id)
    if (res.code === 0) {
      currentPost.value = res.data
    } else {
      currentPost.value = { ...post }
    }
    message.value = ''
  } catch (e) {
    currentPost.value = { ...post }
  }
}

async function saveDraft() {
  if (!currentPost.value.title) {
    showMessage('Title required', 'error')
    return
  }
  saving.value = true
  try {
    const res = await savePost(currentPost.value.id, currentPost.value.title, currentPost.value.content)
    if (res.code === 0) {
      currentPost.value.id = res.data.id
      showMessage('Draft saved', 'success')
      await loadMyPosts()
    } else {
      showMessage(res.msg || 'Save failed', 'error')
    }
  } catch (e) {
    showMessage('Save failed', 'error')
  }
  saving.value = false
}

async function publish() {
  if (!currentPost.value.title) {
    showMessage('Title required', 'error')
    return
  }
  publishing.value = true
  try {
    const res = await publishPost(currentPost.value.id, currentPost.value.title, currentPost.value.content)
    if (res.code === 0) {
      currentPost.value.id = res.data.id
      currentPost.value.status = 1
      showMessage('Published!', 'success')
      await loadMyPosts()
      await loadPublishedPosts()
    } else {
      showMessage(res.msg || 'Publish failed', 'error')
    }
  } catch (e) {
    showMessage('Publish failed', 'error')
  }
  publishing.value = false
}

async function deleteCurrentPost() {
  if (!confirm('Are you sure you want to delete this story?')) return
  try {
    const res = await deletePost(currentPost.value.id)
    if (res.code === 0) {
      showMessage('Deleted', 'success')
      currentPost.value = null
      await loadMyPosts()
      await loadPublishedPosts()
    } else {
      showMessage(res.msg || 'Delete failed', 'error')
    }
  } catch (e) {
    showMessage('Delete failed', 'error')
  }
}

function showMessage(msg, type) {
  message.value = msg
  messageType.value = type
  setTimeout(() => { message.value = '' }, 3000)
}

function formatTime(timestamp) {
  if (!timestamp) return ''
  return new Date(timestamp).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

function goToProfile() {
  router.push('/profile')
}

function logout() {
  localStorage.removeItem('token')
  localStorage.removeItem('userId')
  localStorage.removeItem('userEmail')
  router.push('/')
}
</script>

<style scoped>
.posts-page {
  min-height: 100vh;
  background: var(--color-bg-primary);
  display: flex;
  flex-direction: column;
}

/* ─── Header ───────────────────────────────────────────────────────────────── */
.page-header {
  height: 64px;
  padding: 0 var(--space-xl);
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--color-border);
  background: var(--color-bg-elevated);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-md);
}

.brand-mark {
  width: 32px;
  height: 32px;
  background: var(--color-text-primary);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: var(--font-display);
  font-weight: 600;
  border-radius: 4px;
}

.page-title {
  font-size: 1.25rem;
  margin: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--space-lg);
}

.user-greeting {
  font-size: 0.9rem;
  color: var(--color-text-secondary);
}

.header-actions {
  display: flex;
  gap: var(--space-sm);
}

/* ─── Main Layout ──────────────────────────────────────────────────────────── */
.main-layout {
  display: flex;
  flex: 1;
  height: calc(100vh - 64px);
}

/* ─── Sidebar ──────────────────────────────────────────────────────────────── */
.sidebar {
  width: 280px;
  border-right: 1px solid var(--color-border);
  background: var(--color-bg-secondary);
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: var(--space-md) var(--space-lg);
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--color-border-light);
}

.sidebar-header h3 {
  font-size: 0.95rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-muted);
  margin: 0;
}

.post-list {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-md);
}

.post-item {
  padding: var(--space-md);
  margin-bottom: var(--space-xs);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid transparent;
  display: flex;
  gap: var(--space-sm);
}

.post-item:hover {
  background: var(--color-bg-tertiary);
}

.post-item.active {
  background: var(--color-bg-elevated);
  border-color: var(--color-border);
  box-shadow: var(--shadow-sm);
}

.post-item-status {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  margin-top: 8px;
  flex-shrink: 0;
  background: var(--color-text-muted);
}

.post-item-status.published {
  background: var(--color-success);
}

.post-item-status.draft {
  background: var(--color-warning);
}

.post-item-title {
  font-size: 0.95rem;
  font-weight: 500;
  margin-bottom: 4px;
  color: var(--color-text-primary);
  line-height: 1.4;
}

.post-item-meta {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  font-family: var(--font-body);
}

/* ─── Editor Area ──────────────────────────────────────────────────────────── */
.editor-area {
  flex: 1;
  min-width: 0; /* Prevent flex overflow */
  background: var(--color-bg-primary);
  position: relative;
  overflow-y: auto;
}

.editor-container {
  max-width: 740px;
  margin: 0 auto;
  padding: var(--space-2xl) var(--space-lg);
  min-height: 100%;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-xl);
}

.editor-status {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-weight: 500;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-dot.published { background: var(--color-success); }
.status-dot.draft { background: var(--color-warning); }

.editor-actions {
  display: flex;
  gap: var(--space-sm);
  align-items: center;
}

.editor-title {
  width: 100%;
  font-family: var(--font-display);
  font-size: 2.5rem;
  font-weight: 600;
  border: none;
  background: transparent;
  padding: 0;
  margin-bottom: var(--space-lg);
  color: var(--color-text-primary);
  line-height: 1.2;
}

.editor-content {
  width: 100%;
  font-family: var(--font-body);
  font-size: 1.15rem;
  line-height: 1.8;
  border: none;
  background: transparent;
  color: var(--color-text-secondary);
  resize: none;
  min-height: 60vh;
}

.editor-title::placeholder,
.editor-content::placeholder {
  color: var(--color-text-muted);
}

/* ─── Empty State ──────────────────────────────────────────────────────────── */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  text-align: center;
  color: var(--color-text-secondary);
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: var(--space-md);
  opacity: 0.3;
}

.empty-state h3 {
  margin-bottom: var(--space-xs);
}

.empty-state p {
  color: var(--color-text-muted);
  margin-bottom: var(--space-lg);
}

.empty-state-small {
  padding: var(--space-lg);
  text-align: center;
  color: var(--color-text-muted);
  font-size: 0.9rem;
  font-style: italic;
}

/* ─── Published Feed ───────────────────────────────────────────────────────── */
.published-feed {
  width: 320px;
  border-left: 1px solid var(--color-border);
  background: var(--color-bg-primary);
  padding: var(--space-lg);
  overflow-y: auto;
}

.feed-title {
  font-size: 1rem;
  margin-bottom: var(--space-lg);
  padding-bottom: var(--space-sm);
  border-bottom: 2px solid var(--color-accent-primary);
  display: inline-block;
}

.feed-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.feed-item {
  padding: var(--space-md);
}

.feed-item-title {
  font-size: 1.1rem;
  margin-bottom: var(--space-xs);
  line-height: 1.3;
}

.feed-item-excerpt {
  font-size: 0.9rem;
  color: var(--color-text-secondary);
  margin-bottom: var(--space-sm);
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.feed-item-meta {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.text-error {
  color: var(--color-error);
}

.action-message {
  font-size: 0.85rem;
  margin-right: var(--space-sm);
}

.action-message.success { color: var(--color-success); }
.action-message.error { color: var(--color-error); }

@media (max-width: 1200px) {
  .published-feed {
    display: none;
  }
}

@media (max-width: 768px) {
  .sidebar {
    display: none;
  }
}
</style>
