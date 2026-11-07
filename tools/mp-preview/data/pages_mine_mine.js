const { art } = require("../mock")
module.exports = {
  "user": {
    "id": 3,
    "nickname": "桃桃",
    "avatar": "",
    "points": 12
  },
  "stats": {
    "works": 8,
    "public_works": 3,
    "liked": 5,
    "cart_count": 2,
    "orders": {
      "pending_accept": 1,
      "producing": 2,
      "packing": 0,
      "pending_ship": 0,
      "shipped": 1,
      "done": 4
    }
  },
  "orderEntries": [
    {
      "key": "pending_accept",
      "label": "待接单",
      "icon": "clock",
      "count": 1
    },
    {
      "key": "producing",
      "label": "进行中",
      "icon": "sparkles",
      "count": 2
    },
    {
      "key": "shipped",
      "label": "已发货",
      "icon": "truck",
      "count": 1
    },
    {
      "key": "done",
      "label": "已完成",
      "icon": "check",
      "count": 4
    }
  ],
  "cells": [
    {
      "key": "cart",
      "label": "购物车",
      "icon": "cart",
      "color": "#ff6fa5",
      "bg": "#ffeef4"
    },
    {
      "key": "address",
      "label": "收货地址",
      "icon": "pin",
      "color": "#14b8a6",
      "bg": "#e8f8f4"
    },
    {
      "key": "help",
      "label": "帮助与反馈",
      "icon": "help",
      "color": "#3b82f6",
      "bg": "#eaf1ff"
    },
    {
      "key": "settings",
      "label": "设置",
      "icon": "settings",
      "color": "#6c5ce7",
      "bg": "#f1eeff"
    },
    {
      "key": "about",
      "label": "关于现梦坊",
      "icon": "info",
      "color": "#5d576e",
      "bg": "#f3f1f7"
    }
  ],
  "version": "v2.1.0"
}
