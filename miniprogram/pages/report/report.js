const app = getApp()
const { getDailyReport, getPeriodicReport } = require('../../api/api')

Page({
  data: {
    merchantId: '',
    reportType: 'daily',
    dailyDate: '2026-03-23',
    periodicType: 'weekly',
    periodicDate: '2026-03-23',
    report: null,
    loading: false
  },

  onLoad() {
    const merchantId = app.globalData.merchantId
    if (merchantId) {
      this.setData({ merchantId })
      this.loadReport()
    }
  },

  onShow() {
    if (this.data.merchantId) {
      this.loadReport()
    }
  },

  onTabChange(e) {
    const types = ['daily', 'weekly', 'monthly']
    this.setData({ reportType: types[e.detail.value] })
    this.loadReport()
  },

  onDailyDateChange(e) {
    this.setData({ dailyDate: e.detail.value })
    this.loadReport()
  },

  onPeriodicTypeChange(e) {
    const types = ['weekly', 'monthly']
    this.setData({ periodicType: types[e.detail.value] })
    this.loadReport()
  },

  async loadReport() {
    const { merchantId, reportType } = this.data
    if (!merchantId) return

    this.setData({ loading: true })
    try {
      let res
      if (reportType === 'daily') {
        res = await getDailyReport({
          merchant_id: merchantId,
          date: this.data.dailyDate
        })
      } else {
        res = await getPeriodicReport({
          merchant_id: merchantId,
          type: this.data.periodicType,
          date: this.data.periodicDate
        })
      }
      this.setData({ report: res, loading: false })
    } catch (err) {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  }
})
