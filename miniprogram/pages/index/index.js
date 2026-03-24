const app = getApp()
const { createRecord, createReminder, getRecords } = require('../../api/api')

Page({
  data: {
    merchantId: '',
    messages: [],
    inputText: '',
    inputMode: 'text', // text | voice | image
    voiceRecording: false,
    voiceTempFilePath: '',
    loading: false,
    scrollTop: 0
  },

  onLoad() {
    const merchantId = app.globalData.merchantId
    if (!merchantId) {
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
  },

  onShow() {
    if (!app.globalData.merchantId) return
    this.setData({ merchantId: app.globalData.merchantId })
  },

  // 切换输入模式
  switchMode(e) {
    const mode = e.currentTarget.dataset.mode
    this.setData({ inputMode: mode })
  },

  // 文字输入
  onInputChange(e) {
    this.setData({ inputText: e.detail.value })
  },

  // 发送文字
  async sendText() {
    const text = this.data.inputText.trim()
    if (!text) return
    
    this.setData({ inputText: '' })
    await this.processUserMessage(text, 'text')
  },

  // 切换到语音模式
  switchMode(e) {
    const mode = e.currentTarget.dataset.mode
    this.setData({ inputMode: mode })
  },

  // 开始录音
  startVoice(e) {
    if (this.data.voiceRecording) return
    this.setData({ voiceRecording: true })
    wx.startRecord({
      success: (res) => {
        this.setData({ voiceTempFilePath: res.tempFilePath })
      },
      fail: (err) => {
        console.error('录音失败', err)
        wx.showToast({ title: '录音失败', icon: 'none' })
        this.setData({ voiceRecording: false })
      }
    })
  },

  // 结束录音
  endVoice(e) {
    if (!this.data.voiceRecording) return
    this.setData({ voiceRecording: false })
    wx.stopRecord()
    
    // 发送语音消息
    const voicePath = this.data.voiceTempFilePath
    if (voicePath) {
      this.processUserMessage(voicePath, 'voice')
    }
  },

  // 取消录音
  cancelVoice(e) {
    this.setData({ voiceRecording: false })
    wx.stopRecord()
  },

  // 播放语音
  playVoice(e) {
    const url = e.currentTarget.dataset.url
    wx.playVoice({ filePath: url })
  },

  // 选择图片
  pickImage() {
    wx.chooseImage({
      count: 1,
      sourceType: ['album', 'camera'],
      success: (res) => {
        const tempPath = res.tempFilePaths[0]
        this.processUserMessage(tempPath, 'image')
      }
    })
  },

  // 预览图片
  previewImage(e) {
    wx.previewImage({
      urls: [e.currentTarget.dataset.url]
    })
  },

  // 处理用户消息
  async processUserMessage(content, type) {
    const msgId = Date.now().toString()
    
    // 添加用户消息
    const userMsg = {
      id: msgId,
      role: 'user',
      type: type,
      content: content,
      duration: type === 'voice' ? 5 : 0
    }
    
    const messages = [...this.data.messages, userMsg]
    this.setData({ messages, loading: true, scrollTop: msgId })
    
    try {
      if (type === 'text') {
        // 文字直接调用API
        const res = await createRecord({
          merchant_id: this.data.merchantId,
          user_input: content
        })
        this.addConfirmMessage(res)
      } else if (type === 'image') {
        // 图片需要上传
        const res = await this.uploadImage(content)
        this.addConfirmMessage(res)
      } else if (type === 'voice') {
        // 语音需要上传
        const res = await this.uploadVoice(content)
        this.addConfirmMessage(res)
      }
    } catch (err) {
      console.error('处理消息失败', err)
      wx.showToast({ title: '处理失败：' + err.message, icon: 'none' })
      // 移除刚才添加的用户消息
      this.setData({
        messages: this.data.messages.filter(m => m.id !== msgId),
        loading: false
      })
    }
  },

  // 添加确认消息
  addConfirmMessage(data) {
    const aiMsg = {
      id: (Date.now() + 1).toString(),
      role: 'ai',
      confirmed: false,
      data: {
        id: data.id,
        record_type: data.record_type,
        counterparty: data.counterparty,
        total_amount: data.total_amount,
        items: data.metadata ? JSON.parse(data.metadata).items : null,
        suggestions: data.suggestions || []
      }
    }
    
    this.setData({
      messages: [...this.data.messages, aiMsg],
      loading: false
    })
  },

  // 确认记录
  async confirmRecord(e) {
    const id = e.currentTarget.dataset.id
    // 实际场景中可能需要单独的确认API，这里直接跳转到记录页
    wx.showToast({ title: '已确认', icon: 'success' })
    
    // 从消息列表移除该确认
    const messages = this.data.messages.map(m => {
      if (m.data && m.data.id === id) {
        return { ...m, confirmed: true }
      }
      return m
    })
    this.setData({ messages })
  },

  // 取消记录
  async cancelRecord(e) {
    const id = e.currentTarget.dataset.id
    wx.showToast({ title: '已取消', icon: 'none' })
    
    // 移除该消息
    const messages = this.data.messages.filter(m => m.id !== id && m.data && m.data.id !== id)
    this.setData({ messages })
  },

  // 上传图片
  async uploadImage(filePath) {
    return new Promise((resolve, reject) => {
      wx.uploadFile({
        url: app.globalData.baseUrl + '/record/image',
        filePath: filePath,
        name: 'image',
        formData: { merchant_id: this.data.merchantId },
        success: (res) => {
          const data = JSON.parse(res.data)
          if (data.error) {
            reject(new Error(data.error))
          } else {
            resolve(data)
          }
        },
        fail: reject
      })
    })
  },

  // 上传语音
  async uploadVoice(filePath) {
    return new Promise((resolve, reject) => {
      wx.uploadFile({
        url: app.globalData.baseUrl + '/record/voice',
        filePath: filePath,
        name: 'audio',
        formData: { merchant_id: this.data.merchantId },
        success: (res) => {
          const data = JSON.parse(res.data)
          if (data.error) {
            reject(new Error(data.error))
          } else {
            resolve(data)
          }
        },
        fail: reject
      })
    })
  }
})
