import { http } from '@/api/http'
import type { Category, Product, Sku } from '@/types'

export function listCategories() {
  return http.get<Category[]>('/admin/categories')
}

export function createCategory(name: string) {
  return http.post('/admin/categories', { name, sort: 0 })
}

export function updateCategory(id: number, name: string, sort: number) {
  return http.put(`/admin/categories/${id}`, { name, sort })
}

export function deleteCategory(id: number) {
  return http.del(`/admin/categories/${id}`)
}

export function listProducts() {
  return http.get<Product[]>('/admin/products')
}

export interface ProductInput {
  category_id: number
  title: string
  subtitle: string
  description: string
  /** JSON 数组字符串 */
  images: string
  original_price_cents: number
  unit: string
  detail: string
  /** JSON 数组字符串 */
  detail_images: string
  service_notes: string
  sort: number
  status?: number
}

export function createProduct(input: ProductInput) {
  return http.post('/admin/products', input)
}

export function updateProduct(id: number, input: ProductInput) {
  return http.put(`/admin/products/${id}`, input)
}

export interface SkuInput {
  product_id: number
  specs: string
  price_cents: number
  stock: number
  code: string
}

export function createSku(input: SkuInput) {
  return http.post('/admin/skus', input)
}

export function updateSku(id: number, input: SkuInput) {
  return http.put(`/admin/skus/${id}`, input)
}

export function deleteSku(id: number) {
  return http.del(`/admin/skus/${id}`)
}

export type { Sku }
