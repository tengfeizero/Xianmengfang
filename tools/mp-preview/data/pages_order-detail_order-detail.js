const { art } = require("../mock")
module.exports = {
  "loading": false,
  "steps": [
    {
      "key": "pending_accept",
      "label": "待接单"
    },
    {
      "key": "producing",
      "label": "生产中"
    },
    {
      "key": "packing",
      "label": "打包中"
    },
    {
      "key": "pending_ship",
      "label": "待发货"
    },
    {
      "key": "shipped",
      "label": "已发货"
    },
    {
      "key": "done",
      "label": "已完成"
    }
  ],
  "order": {
    "id": 1,
    "orderNo": "XM20261107001",
    "status": "shipped",
    "statusText": "已发货",
    "statusDesc": "商品已发出，请注意查收",
    "stepIdx": 4,
    "payType": "模拟支付（演示）",
    "payTime": "2026-11-06 12:30",
    "createdAt": "2026-11-06 12:28",
    "cancelReason": "",
    "refundFlag": 0,
    "receiverName": "桃桃",
    "receiverPhone": "13800001111",
    "receiverAddress": "广东省深圳市南山区科技园 1 号 A 座 2201"
  },
  "items": [
    {
      "id": 1,
      "title": "小梦女孩亚克力立牌 · 定制款",
      "specsText": "15cm 双面",
      "qty": 2,
      "priceText": "59.00",
      "subtotalText": "118.00",
      "imageUrl": art(75, 300, 300),
      "workImageUrl": art(76, 300, 300)
    },
    {
      "id": 2,
      "title": "定制帆布包 大容量单肩",
      "specsText": "米白 / 12oz",
      "qty": 1,
      "priceText": "59.00",
      "subtotalText": "59.00",
      "imageUrl": art(77, 300, 300),
      "workImageUrl": ""
    }
  ],
  "events": [
    {
      "id": 1,
      "action": "商家发货",
      "time": "2026-11-07 09:10",
      "note": "顺丰速运 SF20261107"
    },
    {
      "id": 2,
      "action": "打包完成，等待发货",
      "time": "2026-11-06 18:40",
      "note": ""
    },
    {
      "id": 3,
      "action": "商家接单，开始生产",
      "time": "2026-11-06 13:02",
      "note": ""
    },
    {
      "id": 4,
      "action": "提交订单",
      "time": "2026-11-06 12:28",
      "note": ""
    }
  ],
  "amounts": {
    "itemsText": "177.00",
    "shippingText": "0.00",
    "discountText": "0.00",
    "hasDiscount": false,
    "totalText": "177.00"
  },
  "shipment": {
    "company": "顺丰速运",
    "trackingNo": "SF20261107",
    "time": "2026-11-07 09:10"
  },
  "canCancel": false,
  "canConfirm": true
}
