项目结构
<span class="file-node"><span class="text-content">
@workspace</span></span></span>/
│
├── backend/
│ ├── controller/
│ │ ├── auth_controller.go
│ ├── models/
│ │ ├── baseuser.go
│ │ ├── logininput.go
│ │ ├── merchant.go
│ │ ├── riders.go
│ │ ├── orders.go
│ │ └── dish_category.go
│ ├── router/
│ │ ├── router.go
│ ├── config/
│ │ ├── config.go
│ │ ├── config.yaml
│ │ └── db.go
│ ├── global/
│ │ ├── global.go
│ │ └── db.go
│ ├── utils/
│ │ ├── utils.go
│ │ └── jwt.go
│ ├── go.mod
│ ├── go.sum
│ └── main.go

运行步骤

1. 安装 go 环境

2.安装 mysql 数据库

3.安装 go.mod 中的环境依赖

4.配置 config.yaml 中的数据库连接信息,主要是密码

5.在 backend 目录下运行 go run .
