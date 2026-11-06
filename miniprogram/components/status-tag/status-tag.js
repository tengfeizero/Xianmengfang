const { statusText, statusColor } = require('../../utils/util')

Component({
  externalClasses: ['custom-class'],
  properties: {
    status: { type: String, value: '' },
  },
  data: {
    text: '',
    color: '#999999',
  },
  observers: {
    status(v) {
      this.setData({ text: statusText(v), color: statusColor(v) })
    },
  },
})
