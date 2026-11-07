const { art } = require("../mock")
module.exports = {
  "loading": false,
  "items": [
    {
      "id": 1,
      "title": "小梦女孩亚克力立牌 · 定制款",
      "specsText": "15cm 双面",
      "qty": 2,
      "checked": 1,
      "valid": 1,
      "invalidReason": "",
      "priceText": "59.00",
      "subtotalText": "118.00",
      "cover": "",
      "workThumb": art(61, 300, 300)
    },
    {
      "id": 2,
      "title": "定制帆布包 大容量单肩",
      "specsText": "米白 / 12oz",
      "qty": 1,
      "checked": 1,
      "valid": 1,
      "invalidReason": "",
      "priceText": "59.00",
      "subtotalText": "59.00",
      "cover": art(62, 300, 300),
      "workThumb": ""
    },
    {
      "id": 3,
      "title": "已下架商品",
      "specsText": "",
      "qty": 1,
      "checked": 0,
      "valid": 0,
      "invalidReason": "商品已下架，无法结算",
      "priceText": "29.00",
      "subtotalText": "29.00",
      "cover": art(63, 300, 300),
      "workThumb": ""
    }
  ],
  "allChecked": false,
  "checkedCount": 3,
  "totalText": "177.00",
  "selectedAddress": {
    "id": 1,
    "name": "桃桃",
    "phone": "13800001111",
    "full": "广东省 深圳市 南山区 科技园 1 号"
  },
  "submitting": false
}
