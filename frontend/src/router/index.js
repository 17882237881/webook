import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import ProfileView from '../views/ProfileView.vue'

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: '/',
            name: 'login',
            component: LoginView
        },
        {
            path: '/profile',
            name: 'profile',
            component: ProfileView,
            meta: { requiresAuth: true }
        }
    ]
})

// 路由守卫
router.beforeEach((to, from, next) => {
    const userId = localStorage.getItem('userId')
    if (to.meta.requiresAuth && !userId) {
        next('/')
    } else {
        next()
    }
})

export default router
