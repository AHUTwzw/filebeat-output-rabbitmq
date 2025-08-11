## fork版本
- 源项目地址：https://github.com/elastic/beats
- 分支：main
- 时间 2025.5.8
- revision-number:11713f01ac213eec5ea2017059a4f79ef41d0afb
## filebeat-output-rabbitmq
支持rabbitmq为输出端

### configuration
```
output.rabbitmq:
  host: localhost
  port: 5672
  username: guest
  password: guest
  vhost: vhost
  exchange: log.exchange
  routing: filebeat.routingkey
```

### 代码改动部分
- /libbeat/outputs/rabbitmq
- /libbeat/publisher/includes/
- go.mod
- go.sum

### 编译部署
查看.buildkite/filebeat/filebeat-pipeline.yml文件的command命令
### Docker下载
https://hub.docker.com/r/wzwahut/filebeat-rabbitmq
