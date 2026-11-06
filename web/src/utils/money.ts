/** 金额口径统一在服务端用「分」，前端只在展示与输入时换算 */

export function yuan(cents: number): string {
  return (cents / 100).toFixed(2)
}

export function cents(y: number): number {
  return Math.round(y * 100)
}
