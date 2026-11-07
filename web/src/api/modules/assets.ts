import { http, uploadImage } from '@/api/http'
import type { Board, ElementItem } from '@/types'

// ---- 底板：元素按底板归属（一对多），底板页展示每块底板挂了多少元素 ----

export function listBoards() {
  return http.get<Board[]>('/admin/boards')
}

export interface BoardInput {
  name: string
  url: string
  width: number
  height: number
}

export function createBoard(input: BoardInput) {
  return http.post('/admin/boards', input)
}

export function updateBoard(id: number, input: BoardInput) {
  return http.put(`/admin/boards/${id}`, input)
}

/** 下架底板：底板下还有启用元素时后端返回 409 */
export function toggleBoard(id: number) {
  return http.del(`/admin/boards/${id}`)
}

// ---- 元素：按 board_id 取，新增/编辑都必须带 board_id ----

export function listElements(boardId: number, category?: string) {
  const q = new URLSearchParams({ board_id: String(boardId) })
  if (category) q.set('category', category)
  return http.get<ElementItem[]>('/admin/elements?' + q.toString())
}

export interface ElementInput {
  board_id: number
  category: string
  name: string
  image_url: string
  anchor: string
  sort: number
}

export function createElement(input: ElementInput) {
  return http.post('/admin/elements', input)
}

export function updateElement(id: number, input: ElementInput) {
  return http.put(`/admin/elements/${id}`, input)
}

export function toggleElement(id: number) {
  return http.del(`/admin/elements/${id}`)
}

export { uploadImage }
