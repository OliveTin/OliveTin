import { createRouter, createWebHistory } from 'vue-router'

// Import components
import App from './App.vue'
import ExecutionDialog from './ExecutionDialog.vue'
import ActionButton from './ActionButton.vue'
import ArgumentForm from './ArgumentForm.vue'

// Define routes
const routes = [
  {
    path: '/',
    name: 'Home',
    component: () => import('./views/DashboardRoot.vue'),
    meta: { title: 'OliveTin - Dashboard' }
  },
  {
    path: '/logs',
    name: 'Logs',
    component: () => import('./views/LogsView.vue'),
    meta: { title: 'OliveTin - Logs' }
  },
  {
    path: '/diagnostics',
    name: 'Diagnostics',
    component: () => import('./views/DiagnosticsView.vue'),
    meta: { title: 'OliveTin - Diagnostics' }
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('./views/LoginView.vue'),
    meta: { title: 'OliveTin - Login' }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('./views/NotFoundView.vue'),
    meta: { title: 'OliveTin - Page Not Found' }
  }
]

// Create router instance
const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    } else {
      return { top: 0 }
    }
  }
})

// Navigation guard to update page title
router.beforeEach((to, from, next) => {
  if (to.meta && to.meta.title) {
    document.title = to.meta.title
  }
  next()
})

// Navigation guard for authentication (if needed)
router.beforeEach((to, from, next) => {
  // Check if user is authenticated for protected routes
  const isAuthenticated = window.isAuthenticated || true // Default to true for now
  
  if (to.meta.requiresAuth && !isAuthenticated) {
    next('/login')
  } else {
    next()
  }
})

export default router 