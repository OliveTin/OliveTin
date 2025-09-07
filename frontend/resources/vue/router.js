import { createRouter, createWebHistory } from 'vue-router'

import { Wrench01Icon } from '@hugeicons/core-free-icons'
import { LeftToRightListDashIcon } from '@hugeicons/core-free-icons'
import { CellsIcon } from '@hugeicons/core-free-icons'
import { DashboardSquare01Icon } from '@hugeicons/core-free-icons'

const routes = [
  {
    path: '/',
    name: 'Actions',
    component: () => import('./Dashboard.vue'),
    props: { title: 'default' },
    meta: { title: 'Actions', icon: DashboardSquare01Icon }
  },
  {
    path: '/dashboards/:title',
    name: 'Dashboard',
    component: () => import('./Dashboard.vue'),
    props: true,
    meta: { title: 'OliveTin - Dashboard' }
  },
  {
    path: '/actionBinding/:bindingId/argumentForm',
    name: 'ActionBinding',
    component: () => import('./views/ArgumentForm.vue'),
    props: true,
    meta: { title: 'OliveTin - Action Binding' }
  },
  {
    path: '/logs',
    name: 'Logs',
    component: () => import('./views/LogsListView.vue'),
    meta: { 
      title: 'OliveTin - Logs',
      icon: LeftToRightListDashIcon
    }
  },
  {
    path: '/entities',
    name: 'Entities',
    component: () => import('./views/EntitiesView.vue'),
    meta: { 
      title: 'OliveTin - Entities',
      icon: CellsIcon
    }
  },
  {
    path: '/entity-details/:entityType/:entityKey',
    name: 'EntityDetails',
    component: () => import('./views/EntityDetailsView.vue'),
    props: true,
    meta: { 
      title: 'OliveTin - Entity Details', 
      breadcrumb: [
        { name: "Entities", href: "/entities" },
        { name: "Entity Details" }
      ]
    }
  },
  {
    path: '/logs/:executionTrackingId',
    name: 'Execution',
    component: () => import('./views/ExecutionView.vue'),
    props: true,
    meta: { 
      title: 'OliveTin - Execution', 
      breadcrumb: [
        { name: "Logs", href: "/logs" },
        { name: "Execution" },
      ]
    }
  },
  {
    path: '/diagnostics',
    name: 'Diagnostics',
    component: () => import('./views/DiagnosticsView.vue'),
    meta: { 
      title: 'OliveTin - Diagnostics',
      icon: Wrench01Icon
    }
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
