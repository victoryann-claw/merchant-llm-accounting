const app = getApp()
const { getTodayStats } = require('../../api/api')

Page({
  data: {
    merchantId: '',
    stats: null,
    recentRecords: [],
    loading: true
  },

  onLoad() {
    const merchantId = app.globalData.merchantId
    if (!merchantId) {
      // 跳转到录入页创建商户
      wx.showModal({
        title: '提示',
        content: '首次使用请先创建商户',
        showCancel: false,
        success: () => {
          wx.switchTab({ url: '/pages/record/record' })
        }
      })
      return
    }
    this.setData({ merchantId })
    this.loadStats()
  },

  onShow() {
    if (this.data.merchantId) {
      this.loadStats()
    }
  },

  async loadStats() {
    this.setData({ loading: true })
    try {
      const res = await getTodayStats(this.data.merchantId)
      this.setData({
        stats: res,
        recentRecords: res.recent_records || [],
        loading: false
      })
    } catch (err) {
      console.error('加载统计数据失败', err)
      this.setData({ loading: false })
      wx.showToast({
        title: '加载失败',
        icon: 'none'
      })
    }
  },

  goToRecord() {
    wx.switchTab({ url: '/pages/record/record' })
  },

  goToRecords() {
    wx.switchTab({ url: '/pages/records/records' })
  }
})
