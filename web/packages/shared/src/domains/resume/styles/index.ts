/**
 * Author: go-blog — 简历纸面样式入口
 *
 * 加皮肤：
 * 1. templates/ 下新增 `{id}.css`（选择器挂在 `.resume-template--{id}`）
 * 2. templates/manifest.json 增加一条（structure 用 sheet-v1，css 指向该文件）
 *
 * 加结构性模板：
 * 1. 后端 resume_html.STRUCTURE_BUILDERS 注册新 structure 与 DOM 构建
 * 2. manifest 增一条，structure 指向新族
 * 3. 非 sheet-v1 的 Admin/Portal 预览走后端 HTML，不必再写前端 DOM
 */
import './sheet.css'

// 自动装载全部皮肤 CSS；新增文件无需再改本入口
import.meta.glob('./templates/*.css', { eager: true })
