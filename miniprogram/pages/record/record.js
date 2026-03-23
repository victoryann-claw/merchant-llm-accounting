const app = getApp()
const { createMerchant, createRecord } = require('../../api/api')

Page({
  data: {
    merchantId: '',
    inputText: '',
    parsedResult: null,
    loading: false,
    recording: false
  },

  onLoad() {
    const merchantId = app.globalData.merchantId
    this.setData({ merchantId })
  },

  onInput(e) {
    this.setData({ inputText: e.detail.value })
  },

  // 开始录音
  startRecord() {
    this.setData({ recording: true })
    wx.startRecord({
      success: (res) => {
        this.setData({ recording: false })
        // 录音完成后，调用语音识别（这里简化处理，实际需要接入语音识别API）
        wx.showToast({
          title: '录音完成',
          icon: 'success'
        })
      },
      fail: (err) => {
        this.setData({ recording: false })
        wx.showToast({
          title: '录音失败',
          icon: 'none'
        })
      }
    })
  },

  // 停止录音
  stopRecord() {
    wx.stopRecord()
  },

  // 提交记录
  async submitRecord() {
    const { inputText, merchantId } = this.data
    if (!inputText.trim()) {
      wx.showToast({ title: '请输入内容', icon: 'none' })
      return
    }

    // 如果没有商户ID，先创建
    let finalMerchantId = merchantId
    if (!finalMerchantId) {
      try {
        const merchantRes = await createMerchant({
          name: '我的店铺',
          business_type: 'fish'
        })
        finalMerchantId = merchantRes.id
        app.setMerchantId(finalMerchantId)
        this.setData({ merchantId: finalMerchantId })
      } catch (err) {
        wx.showToast({ title: '创建商户失败', icon: 'none' })
        return
      }
    }

    this.setData({ loading: true })
    try {
      const res = await createRecord({
        merchant_id: finalMerchantId,
        user_input: inputText
      })
      this.setData({ 
        loading: false,
        parsedResult: res,
        inputText: ''
      })
      wx.showModal({
        title: '保存成功',
        content: `已识别为：${res.record_type === 'delivery' ? '送货' : res.record_type === 'purchase' ? '进货' : '付款'}，对方：${res.counterparty || '无'}`,
        showCancel: false,
        success: () => {
          // 返回首页刷新
          wx.switchTab({ url: '/pages/index/index' })
        }
      })
    } catch (err) {
      this.setData({ loading: false })
      wx.showToast({ title: '提交失败', icon: 'none' })
    }
  },

  // 重置
  reset() {
    this.setData({
      inputText: '',
      parsedResult: null
    })
  }
})
