const app = getApp()
const { wechatLogin, createRecord } = require('../../api/api')

Page({
  data: {
    merchantId: '',
    messages: [],
    inputText: '',
    inputMode: 'text',
    voiceRecording: false,
    voiceTempFilePath: '',
    loading: false,
    scrollTop: 0
  },

  onLoad() {
    this.initMerchant()
  },

  onShow() {
    // 刷新消息列表
  },

  // 初始化商户（微信登录）
  async initMerchant() {
    let merchantId = app.globalData.merchantId
    
    if (!merchantId) {
      try {
        // 先获取登录code
        const loginRes = await this.wxLogin()
        
        // 用code换取商户信息（会自动创建）
        const res = await wechatLogin(loginRes.code)
        
        merchantId = res.id
        app.setMerchantId(merchantId)
      } catch (err) {
        console.error('微信登录失败', err)
        wx.showToast({ title: '登录失败', icon: 'none' })
        return
      }
    }
    
    this.setData({ merchantId })
  },

  // 微信登录获取code
  wxLogin() {
    return new Promise((resolve, reject) => {
      wx.login({
        success: resolve,
        fail: reject
      })
    })
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
    if (!this.data.merchantId) {
      wx.showToast({ title: '正在初始化...', icon: 'none' })
      return
    }
    
    this.setData({ inputText: '' })
    await this.processUserMessage(text, 'text')
  },

  // 开始录音
  startVoice(e) {
    if (this.data.voiceRecording) return
    if (!this.data.merchantId) {
      wx.showToast({ title: '正在初始化...', icon: 'none' })
      return
    }
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
    if (!this.data.merchantId) {
      wx.showToast({ title: '正在初始化...', icon: 'none' })
      return
    }
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
    
    const userMsg = {
      id: msgId,
      role: 'user',
      type: type,
      content: content,
      duration: type === 'voice' ? 5 : 0
    }
    
    const messages = [...this.data.messages, userMsg]
    this.setData({ messages: messages, loading: true })
    
    try {
      if (type === 'text') {
        const res = await createRecord({
          merchant_id: this.data.merchantId,
          user_input: content
        })
        this.addConfirmMessage(res)
      } else if (type === 'image') {
        const res = await this.uploadImage(content)
        this.addConfirmMessage(res)
      } else if (type === 'voice') {
        const res = await this.uploadVoice(content)
        this.addConfirmMessage(res)
      }
    } catch (err) {
      console.error('处理消息失败', err)
      wx.showToast({ title: '处理失败', icon: 'none' })
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
    wx.showToast({ title: '已确认', icon: 'success' })
    
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
