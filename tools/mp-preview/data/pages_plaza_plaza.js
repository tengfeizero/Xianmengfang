const { art } = require("../mock")
module.exports = {
  "tabs": [
    {
      "key": "latest",
      "label": "最新"
    },
    {
      "key": "uses",
      "label": "最常用"
    },
    {
      "key": "likes",
      "label": "最多赞"
    }
  ],
  "tabKey": "latest",
  "items": [
    {
      "id": 1,
      "cover": art(1, 400, 500),
      "prompt": "",
      "typeText": "拼装作品",
      "isAI": false,
      "likeCount": 3,
      "useCount": 0,
      "liked": true,
      "isMine": false,
      "author": "桃桃",
      "avatar": "",
      "pop": false
    },
    {
      "id": 2,
      "cover": art(2, 400, 500),
      "prompt": "扎双马尾的小姑娘，粉色背景，卡通风格，柔和光线",
      "typeText": "AI 文生图",
      "isAI": true,
      "likeCount": 8,
      "useCount": 1,
      "liked": false,
      "isMine": false,
      "author": "阿岚",
      "avatar": "",
      "pop": false
    },
    {
      "id": 3,
      "cover": art(3, 400, 500),
      "prompt": "",
      "typeText": "AI 文生图",
      "isAI": true,
      "likeCount": 13,
      "useCount": 2,
      "liked": false,
      "isMine": false,
      "author": "Momo",
      "avatar": "",
      "pop": false
    },
    {
      "id": 4,
      "cover": art(4, 400, 500),
      "prompt": "扎双马尾的小姑娘，粉色背景，卡通风格，柔和光线",
      "typeText": "拼装作品",
      "isAI": false,
      "likeCount": 18,
      "useCount": 3,
      "liked": false,
      "isMine": false,
      "author": "小圆",
      "avatar": "",
      "pop": false
    },
    {
      "id": 5,
      "cover": art(5, 400, 500),
      "prompt": "",
      "typeText": "AI 文生图",
      "isAI": true,
      "likeCount": 23,
      "useCount": 4,
      "liked": true,
      "isMine": false,
      "author": "桃桃",
      "avatar": "",
      "pop": false
    },
    {
      "id": 6,
      "cover": art(6, 400, 500),
      "prompt": "扎双马尾的小姑娘，粉色背景，卡通风格，柔和光线",
      "typeText": "AI 文生图",
      "isAI": true,
      "likeCount": 28,
      "useCount": 5,
      "liked": false,
      "isMine": false,
      "author": "阿岚",
      "avatar": "",
      "pop": false
    }
  ],
  "loading": false,
  "hasMore": false,
  "loadingMore": false
}
