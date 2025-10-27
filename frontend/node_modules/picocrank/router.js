import { createRouter, createWebHistory } from 'vue-router';

import { HomeIcon } from '@hugeicons/core-free-icons';
import { TableIcon } from '@hugeicons/core-free-icons';
import { ViewIcon } from '@hugeicons/core-free-icons';
import { SecurityValidationIcon } from '@hugeicons/core-free-icons';
import { CalendarIcon } from '@hugeicons/core-free-icons';

const routes = [
  {
    name: 'Welcome',
    path: '/',
    component: () => import('./vue/views/Welcome.vue'),
    meta: {
      title: 'Welcome',
      icon: HomeIcon,
    }
  },
  {
    name: 'TableExample',
    path: '/table-example',
    title: 'Table Example',
    component: () => import('./vue/views/TableExample.vue'),
    meta: {
      title: 'Table Example',
      icon: TableIcon,
    }
  },
  {
    name: 'ViewItem',
    path: '/view-item/:id',
    component: () => import('./vue/views/ViewItem.vue'),
    props: true,
    meta: {
      title: 'View Item',
      icon: ViewIcon,
    }
  },
  {
    name: 'Admin',
    path: '/admin',
    component: () => import('./vue/views/Admin.vue'),
    meta: {
      title: 'Admin',
      icon: SecurityValidationIcon,
    }
  },
  {
    name: 'CalendarExample',
    path: '/calendar-example',
    component: () => import('./vue/views/CalendarExample.vue'),
    props: true,
    meta: {
      title: 'Calendar Example',
      icon: CalendarIcon,
    }
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router
