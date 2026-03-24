const app = getApp()
const { wechatLogin, createMerchant, joinMerchant, createRecord } = require('../../api/api')

Page({
  data: {
    userId: '',
    merchantId: '',
    merchants: [],
    hasMerchant: false,
    showAuthModal: false,
    showJoinModal: false,
    showCreateModal: false,
    inviteCode: '',
    newMerchantName: '',
    newMerchantType: 'fish',
    messages: [],
    inputText: '',
    inputMode: 'text',
    voiceRecording: false,
    voiceTempFilePath: '',
    loading: false
  },

  onLoad() {
    this.initApp()
  },

  onShow() {
    // 如果已选择商户，刷新数据
    if (app.globalData.merchantId) {
      this.setData({ merchantId: app.globalData.merchantId })
    }
  },

  // 初始化应用（微信登录）
  async initApp() {
    try {
      // 获取登录code
      const loginRes = await this.wxLogin()
      
      // 调用登录接口
      const res = await wechatLogin(loginRes.code)
      
      // 保存用户信息
      app.setUserId(res.user.id)
      this.setData({ userId: res.user.id })
      
      // 保存商户列表
      this.setData({
        merchants: res.merchants || [],
        hasMerchant: res.has_merchant
      })
      
      if (res.has_merchant && res.merchants.length > 0) {
        // 已有商户，自动选择第一个
        const merchant = res.merchants[0]
        app.setMerchantId(merchant.id)
        this.setData({ merchantId: merchant.id })
      } else {
        // 没有商户，弹出选择
        this.setData({ showAuthModal: true })
      }
    } catch (err) {
      console.error('初始化失败', err)
      wx.showToast({ title: '初始化失败', icon: 'none' })
    }
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

  // 关闭授权弹窗
  closeAuthModal() {
    this.setData({ showAuthModal: false })
  },

  // 打开发起商户弹窗
  showCreateModal() {
    this.setData({ showCreateModal: true, showAuthModal: false })
  },

  // 关闭创建商户弹窗
  closeCreateModal() {
    this.setData({ showCreateModal: false })
  },

  // 打开加入商户弹窗
  showJoinModal() {
    this.setData({ showJoinModal: true, showAuthModal: false })
  },

  // 关闭加入商户弹窗
  closeJoinModal() {
    this.setData({ showJoinModal: false })
  },

  // 输入新商户名称
  onNewMerchantName(e) {
    this.setData({ newMerchantName: e.detail.value })
  },

  // 选择商户类型
  onMerchantTypeChange(e) {
    const types = ['fish', 'vegetable', 'oil']
    this.setData({ newMerchantType: types[e.detail.value] })
  },

  // 创建商户
  async handleCreateMerchant() {
    const { userId, newMerchantName, newMerchantType } = this.data
    if (!newMerchantName.trim()) {
      wx.showToast({ title: '请输入商户名称', icon: 'none' })
      return
    }

    try {
      const res = await createMerchant({
        user_id: userId,
        name: newMerchantName,
        business_type: newMerchantType
      })
      
      // 保存商户ID
      app.setMerchantId(res.id)
      this.setData({
        merchantId: res.id,
        showCreateModal: false,
        hasMerchant: true,
        merchants: [res]
      })
      
      wx.showToast({ title: '创建成功', icon: 'success' })
    } catch (err) {
      console.error('创建商户失败', err)
      wx.showToast({ title: '创建失败', icon: 'none' })
    }
  },

  // 输入邀请码
  onInviteCode(e) {
    this.setData({ inviteCode: e.detail.value })
  },

  // 加入商户
  async handleJoinMerchant() {
    const { userId, inviteCode } = this.data
    if (!inviteCode.trim()) {
      wx.showToast({ title: '请输入邀请码', icon: 'none' })
      return
    }

    try {
      const res = await joinMerchant({
        user_id: userId,
        invite_code: inviteCode.toUpperCase()
      })
      
      // 保存商户ID
      app.setMerchantId(res.id)
      this.setData({
        merchantId: res.id,
        showJoinModal: false,
        hasMerchant: true,
        merchants: [...this.data.merchants, res]
      })
      
      wx.showToast({ title: '加入成功', icon: 'success' })
    } catch (err) {
      console.error('加入商户失败', err)
      wx.showToast({ title: err.error || '加入失败', icon: 'none' })
    }
  },

  // 切换商户
  switchMerchant(e) {
    const merchantId = e.currentTarget.dataset.id
    app.setMerchantId(merchantId)
    this.setData({ merchantId })
    wx.showToast({ title: '已切换', icon: 'success' })
  },

  // 复制邀请码
  copyInviteCode() {
    const merchant = this.data.merchants.find(m => m.id === this.data.merchantId)
    if (merchant && merchant.invite_code) {
      wx.setClipboardData({
        data: merchant.invite_code,
        success: () => {
          wx.showToast({ title: '已复制', icon: 'success' })
        }
      })
    }
  },

  // ============ 以下是聊天相关功能 ============

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
      wx.showToast({ title: '请先选择商户', icon: 'none' })
      return
    }
    
    this.setData({ inputText: '' })
    await this.processUserMessage(text, 'text')
  },

  // 开始录音
  startVoice(e) {
    if (this.data.voiceRecording) return
    if (!this.data.merchantId) {
      wx.showToast({ title: '请先选择商户', icon: 'none' })
      return
    }
    this.setData({ voiceRecording: true })
    wx.startRecord({
      success: (res) => {
        this.setData({ voiceTempFilePath: res.tempFilePath })
      },
      fail: (err) => {
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
    wx.playVoice({ filePath: e.currentTarget.dataset.url })
  },

  // 选择图片
  pickImage() {
    if (!this.data.merchantId) {
      wx.showToast({ title: '请先选择商户', icon: 'none' })
      return
    }
    wx.chooseImage({
      count: 1,
      sourceType: ['album', 'camera'],
      success: (res) => {
        this.processUserMessage(res.tempFilePaths[0], 'image')
      }
    })
  },

  // 预览图片
  previewImage(e) {
    wx.previewImage({ urls: [e.currentTarget.dataset.url] })
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
    this.setData({ messages, loading: true })
    
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
  confirmRecord(e) {
    wx.showToast({ title: '已确认', icon: 'success' })
    const messages = this.data.messages.map(m => {
      if (m.data && m.data.id === e.currentTarget.dataset.id) {
        return { ...m, confirmed: true }
      }
      return m
    })
    this.setData({ messages })
  },

  // 取消记录
  cancelRecord(e) {
    wx.showToast({ title: '已取消', icon: 'none' })
    const messages = this.data.messages.filter(m => 
      m.id !== e.currentTarget.dataset.id && m.data && m.data.id !== e.currentTarget.dataset.id
    )
    this.setData({ messages })
  },

  // 上传图片
  uploadImage(filePath) {
    return new Promise((resolve, reject) => {
      wx.uploadFile({
        url: app.globalData.baseUrl + '/record/image',
        filePath,
        name: 'image',
        formData: { merchant_id: this.data.merchantId },
        success: (res) => {
          const data = JSON.parse(res.data)
          if (data.error) reject(new Error(data.error))
          else resolve(data)
        },
        fail: reject
      })
    })
  },

  // 上传语音
  uploadVoice(filePath) {
    return new Promise((resolve, reject) => {
      wx.uploadFile({
        url: app.globalData.baseUrl + '/record/voice',
        filePath,
        name: 'audio',
        formData: { merchant_id: this.data.merchantId },
        success: (res) => {
          const data = JSON.parse(res.data)
          if (data.error) reject(new Error(data.error))
          else resolve(data)
        },
        fail: reject
      })
    })
  }
})
