/** Author: go-blog */

import type { RouteRecordRaw } from 'vue-router'

/**
 * 内置基础路由。
 *
 * 登录页路径用 /login，避免与后端 /auth/* API（Vite proxy）冲突。
 */
export const routes: RouteRecordRaw[] = [
  {
    // 根路径只作为入口，由路由守卫重定向到 VITE_HOME_PATH。
    path: '/',
    name: 'root',
    children: [],
  },
  {
    // 显式 404 页面，用于守卫初始化失败或主动跳转。
    path: '/not-found',
    name: 'not-found',
    component: () => import('@/views/error/NotFound.vue'),
  },
  {
    path: '/login',
    name: 'auth-login',
    component: () => import('@/views/auth/Login.vue'),
    meta: { name: '登录' },
  },
  {
    path: '/auth/forgot-password',
    name: 'auth-forgot-password',
    component: () => import('@/views/auth/ForgotPassword.vue'),
    meta: { name: '找回密码' },
  },
  {
    // 兼容旧书签
    path: '/auth/login',
    redirect: '/login',
  },
  {
    path: '/forgot-password',
    redirect: '/auth/forgot-password',
  },
  {
    // 兜底路由必须放在最后。动态路由注册完成后，守卫会重新匹配原始目标地址。
    path: '/:pathMatch(.*)*',
    name: 'not-found-catch',
    component: () => import('@/views/error/NotFound.vue'),
  },
]
