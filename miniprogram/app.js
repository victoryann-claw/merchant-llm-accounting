App({
  globalData: {
    merchantId: null,
    baseUrl: 'http://localhost:8080/api/v1'
  },

  onLaunch() {
    // 检查本地存储的商户ID
    const merchantId = wx.getStorageSync('merchantId')
    if (merchantId) {
      this.globalData.merchantId = merchantId
    }
  },

  setMerchantId(id) {
    this.globalData.merchantId = id
    wx.setStorageSync('merchantId', id)
  }
})
