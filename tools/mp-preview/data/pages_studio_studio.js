const { art } = require("../mock")
module.exports = {
  "boards": [
    {
      "id": 1,
      "name": "小姑娘",
      "thumb": art(31, 200, 250)
    },
    {
      "id": 2,
      "name": "小男孩",
      "thumb": art(32, 200, 250)
    },
    {
      "id": 3,
      "name": "小兔子",
      "thumb": art(33, 200, 250)
    }
  ],
  "boardIdx": 0,
  "cats": [
    {
      "key": "hair",
      "label": "头发"
    },
    {
      "key": "hat",
      "label": "帽子"
    },
    {
      "key": "outfit",
      "label": "服装"
    },
    {
      "key": "expression",
      "label": "表情"
    }
  ],
  "catIdx": 0,
  "elements": [
    {
      "id": 1,
      "name": "素材 1",
      "full": art(40, 200, 200)
    },
    {
      "id": 2,
      "name": "素材 2",
      "full": art(41, 200, 200)
    },
    {
      "id": 3,
      "name": "素材 3",
      "full": art(42, 200, 200)
    },
    {
      "id": 4,
      "name": "素材 4",
      "full": art(43, 200, 200)
    },
    {
      "id": 5,
      "name": "素材 5",
      "full": art(44, 200, 200)
    },
    {
      "id": 6,
      "name": "素材 6",
      "full": art(45, 200, 200)
    },
    {
      "id": 7,
      "name": "素材 7",
      "full": art(46, 200, 200)
    },
    {
      "id": 8,
      "name": "素材 8",
      "full": art(47, 200, 200)
    }
  ],
  "placements": [],
  "selectedIdx": -1,
  "composing": false
}
