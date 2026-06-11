const icons = ['点', '邀', '我']

Component({
  data: {
    icons
  },

  properties: {
    current: {
      type: Number,
      value: 0
    },
    list: {
      type: Array,
      value: [
        { text: '点餐', key: 'home' },
        { text: '邀请', key: 'invite' },
        { text: '我的', key: 'profile' }
      ]
    }
  },

  methods: {
    switchTab(e) {
      const index = e.currentTarget.dataset.index
      if (index === this.data.current) return
      this.triggerEvent('change', { index })
    }
  }
})