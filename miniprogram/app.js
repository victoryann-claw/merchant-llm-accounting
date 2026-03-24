App({
  globalData: {
    userId: null,
    merchantId: null,
    baseUrl: 'http://localhost:8080/api/v1'
  },

  onLaunch() {
    // 检查本地存储
    const userId = wx.getStorageSync('userId')
    const merchantId = wx.getStorageSync('merchantId')
    if (userId) {
      this.globalData.userId = userId
    }
    if (merchantId) {
      this.globalData.merchantId = merchantId
    }
  },

  setUserId(id) {
    this.globalData.userId = id
    wx.setStorageSync('userId', id)
  },

  setMerchantId(id) {
    this.globalData.merchantId = id
    wx.setStorageSync('merchantId', id)
  }
})
