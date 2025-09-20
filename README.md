# KubeSphere-OpenSearch 告警工具

基于 Go 语言开发的 OpenSearch 告警工具，专为 KubeSphere 环境设计，类似于 ElastAlert2，支持多种通知渠道和智能告警级别管理。

## 功能特性

- 🔍 **多索引查询**: 支持查询不同类型的 OpenSearch 索引（日志、事件、审计）
- 📊 **灵活规则**: 支持频率、任意、突增、突降等多种告警规则类型
- 🔔 **多通知渠道**: 支持邮件、钉钉、企业微信、飞书通知
- 🎯 **智能告警级别**: 支持 5 级告警分类，可自定义或自动判断
- 🚫 **告警抑制**: 支持告警去重和指数级抑制机制
- 📱 **消息格式优化**: 针对不同通知渠道优化消息格式和显示效果
- ⚡ **高性能**: 基于 Go 语言，性能优异
- 🐳 **容器化**: 支持 Docker 和 Kubernetes 部署
- 📝 **详细日志**: 提供 DEBUG 级别日志，便于调试和监控

## 项目结构

```
opensearch-alert/
├── cmd/alert/                 # 主程序入口
├── internal/                  # 内部包
│   ├── config/               # 配置管理
│   ├── opensearch/           # OpenSearch 客户端
│   ├── alert/                # 告警引擎
│   └── notification/         # 通知渠道
├── pkg/types/                # 类型定义
├── configs/                  # 配置文件
│   ├── config.yaml          # 主配置
│   └── rules/               # 告警规则
├── k8s/                     # Kubernetes 部署文件
└── Dockerfile               # Docker 构建文件
```

## 快速开始

### 1. 构建镜像

```bash
# 构建 Docker 镜像
docker build -t opensearch-alert:latest .

# 或者使用 Go 直接运行
go mod tidy
go run cmd/alert/main.go -config=configs/config.yaml -rules=configs/rules
```

### 2. 配置 OpenSearch 连接

编辑 `configs/config.yaml`:

```yaml
opensearch:
  host: "opensearch-cluster-data.kubesphere-logging-system"
  port: 9200
  protocol: "https"
  username: "admin"
  password: "admin"
  verify_certs: false
```

### 3. 配置通知渠道

#### 邮件配置
```yaml
notifications:
  email:
    enabled: true
    smtp_server: "smtp.qq.com"
    smtp_port: 587
    username: "your-email@qq.com"
    password: "your-auth-code"  # QQ邮箱使用授权码
    from_email: "your-email@qq.com"
    to_emails: ["admin@company.com"]
    use_tls: true
```

#### 钉钉配置
```yaml
notifications:
  dingtalk:
    enabled: true
    webhook_url: "https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN"
    secret: "YOUR_SECRET"
    at_mobiles: ["13800138000"]
    at_all: false
```

#### 企业微信配置
```yaml
notifications:
  wechat:
    enabled: true
    webhook_url: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR_KEY"
    mentioned_list: []
    mentioned_mobile_list: ["13800138000"]
    at_all: false
```

#### 飞书配置
```yaml
notifications:
  feishu:
    enabled: true
    webhook_url: "https://open.feishu.cn/open-apis/bot/v2/hook/YOUR_HOOK"
    secret: ""
    at_mobiles: []
    at_all: true
```

### 4. 配置告警规则

在 `configs/rules/` 目录下创建 YAML 规则文件:

```yaml
name: "Pod 异常事件告警"
type: "frequency"
index: "ks-whizard-events-*"
threshold: 1
timeframe: 60
query:
  bool:
    must:
      - term:
          type: "Warning"
      - bool:
          should:
            - term:
                reason: "BackOff"
            - term:
                reason: "Failed"
          minimum_should_match: 1
alert:
  - "dingtalk"
  - "wechat"
level: "High"  # 可选：Critical, High, Medium, Low, Info
enabled: true
```

### 告警级别配置

系统支持 5 个告警级别，用户可以在规则文件中自定义：

- **Critical（严重）**：系统组件错误、安全事件等，会@用户
- **High（高优先级）**：应用错误日志等，会@用户  
- **Medium（中等优先级）**：警告日志等，不会@用户
- **Low（低优先级）**：一般告警，不会@用户
- **Info（信息）**：测试或信息通知，不会@用户

#### 自定义告警级别
在规则文件中添加 `level` 字段：
```yaml
name: "数据库连接失败告警"
level: "Critical"  # 自定义级别
enabled: true
```

#### 自动级别判断
如果不指定 `level` 字段，系统会根据规则名称自动判断：
- 包含"系统组件"+"错误" → Critical
- 包含"安全" → Critical  
- 包含"fatal"或"panic" → Critical
- 包含"错误" → High
- 包含"警告" → Medium
- 其他 → Low

## 消息格式优化

### 钉钉消息
- 支持 Markdown 格式，使用 `  \n  ` 实现垂直排列
- 根据告警级别自动@用户（Critical、High 级别）
- 消息内容清晰，字段垂直排列

### 企业微信消息  
- 转换为纯文本格式，去除 Markdown 符号
- 根据告警级别自动@用户
- 消息简洁易读

### 飞书消息
- 使用 `lark_md` 格式，优化代码块显示
- 自动清理多余空行，保持格式整洁
- 根据告警级别自动@用户

### 邮件消息
- 使用 HTML 格式，支持丰富的样式
- 字段垂直排列，告警级别用颜色区分
- 自动转换 Markdown 为 HTML 格式
- 支持代码块、粗体等格式

## 日志配置

系统提供详细的日志记录，便于调试和监控：

```yaml
logging:
  level: "DEBUG"            # 日志级别：DEBUG, INFO, WARN, ERROR
  format: "2006-01-02 15:04:05 - %s - %s - %s"  # 日志格式
  file: "log/opensearch-alert/alert.log"  # 日志文件路径
  max_size: "10MB"          # 单个日志文件最大大小
  backup_count: 5           # 保留的日志文件备份数量
```

### 日志内容
- **OpenSearch 连接日志**：查询执行、响应状态等
- **规则加载日志**：规则文件扫描、加载状态等  
- **告警级别判断日志**：自动级别判断过程
- **通知发送日志**：各渠道发送成功/失败状态
- **告警引擎日志**：规则执行、告警触发等

## Kubernetes 部署

### 1. 创建 ConfigMap

```bash
kubectl apply -f k8s/configmap.yaml
```

### 2. 部署应用

```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

### 3. 查看日志

```bash
kubectl logs -f deployment/opensearch-alert -n kube-logging
```

## 告警规则类型

### 1. frequency (频率型)
在指定时间窗口内达到阈值时触发告警。

```yaml
type: "frequency"
threshold: 5
timeframe: 300  # 5分钟内
```

### 2. any (任意型)
匹配到任何记录即触发告警。

```yaml
type: "any"
threshold: 1
```

### 3. spike (突增型)
检测流量突增时触发告警。

```yaml
type: "spike"
threshold: 10
timeframe: 60
```

### 4. flatline (突降型)
检测流量低于阈值时触发告警。

```yaml
type: "flatline"
threshold: 5
timeframe: 300
```

## 告警抑制机制

### 1. 基础抑制
相同告警的最小间隔时间。

```yaml
realert: 5  # 5分钟内相同告警只发送一次
```

### 2. 指数级抑制
告警间隔逐渐增加，避免告警风暴。

```yaml
alert_suppression:
  enabled: true
  realert_minutes: 5
  exponential_realert:
    enabled: true
    hours: 1  # 每次重复告警间隔乘以1小时
```

## 监控的索引类型

### 1. 事件索引 (ks-whizard-events-*)
- Pod 异常事件
- 资源删除事件
- 重启事件

### 2. 日志索引 (ks-whizard-logging-*)
- 应用错误日志
- 系统错误日志
- 异常堆栈信息

### 3. 审计索引 (ks-whizard-auditing-*)
- 安全操作审计
- 权限变更审计
- 敏感资源操作

## 故障排除

### 常见问题

#### 1. 邮件发送失败
**错误信息**：`535 Login fail. Account is abnormal...`

**解决方案**：
- 检查 QQ 邮箱是否开启 SMTP 服务
- 使用授权码而不是登录密码
- 参考 `QQ邮箱配置指南.md` 详细配置步骤

#### 2. 钉钉消息格式问题
**问题**：消息横向挤在一起，没有垂直排列

**解决方案**：
- 确保使用 `  \n  ` 格式（两个空格+换行+两个空格）
- 检查 Markdown 格式是否正确

#### 3. 企业微信显示 Markdown 符号
**问题**：消息中显示 `**`、`---` 等符号

**解决方案**：
- 系统已自动转换为纯文本格式
- 如仍有问题，检查消息内容处理逻辑

#### 4. 飞书代码块不显示
**问题**：` ``` ` 代码块标记不生效

**解决方案**：
- 系统已自动清理代码块标记
- 使用 `lark_md` 格式优化显示

### 调试方法

1. **启用 DEBUG 日志**：
   ```yaml
   logging:
     level: "DEBUG"
   ```

2. **查看详细日志**：
   ```bash
   tail -f log/opensearch-alert/alert.log
   ```

3. **检查通知发送状态**：
   日志中会显示各渠道发送成功/失败的详细信息

## 开发说明

### 添加新的通知渠道

1. 在 `internal/notification/` 下创建新的通知器
2. 实现 `Notifier` 接口
3. 在 `notifier.go` 中注册新渠道
4. 添加相应的配置验证和错误处理

### 添加新的告警规则类型

1. 在 `internal/alert/engine.go` 的 `shouldTriggerAlert` 方法中添加新类型
2. 实现相应的检测逻辑
3. 添加相应的日志记录

### 自定义消息格式

1. 修改各通知渠道的 `buildXXXMessage` 方法
2. 添加消息内容格式化方法
3. 测试不同告警级别的显示效果

## 更新日志

### v2.0.0 (2025-09-20)
- ✨ 新增飞书通知支持
- ✨ 新增智能告警级别管理
- ✨ 新增邮件通知支持
- 🎨 优化各渠道消息格式
- 🐛 修复钉钉消息排列问题
- 🐛 修复企业微信 Markdown 符号问题
- 🐛 修复飞书代码块显示问题
- 📝 增强日志记录和调试功能
- 📚 完善配置注释和文档

## 许可证

MIT License
