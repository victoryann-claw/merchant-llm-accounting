const app = getApp()
const { getRecords, deleteRecord } = require('../../api/api')

Page({
  data: {
    merchantId: '',
    records: [],
    loading: false,
    query: '',
    dateRange: ['2026-03-01', '2026-03-23'],
    filterType: ''
  },

  onLoad() {
    const merchantId = app.globalData.merchantId
    if (merchantId) {
      this.setData({ merchantId })
      this.loadRecords()
    }
  },

  onShow() {
    if (this.data.merchantId) {
      this.loadRecords()
    }
  },

  async loadRecords() {
    const { merchantId, query, filterType, dateRange } = this.data
    if (!merchantId) return

    this.setData({ loading: true })
    try {
      const params = { merchant_id: merchantId }
      if (query) params.q = query
      if (filterType) params.type = filterType
      if (dateRange) {
        params.start = dateRange[0]
        params.end = dateRange[1]
      }

      const res = await getRecords(params)
      this.setData({
        records: res.records || [],
        loading: false
      })
    } catch (err) {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },

  onSearch(e) {
    this.setData({ query: e.detail.value })
    this.loadRecords()
  },

  onDateChange(e) {
    const values = e.detail.value
    const dateRange = values.split(' ~ ')
    this.setData({ dateRange })
    this.loadRecords()
  },

  onTypeChange(e) {
    const types = ['', 'purchase', 'delivery', 'payment']
    this.setData({ filterType: types[e.detail.value] })
    this.loadRecords()
  },

  async onDelete(e) {
    const { id } = e.currentTarget.dataset
    wx.showModal({
      title: '确认删除',
      content: '删除后无法恢复',
      success: async (res) => {
        if (res.confirm) {
          try {
            await deleteRecord(id)
            wx.showToast({ title: '删除成功', icon: 'success' })
            this.loadRecords()
          } catch (err) {
            wx.showToast({ title: '删除失败', icon: 'none' })
          }
        }
      }
    })
  }
})
