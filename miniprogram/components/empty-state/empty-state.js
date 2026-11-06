Component({
  properties: {
    loading: { type: Boolean, value: false },
    text: { type: String, value: '暂无数据' },
    actionText: { type: String, value: '' },
  },

  methods: {
    onAction() {
      this.triggerEvent('action')
    },
  },
})
