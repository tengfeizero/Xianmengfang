Component({
  properties: {
    loading: { type: Boolean, value: false },
    text: { type: String, value: '这里还什么都没有' },
    desc: { type: String, value: '' },
    icon: { type: String, value: 'sparkles' },
    actionText: { type: String, value: '' },
  },
  methods: {
    onAction() {
      this.triggerEvent('action')
    },
  },
})
