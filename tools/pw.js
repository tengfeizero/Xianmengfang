// playwright-core 解析器：让截图脚本在不同机器上都能跑。
// 查找顺序：环境变量 PLAYWRIGHT_CORE → 本地 npm 依赖 → 系统全局 → 本机 WorkBuddy 内置副本。
// 用法：const { chromium } = require('./pw')
const path = require('path')
const os = require('os')

function candidates() {
  const list = [
    process.env.PLAYWRIGHT_CORE,
    'playwright-core',
    'playwright',
    // 兜底：WorkBuddy 桌面端自带的 playwright-core（仅本机存在时才命中）
    path.join(
      os.homedir(),
      '.workbuddy/binaries/node/versions',
      '22.22.2-6/node_modules/@playwright/cli/node_modules/playwright-core',
    ),
  ]
  return list.filter(Boolean)
}

for (const c of candidates()) {
  try {
    module.exports = require(c)
    return
  } catch {
    /* 试下一个 */
  }
}

console.error(
  '[pw] 未找到 playwright-core。请在项目根执行：npm i -D playwright-core\n' +
    '     或设置环境变量 PLAYWRIGHT_CORE=/绝对路径/node_modules/playwright-core',
)
process.exit(1)
