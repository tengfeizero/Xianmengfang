const { icon } = require('../../utils/icons')

Component({
  properties: {
    name: { type: String, value: '' },
    color: { type: String, value: '#1a1a1d' },
    size: { type: Number, value: 40 }, // rpx
    fill: { type: Boolean, value: false },
    sw: { type: Number, value: 1.8 },
  },
  data: { src: '' },
  observers: {
    'name, color, fill, sw': function (name, color, fill, sw) {
      this.setData({ src: name ? icon(name, color, { fill, sw }) : '' })
    },
  },
})
