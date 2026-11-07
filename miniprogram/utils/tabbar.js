// tab 页统一在这里登记选中项，避免每页各写一遍 getTabBar 判空
function setTab(page, index) {
  if (typeof page.getTabBar === 'function' && page.getTabBar()) {
    page.getTabBar().setData({ selected: index })
  }
}

module.exports = { setTab }
