/** Author: go-blog */

import { defineStore } from 'pinia'
import { router } from '@/router'
import { resourceApi } from '@/api'
import {
  createFullscreenRoutes,
  createMenus,
  createRoutes,
  generateCacheRoutes,
  getActiveMenuPath,
  groupResourcesByModule,
  normalizeRowRoutes,
  resolveHomePath,
} from './route/helper'

export { resolveHomePath }

/**
 * 路由 store 状态。
 */
interface RouteState {
  isInitAuthRoute: boolean
  menus: AppRoute.MenuOption[]
  rowRoutes: AppRoute.RowRoute[]
  resourceModules: AppRoute.ResourceModule[]
  activeModuleId: string | null
  currentMenuPath: string | null
  cacheRoutes: string[]
  fullscreenRouteNames: string[]
}

type RouteGetters = Record<string, never>

interface RouteActions {
  resetRouteStore: () => void
  resetRoutes: () => void
  setCurrentMenuPath: (path: string) => void
  setActiveModule: (moduleId: string | null) => void
  syncActiveModuleByPath: (path: string) => void
  initRouteInfo: () => Promise<AppRoute.RowRoute[]>
  initAuthRoute: () => Promise<void>
}

export const useRouteStore = defineStore<'route-store', RouteState, RouteGetters, RouteActions>(
  'route-store',
  {
    state: (): RouteState => ({
      isInitAuthRoute: false,
      menus: [],
      rowRoutes: [],
      resourceModules: [],
      activeModuleId: null,
      currentMenuPath: null,
      cacheRoutes: [],
      fullscreenRouteNames: [],
    }),
    actions: {
      resetRouteStore() {
        this.resetRoutes()
        this.$reset()
      },

      resetRoutes() {
        if (router.hasRoute('appRoot')) {
          router.removeRoute('appRoot')
        }
        this.fullscreenRouteNames.forEach((routeName) => {
          if (router.hasRoute(routeName)) {
            router.removeRoute(routeName)
          }
        })
        this.fullscreenRouteNames = []
      },

      setCurrentMenuPath(path: string) {
        this.currentMenuPath = getActiveMenuPath(this.rowRoutes, path)
        this.syncActiveModuleByPath(path)
      },

      setActiveModule(moduleId: string | null) {
        const resourceModule = this.resourceModules.find((item) => item.id === moduleId)
        this.activeModuleId = resourceModule?.id ?? null
        this.menus = resourceModule ? createMenus(resourceModule.resources) : []
      },

      async initRouteInfo() {
        const res = await resourceApi.current()
        const rows = Array.isArray(res.data) ? res.data : []
        return normalizeRowRoutes(rows as AppRoute.RowRoute[])
      },

      async initAuthRoute() {
        this.isInitAuthRoute = false

        const rowRoutes = await this.initRouteInfo()
        this.rowRoutes = rowRoutes
        this.resourceModules = groupResourcesByModule(rowRoutes)

        if (router.hasRoute('appRoot')) {
          router.removeRoute('appRoot')
        }

        router.addRoute(createRoutes(rowRoutes))
        const fullscreenRoutes = createFullscreenRoutes(rowRoutes)
        fullscreenRoutes.forEach((route) => {
          router.addRoute(route)
        })
        this.fullscreenRouteNames = fullscreenRoutes
          .map((route) => route.name)
          .filter((name): name is string => typeof name === 'string')
        this.setActiveModule(this.resourceModules[0]?.id ?? null)
        this.cacheRoutes = generateCacheRoutes(rowRoutes)
        this.isInitAuthRoute = true
      },

      syncActiveModuleByPath(path: string) {
        const activePath = this.currentMenuPath ?? path
        const resource =
          this.rowRoutes.find((item) => item.path === path) ??
          this.rowRoutes.find((item) => item.path === activePath)

        if (
          resource?.module_id &&
          resource.module_id !== this.activeModuleId &&
          this.resourceModules.some((item) => item.id === resource.module_id)
        ) {
          this.setActiveModule(resource.module_id)
        }
      },
    },
  },
)

export function getRouteTitle(route: {
  name?: string | symbol | null
  path?: string
  meta: {
    name?: string
  }
}) {
  return route.meta.name ?? String(route.name ?? route.path)
}
