const { statusText, statusColor, statusBg } = require('../../utils/util')

Component({
  externalClasses: ['custom-class'],
  properties: {
    status: { type: String, value: '' },
    dot: { type: Boolean, value: true },
  },
  data: {
    text: '',
    color: '#9a99a2',
    bg: '#f4f3f4',
  },
  observers: {
    status(v) {
      this.setData({ text: statusText(v), color: statusColor(v), bg: statusBg(v) })
    },
  },
})
