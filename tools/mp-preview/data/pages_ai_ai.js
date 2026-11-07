const { art } = require("../mock")
module.exports = {
  "mode": "ai_text",
  "suggestions": [
    "扎双马尾的小姑娘",
    "水彩风格小猫",
    "赛博朋克城市夜景",
    "复古邮票插画"
  ],
  "prompt": "扎双马尾的小姑娘，粉色背景，卡通风格",
  "photoPath": "",
  "generating": false,
  "result": {
    "cover": art(95, 700, 700),
    "type": "ai_text"
  },
  "quota": {
    "freeLimit": 5,
    "freeLeft": 2,
    "points": 12
  },
  "btnText": "开始生成（今日还剩 2 张）"
}
