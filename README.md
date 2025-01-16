# 项目说明
## 扩展安装
````
go get github.com/spf13/viper
go get -u github.com/gin-gonic/gin
go get -u gorm.io/gorm
go get -u gorm.io/driver/mysql
go get -u github.com/golang-jwt/jwt/v4
go get github.com/gorilla/websocket
````
## 项目启动
dockerdocker 安装 MySQL

````
# MySQL 
docker run -d --name mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=123456 mysql
````

1.连接 MySQL，创建 gochat 库

2.执行 sql/create_table.sql 文件中 SQL 代码

# 接口文档

## HTTP接口文档

### 用户模块
1.用户注册
* 地址：/register
* 请求方式：POST
* 请求参数：

  | 参数           | 必填 | 类型     | 说明  | 示例 |
  |--------------|----|--------|-----|---|
  | phone_number | 是  | string | 手机号 | 1 |
  | nickname     | 是  | string | 昵称  | 张三 |
  |  password     | 是  | string | 密码  | 1 |

* 返回参数：
  ````
  {
    "code": 200,
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.-yYQqERIlBq_D5t14eUg7fCDbYoH1t3LQO0vb4bt2kA",
        "user_id": "1"
    },
    "msg": "注册成功"
  }
  ````

2.用户登录
* 地址：/login
* 请求方式：POST
* 请求参数：

  | 参数           | 必填 | 类型     | 说明  | 示例 |
  |--------------|----|--------|-----|---|
  | phone_number | 是  | string | 手机号 | 1 |
  |  password     | 是  | string | 密码  | 1 |

* 返回参数：
  ````
  {
    "code": 200,
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.-yYQqERIlBq_D5t14eUg7fCDbYoH1t3LQO0vb4bt2kA",
        "user_id": "1"
    },
    "msg": "登录成功"
  }
  ````
  
3.个人账号详情
* 地址：/user/detail
* 请求方式：GET
* 返回参数：
  ````
  {
    "code": 200,
    "data": {
        "id": 1,
        "phone_number": "1",
        "nickname": "张三",
        "password": "c4ca4238a0b923820dcc509a6f75849b",
        "create_time": "2025-01-10T10:50:45+08:00",
        "update_time": "2025-01-10T10:50:45+08:00"
    },
    "msg": "数据加载成功"
  }
  ````

3.添加好友
* 地址：/user/add
* 请求方式：POST
* 请求参数：

  | 参数           | 必填 | 类型     | 说明    | 示例 |
  |--------------|----|--------|-------|----|
  | phone_number | 是  | string | 好友手机号 | 2  |
* 返回参数：
  ````
  {
    "code": 200,
    "msg": "添加成功"
  }
  ````
3.查询指定用户详情
* 地址：/user/list
* 请求方式：GET
* 请求参数：

  | 参数           | 必填 | 类型     | 说明    | 示例 |
  |--------------|----|--------|-------|----|
  | phone_number | 是  | string | 好友手机号 | 2  |

* 返回参数：
  ````
  {
    "code": 200,
    "data": {
        "phone_number": "2",
        "nickname": "李四",
        "is_friend": true
    },
    "msg": "数据加载成功"
  }
  ````
4.删除好友
* 地址：/user/delete
* 请求方式：DELETE
* 请求参数：

  | 参数           | 必填 | 类型     | 说明    | 示例 |
  |--------------|----|--------|-------|----|
  | phone_number | 是  | string | 好友手机号 | 2  |

* 返回参数：
  ````
  {
    "code": 200,
    "msg": "删除成功"
  }
  ````
### 通讯模块
1.基于GIN搭WebScocket服务

2.发送、接收消息
* 地址：ws://localhost:8080/websocket/message
* 消息格式：
  ````
  {
    "message": "你好，我是2",
    "group_id": 1
  }
  ````
3.聊天记录列表
* 地址：/message/list
* 请求方式：GET
  * 请求参数：
  
    | 参数         | 必填 | 类型     | 说明     | 示例 |
    |------------|----|--------|--------|----|
    | group_id   | 是  | string | 所在群组id | 1  |
    | page_index | 是  | string | 页面索引   | 1  |
    |  page_size  | 是  | string | 页面大小   | 1  |

* 返回参数：
  ````
  {
    "code": 200,
    "data": {
        "list": [
            {
                "id": 1,
                "user_id": 1,
                "GroupID": 1,
                "content": "5L2g5aW977yM5oiR5pivMQ==",
                "create_time": "2025-01-10T19:08:33+08:00",
                "update_time": "2025-01-10T19:08:33+08:00"
            }
        ]
    },
    "msg": "数据加载成功"
  }
  ````

# todo
* 通讯模块
  * 心跳检测
  * websocket优雅退出
  * 群聊功能
  * 消息可靠性&有序性：超时重传 + 消息确认机制
  * 在离线分离