const app = getApp()

const request = (url, method, data) => {
  return new Promise((resolve, reject) => {
    wx.request({
      url: app.globalData.baseUrl + url,
      method: method,
      data: data,
      header: {
        'Content-Type': 'application/json'
      },
      success: (res) => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(res.data)
        } else {
          reject(res.data)
        }
      },
      fail: (err) => {
        reject(err)
      }
    })
  })
}

// 微信登录
export const wechatLogin = (code) => request('/auth/wechat', 'POST', { code })

// 创建商户
export const createMerchant = (data) => request('/auth/create-merchant', 'POST', data)

// 通过邀请码加入商户
export const joinMerchant = (data) => request('/auth/join-merchant', 'POST', data)

// 商户相关
export const getMerchant = (id) => request(`/merchant/${id}`, 'GET')

// 记录相关
export const createRecord = (data) => request('/record', 'POST', data)
export const getRecords = (params) => request('/records', 'GET', null, { params })
export const getRecord = (id) => request(`/records/${id}`, 'GET')
export const updateRecord = (id, data) => request(`/records/${id}`, 'PUT', data)
export const deleteRecord = (id) => request(`/records/${id}`, 'DELETE')

// 统计报表
export const getTodayStats = (merchantId) => request('/stats/today', 'GET', null, { params: { merchant_id: merchantId } })
export const getDailyReport = (params) => request('/report/daily', 'GET', null, { params })
export const getPeriodicReport = (params) => request('/report/periodic', 'GET', null, { params })

// 送货提醒
export const createReminder = (data) => request('/reminders', 'POST', data)
export const getReminders = (params) => request('/reminders', 'GET', null, { params })
export const updateReminder = (id, data) => request(`/reminders/${id}`, 'PUT', data)
