export interface CommonResponse<T = unknown> {
  code: number
  message: string
  data: T
}

export interface ListResponse<T> {
  items: T[]
  total: number
}
