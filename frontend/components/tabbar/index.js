// 自定义底部 tabbar（v6 设计稿：线描图标 房子/加人/人像，激活墨绿 / 未激活灰）。
// 替换 weui mp-tabbar，使图标与目标稿一致。父页通过 current 控制激活项，
// 监听 change 事件（detail.index）做路由，与原 tabChange 契约一致。
Component({
  properties: {
    current: { type: Number, value: 0 },
  },
  data: {
    items: [
      { text: '点餐', cls: 'menu' },
      { text: '邀请', cls: 'invite' },
      { text: '我的', cls: 'profile' },
    ],
  },
  methods: {
    onTap(e) {
      const index = Number(e.currentTarget.dataset.index)
      this.triggerEvent('change', { index })
    },
  },
})
