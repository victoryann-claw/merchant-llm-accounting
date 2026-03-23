const app = getApp()
const { createMerchant, createRecord } = require('../../api/api')

Page({
  data: {
    merchantId: '',
    inputText: '',
    parsedResult: null,
    loading: false,
    recording: false,
    tempFilePath: ''
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
    
    // 开始录音
    wx.startRecord({
      success: (res) => {
        this.setData({ 
          recording: false,
          tempFilePath: res.tempFilePath
        })
        wx.showToast({
          title: '录音完成',
          icon: 'success',
          duration: 1000
        })
      },
      fail: (err) => {
        this.setData({ recording: false })
        wx.showToast({
          title: '录音失败',
          icon: 'none'
        })
        console.error('录音失败', err)
      }
    })
  },

  // 停止录音
  stopRecord() {
    wx.stopRecord()
  },

  // 提交文字记录
  async submitText() {
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
          wx.switchTab({ url: '/pages/index/index' })
        }
      })
    } catch (err) {
      this.setData({ loading: false })
      wx.showToast({ title: '提交失败', icon: 'none' })
    }
  },

  // 提交语音记录
  async submitVoice() {
    const { tempFilePath, merchantId } = this.data
    if (!tempFilePath) {
      wx.showToast({ title: '请先录音', icon: 'none' })
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

    wx.showLoading({ title: '识别中...' })
    
    // 上传音频文件到后端
    wx.uploadFile({
      url: app.globalData.baseUrl + '/record/voice',
      filePath: tempFilePath,
      name: 'audio',
      formData: {
        merchant_id: finalMerchantId
      },
      success: (res) => {
        wx.hideLoading()
        const data = JSON.parse(res.data)
        if (res.statusCode === 201) {
          this.setData({ tempFilePath: '' })
          wx.showModal({
            title: '保存成功',
            content: `语音已识别并保存为：${data.record_type === 'delivery' ? '送货' : data.record_type === 'purchase' ? '进货' : '付款'}`,
            showCancel: false,
            success: () => {
              wx.switchTab({ url: '/pages/index/index' })
            }
          })
        } else {
          wx.showToast({ title: data.error || '提交失败', icon: 'none' })
        }
      },
      fail: (err) => {
        wx.hideLoading()
        wx.showToast({ title: '提交失败', icon: 'none' })
        console.error('上传音频失败', err)
      }
    })
  },

  // 重置
  reset() {
    this.setData({
      inputText: '',
      parsedResult: null,
      tempFilePath: ''
    })
  }
})
