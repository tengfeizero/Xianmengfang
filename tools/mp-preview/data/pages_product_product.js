const { art } = require("../mock")
module.exports = {
  "loading": false,
  "galleryIdx": 0,
  "cartCount": 2,
  "product": {
    "id": 1,
    "title": "小梦女孩亚克力立牌 · 定制款",
    "subtitle": "双面高清印刷，附透明底座",
    "images": [
      art(51, 600, 600),
      art(52, 600, 600)
    ],
    "priceText": "39.00",
    "originalText": "49.00",
    "unit": "件",
    "detailParagraphs": [
      "采用 5mm 高透明亚克力，UV 打印双面成像，边缘打磨圆润。",
      "支持任意作品图案定制，下单后 48 小时内发出。"
    ],
    "detailImages": [
      art(53, 700, 400)
    ],
    "serviceNotes": "48 小时内发货"
  },
  "skus": [
    {
      "id": 1,
      "specsText": "10cm 单面",
      "priceText": "39.00",
      "stock": 10
    },
    {
      "id": 2,
      "specsText": "15cm 双面",
      "priceText": "59.00",
      "stock": 5
    }
  ],
  "selectedSku": {
    "id": 2,
    "specsText": "15cm 双面",
    "priceText": "59.00",
    "stock": 5
  },
  "qty": 1,
  "myWorks": [
    {
      "id": 21,
      "cover": art(54, 300, 380)
    },
    {
      "id": 22,
      "cover": art(55, 300, 380)
    },
    {
      "id": 23,
      "cover": art(56, 300, 380)
    }
  ],
  "likedWorks": [
    {
      "id": 31,
      "cover": art(57, 300, 380)
    },
    {
      "id": 32,
      "cover": art(58, 300, 380)
    }
  ],
  "selectedWork": {
    "id": 22,
    "cover": art(55, 300, 380)
  },
  "addresses": [],
  "selectedAddress": {
    "id": 1,
    "name": "桃桃",
    "phone": "13800001111",
    "full": "广东省 深圳市 南山区 科技园 1 号",
    "isDefault": true
  },
  "submitting": false,
  "totalText": "59.00"
}
