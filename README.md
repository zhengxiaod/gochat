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

  | 参数           | 必填 | 类型     | 说明  | 实例     |
  |--------------|----|--------|-----|--------|
  | phone_number | 是  | string | 手机号 | 123456 |
  | nickname     | 是  | string | 昵称  | Tom    |
  |  password     | 是  | string | 密码  | 123456 |
    
* 返回参数：
    ````
    {
        "code": 200,
        "data": {
            "id": "3",
            "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjozfQ.buh_alYpfSGHTX5I31yMZ3NFt1h6nRUKgUiizRH73j0"
        },
        "msg": "登录成功"
    }
    ````
2.用户登录

### 通讯模块
1.基于GIN搭WebScocket服务

2.发送、接收消息

3.聊天记录列表

# todo
* 用户模块
  * 通过手机号发送验证码实现用户注册，基于Redis实现验证码存储
