/** Author: go-blog */

import { shallowRef } from 'vue'

const DICT_TREE_STORAGE_KEY = 'mypage:admin:dict-tree'

export const dictTreeState = shallowRef<any[]>([])

export function syncDictTree() {
  dictTreeState.value = readStoredDictTree()
  return dictTreeState.value
}

export function isDictLoaded() {
  return dictTreeState.value.length > 0 || syncDictTree().length > 0
}

export async function refreshDict() {
  // mypage 无字典服务
  return Promise.resolve()
}

export function clearDict() {
  dictTreeState.value = []
  localStorage.removeItem(DICT_TREE_STORAGE_KEY)
}

/** 无字典服务时回退内置项（资源管理） */
const BUILTIN_DICTS: any[] = [
  {
    code: 'COMMON_STATUS',
    children: [
      { label: '启用', value: 'ENABLED', status: 'ENABLED', color: '#18a058' },
      { label: '停用', value: 'DISABLED', status: 'ENABLED', color: '#d03050' },
    ],
  },
  {
    code: 'RESOURCE_TYPE',
    children: [
      { label: '目录', value: 'CATALOG', status: 'ENABLED' },
      { label: '菜单', value: 'MENU', status: 'ENABLED' },
      { label: '页面', value: 'PAGE', status: 'ENABLED' },
      { label: '按钮', value: 'BUTTON', status: 'ENABLED' },
    ],
  },
]

export function dictDataAll() {
  return dictTreeState.value.length ? dictTreeState.value : BUILTIN_DICTS
}

export function dictTypeList(dictCode: string, tree = dictDataAll()) {
  return findDictRoot(tree, dictCode)?.children ?? []
}

export function dictList(dictCode: any, tree = dictDataAll()) {
  return dictTypeList(dictCode, tree)
    .filter(isEnabledDict)
    .map((item: any) => ({
      label: getDictLabel(item),
      value: getDictValue(item),
    }))
}

export function dictTypeData(
  dictCode: string,
  value?: string | number | null,
  tree = dictDataAll(),
) {
  const dict = findDictItem(dictCode, value, tree)
  return dict ? getDictLabel(dict) : ''
}

export function dictTypeColor(
  dictCode: string,
  value?: string | number | null,
  tree = dictDataAll(),
) {
  const dict = findDictItem(dictCode, value, tree)
  return dict?.color || ''
}

export function translateDictTree(
  dictCode: string,
  value?: string | number | null,
  tree = dictDataAll(),
) {
  const root = findDictRoot(tree, dictCode)
  const dict = findNodeByValue(root, value)
  return dict ? getDictLabel(dict) : ''
}

export function getDictValue(item: any) {
  return item.value || item.code
}

export function getDictLabel(item: any) {
  return item.label || item.code
}

export function isEnabledDict(item: any) {
  return item.status === undefined || item.status === null || item.status === 'ENABLED'
}

function findDictItem(dictCode: string, value?: string | number | null, tree = dictDataAll()) {
  if (value === undefined || value === null || value === '') {
    return undefined
  }
  const normalizedValue = String(value)
  return dictTypeList(dictCode, tree).find((item: any) => getDictValue(item) === normalizedValue)
}

function findDictRoot(tree: any[], dictCode: string) {
  return tree.find((item) => item.code === dictCode)
}

function findNodeByValue(node: any, value?: string | number | null): any {
  if (!node || value === undefined || value === null || value === '') {
    return undefined
  }

  if (getDictValue(node) === String(value)) {
    return node
  }

  for (const child of node.children ?? []) {
    const result = findNodeByValue(child, value)
    if (result) {
      return result
    }
  }

  return undefined
}

function setDictTree(tree: any[]) {
  dictTreeState.value = Array.isArray(tree) ? tree : []
  localStorage.setItem(DICT_TREE_STORAGE_KEY, JSON.stringify(dictTreeState.value))
}

function readStoredDictTree() {
  const raw = localStorage.getItem(DICT_TREE_STORAGE_KEY)
  if (!raw) {
    return []
  }
  try {
    const tree = JSON.parse(raw)
    return Array.isArray(tree) ? tree : []
  } catch {
    localStorage.removeItem(DICT_TREE_STORAGE_KEY)
    return []
  }
}

syncDictTree()
