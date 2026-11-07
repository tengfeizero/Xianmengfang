const { statusText, statusColor, statusBg } = require('../../utils/util')

Component({
  externalClasses: ['custom-class'],
  properties: {
    status: { type: String, value: '' },
    dot: { type: Boolean, value: true },
  },
  data: {
    text: '',
    color: '#9a93a8',
    bg: '#f3f1f7',
  },
  observers: {
    status(v) {
      this.setData({ text: statusText(v), color: statusColor(v), bg: statusBg(v) })
    },
  },
})
