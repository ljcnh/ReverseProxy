# Demo
反向代理的简单使用

### main

反向代理的具体实现，通过解析参数中的内容用来判断转发到哪个连接。

go run main.go

1. proxy_condition == "A"  转到A_CONDITION_URL
2. proxy_condition == "B"  转到B_CONDITION_URL
3. 其他：转到DEFAULT_CONDITION_URL

### 开启http服务

go run serverHttp1.go

go run serverHttp2.go

go run serverHttp3.go


### curl

    curl --request GET   --url http://localhost:1330/   --header 'content-type: application/json'   --data '{"proxy_condition":"b"}'