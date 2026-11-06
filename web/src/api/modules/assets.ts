import { http, uploadImage } from '@/api/http'
import type { Board, ElementItem } from '@/types'

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

export function toggleBoard(id: number) {
  return http.del(`/admin/boards/${id}`)
}

export function listElements(category?: string) {
  return http.get<ElementItem[]>('/admin/elements' + (category ? `?category=${category}` : ''))
}

export interface ElementInput {
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
