const { art } = require("../mock")
module.exports = {
  "profile": {
    "nickname": "桃桃",
    "avatar": ""
  },
  "points": 12,
  "stats": {
    "total": 8,
    "published": 3,
    "likes": 26
  },
  "filters": [
    {
      "key": "all",
      "label": "全部"
    },
    {
      "key": "public",
      "label": "已公开"
    },
    {
      "key": "private",
      "label": "私密"
    }
  ],
  "filterKey": "all",
  "shown": [
    {
      "id": 20,
      "cover": art(20, 400, 500),
      "prompt": "",
      "typeText": "拼装作品",
      "isAI": false,
      "likeCount": 3,
      "useCount": 0,
      "liked": true,
      "isMine": true,
      "author": "桃桃",
      "avatar": "",
      "pop": false,
      "isPublic": true
    },
    {
      "id": 21,
      "cover": art(21, 400, 500),
      "prompt": "扎双马尾的小姑娘，粉色背景，卡通风格，柔和光线",
      "typeText": "AI 文生图",
      "isAI": true,
      "likeCount": 8,
      "useCount": 1,
      "liked": false,
      "isMine": true,
      "author": "阿岚",
      "avatar": "",
      "pop": false,
      "isPublic": false
    },
    {
      "id": 22,
      "cover": art(22, 400, 500),
      "prompt": "",
      "typeText": "AI 文生图",
      "isAI": true,
      "likeCount": 13,
      "useCount": 2,
      "liked": false,
      "isMine": true,
      "author": "Momo",
      "avatar": "",
      "pop": false,
      "isPublic": true
    },
    {
      "id": 23,
      "cover": art(23, 400, 500),
      "prompt": "扎双马尾的小姑娘，粉色背景，卡通风格，柔和光线",
      "typeText": "拼装作品",
      "isAI": false,
      "likeCount": 18,
      "useCount": 3,
      "liked": false,
      "isMine": true,
      "author": "小圆",
      "avatar": "",
      "pop": false,
      "isPublic": false
    }
  ],
  "loading": false
}
