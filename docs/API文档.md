# 商户记账系统 API 文档 v1.0

## 概述

- **Base URL**: `http://localhost:8080/api/v1`
- **认证方式**: 简化版（merchant_id 直接传递，后续迭代加入token）
- **数据格式**: JSON

---

## 1. 商户管理

### 1.1 创建商户

**POST** `/merchant`

Request:
```json
{
  "name": "老王鱼档",
  "business_type": "fish"
}
```

Response (201):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "老王鱼档",
  "business_type": "fish",
  "created_at": "2026-03-23T08:00:00Z"
}
```

---

## 2. 记录管理

### 2.1 提交记录（核心接口 - LLM解析）

**POST** `/record`

Request:
```json
{
  "merchant_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_input": "给老王送了3斤鲈鱼，45一斤，另外送了2把葱"
}
```

Response (201):
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "merchant_id": "550e8400-e29b-41d4-a716-446655440000",
  "record_type": "delivery",
  "occurred_at": "2026-03-23T08:30:00Z",
  "counterparty": "老王",
  "total_amount": 135.00,
  "metadata": {
    "items": [
      {"name": "鲈鱼", "qty": 3, "unit": "斤", "price": 45},
      {"name": "葱", "qty": 2, "unit": "把"}
    ]
  },
  "created_at": "2026-03-23T08:30:00Z"
}
```

### 2.2 查询记录列表

**GET** `/records`

Query Parameters:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| merchant_id | string | 是 | 商户ID |
| start | string | 否 | 开始日期 (2006-01-02) |
| end | string | 否 | 结束日期 (2006-01-02) |
| type | string | 否 | 记录类型: purchase/delivery/payment |
| q | string | 否 | 自然语言查询，如 "老王的账" |

Example:
```
GET /records?merchant_id=xxx&start=2026-03-01&end=2026-03-23
GET /records?merchant_id=xxx&q=上周从老王那里进了哪些货
```

Response (200):
```json
{
  "records": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "merchant_id": "550e8400-e29b-41d4-a716-446655440000",
      "record_type": "delivery",
      "occurred_at": "2026-03-23T08:30:00Z",
      "counterparty": "老王",
      "total_amount": 135.00,
      "metadata": {...},
      "created_at": "2026-03-23T08:30:00Z"
    }
  ],
  "total": 1
}
```

### 2.3 获取单条记录

**GET** `/records/:id`

Response (200):
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "merchant_id": "550e8400-e29b-41d4-a716-446655440000",
  "record_type": "delivery",
  "occurred_at": "2026-03-23T08:30:00Z",
  "counterparty": "老王",
  "total_amount": 135.00,
  "metadata": {
    "items": [
      {"name": "鲈鱼", "qty": 3, "unit": "斤", "price": 45},
      {"name": "葱", "qty": 2, "unit": "把"}
    ]
  },
  "created_at": "2026-03-23T08:30:00Z"
}
```

### 2.4 更新记录

**PUT** `/records/:id`

Request:
```json
{
  "counterparty": "老李",
  "total_amount": 150.00,
  "metadata": {
    "items": [
      {"name": "鲈鱼", "qty": 3, "unit": "斤", "price": 45},
      {"name": "葱", "qty": 3, "unit": "把", "price": 5}
    ]
  }
}
```

Response (200):
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "merchant_id": "550e8400-e29b-41d4-a716-446655440000",
  "record_type": "delivery",
  "occurred_at": "2026-03-23T08:30:00Z",
  "counterparty": "老李",
  "total_amount": 150.00,
  "metadata": {...},
  "created_at": "2026-03-23T08:30:00Z"
}
```

### 2.5 删除记录

**DELETE** `/records/:id`

Response (204): No Content

---

## 3. 统计报表

### 3.1 今日统计

**GET** `/stats/today`

Query Parameters:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| merchant_id | string | 是 | 商户ID |

Response (200):
```json
{
  "date": "2026-03-23",
  "total_delivery_amount": 1350.00,
  "total_delivery_count": 5,
  "total_purchase_amount": 2000.00,
  "total_purchase_count": 3,
  "total_payment_amount": 500.00,
  "total_payment_count": 2,
  "recent_records": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "record_type": "delivery",
      "counterparty": "老王",
      "total_amount": 135.00,
      "occurred_at": "2026-03-23T08:30:00Z"
    }
  ]
}
```

### 3.2 日报

**GET** `/report/daily`

Query Parameters:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| merchant_id | string | 是 | 商户ID |
| date | string | 否 | 日期 (2006-01-02)，默认当天 |

Response (200):
```json
{
  "date": "2026-03-23",
  "records": [...],
  "summary": {
    "total_in": 1350.00,
    "total_out": 2000.00,
    "net": -650.00
  }
}
```

### 3.3 周期报表

**GET** `/report/periodic`

Query Parameters:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| merchant_id | string | 是 | 商户ID |
| type | string | 是 | 周期类型: weekly/monthly |
| date | string | 否 | 基准日期，默认当天 |

Example:
```
GET /report/periodic?merchant_id=xxx&type=weekly&date=2026-03-23
```

Response (200):
```json
{
  "type": "weekly",
  "start_date": "2026-03-17",
  "end_date": "2026-03-23",
  "records": [...],
  "summary": {
    "total_delivery": 5000.00,
    "total_purchase": 3000.00,
    "total_payment": 1000.00,
    "profit": 2000.00
  },
  "by_counterparty": [
    {"name": "老王", "total": 1000.00, "count": 5},
    {"name": "老李", "total": 800.00, "count": 3}
  ]
}
```

---

## 4. 送货提醒

### 4.1 创建提醒

**POST** `/reminders`

Request:
```json
{
  "merchant_id": "550e8400-e29b-41d4-a716-446655440000",
  "customer": "老王饭店",
  "items": "3斤鲈鱼，2把葱",
  "remind_at": "2026-03-24T06:00:00Z",
  "notes": "明早6点送到"
}
```

Response (201):
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440002",
  "merchant_id": "550e8400-e29b-41d4-a716-446655440000",
  "customer": "老王饭店",
  "items": "3斤鲈鱼，2把葱",
  "remind_at": "2026-03-24T06:00:00Z",
  "notes": "明早6点送到",
  "status": "pending",
  "created_at": "2026-03-23T08:30:00Z"
}
```

### 4.2 获取提醒列表

**GET** `/reminders`

Query Parameters:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| merchant_id | string | 是 | 商户ID |
| status | string | 否 | pending/completed/expired |

Response (200):
```json
{
  "reminders": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440002",
      "merchant_id": "550e8400-e29b-41d4-a716-446655440000",
      "customer": "老王饭店",
      "items": "3斤鲈鱼，2把葱",
      "remind_at": "2026-03-24T06:00:00Z",
      "status": "pending",
      "created_at": "2026-03-23T08:30:00Z"
    }
  ]
}
```

### 4.3 更新提醒状态

**PUT** `/reminders/:id`

Request:
```json
{
  "status": "completed"
}
```

---

## 5. 错误响应

所有错误返回统一格式：

```json
{
  "error": {
    "code": "MERCHANT_NOT_FOUND",
    "message": "商户不存在"
  }
}
```

常见错误码：

| Code | HTTP Status | 说明 |
|------|-------------|------|
| VALIDATION_ERROR | 400 | 参数校验失败 |
| MERCHANT_NOT_FOUND | 404 | 商户不存在 |
| RECORD_NOT_FOUND | 404 | 记录不存在 |
| INTERNAL_ERROR | 500 | 服务器内部错误 |
| LLM_PARSE_ERROR | 500 | LLM解析失败 |

---

## 6. 业务类型 (business_type)

| Type | 说明 |
|------|------|
| fish | 鱼类商贩 |
| vegetable | 蔬菜商贩 |
| oil | 粮油商贩 |
| other | 其他 |

---

## 7. 记录类型 (record_type)

| Type | 说明 |
|------|------|
| purchase | 进货 |
| delivery | 送货 |
| payment | 付款/收款 |

---

*文档版本：v1.0*
*最后更新：2026-03-23*
