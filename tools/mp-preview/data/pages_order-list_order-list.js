const { art } = require("../mock")
module.exports = {
  "tabs": [
    {
      "key": "",
      "label": "全部"
    },
    {
      "key": "pending_accept",
      "label": "待接单"
    },
    {
      "key": "active",
      "label": "进行中"
    },
    {
      "key": "shipped",
      "label": "已发货"
    },
    {
      "key": "done",
      "label": "已完成"
    },
    {
      "key": "canceled",
      "label": "已取消"
    }
  ],
  "tabIdx": 0,
  "loading": false,
  "orders": [
    {
      "id": 1,
      "orderNo": "XM20261107001",
      "status": "shipped",
      "totalText": "177.00",
      "itemCount": 3,
      "thumbs": [
        art(71, 300, 300),
        art(72, 300, 300)
      ],
      "firstTitle": "小梦女孩亚克力立牌 · 15cm 双面",
      "shipment": {
        "company": "顺丰速运",
        "trackingNo": "SF20261107"
      }
    },
    {
      "id": 2,
      "orderNo": "XM20261106002",
      "status": "producing",
      "totalText": "59.00",
      "itemCount": 1,
      "thumbs": [
        art(73, 300, 300)
      ],
      "firstTitle": "定制帆布包 · 米白",
      "shipment": null
    },
    {
      "id": 3,
      "orderNo": "XM20261105003",
      "status": "done",
      "totalText": "39.00",
      "itemCount": 1,
      "thumbs": [
        art(74, 300, 300)
      ],
      "firstTitle": "小梦女孩亚克力立牌 · 10cm",
      "shipment": null
    }
  ]
}
