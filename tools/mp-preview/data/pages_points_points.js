const { art } = require("../mock")
module.exports = {
  "points": 12,
  "total": 3,
  "loading": false,
  "logs": [
    {
      "id": 1,
      "delta": 2,
      "deltaText": "+2",
      "balance": 12,
      "title": "作品被使用",
      "icon": "sparkles",
      "remark": "",
      "time": "2026-11-07 10:20"
    },
    {
      "id": 2,
      "delta": -1,
      "deltaText": "-1",
      "balance": 10,
      "title": "AI 生成抵扣",
      "icon": "wand",
      "remark": "",
      "time": "2026-11-06 21:02"
    },
    {
      "id": 3,
      "delta": 1,
      "deltaText": "+1",
      "balance": 11,
      "title": "作品被使用",
      "icon": "sparkles",
      "remark": "",
      "time": "2026-11-05 09:14"
    }
  ],
  "hasMore": false,
  "loadingMore": false
}
