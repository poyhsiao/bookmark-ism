# Language: zh-CN
# 功能: 安全监控和威胁检测
# 作为安全管理员，我需要实时监控安全威胁，以保护系统和用户数据

Feature: 安全监控和威胁检测
  为了及时发现和应对安全威胁
  作为安全管理员
  我需要全面的安全监控功能

  Background:
    Given 安全监控系统已经部署
    And 安全规则和告警已配置
    And 日志分析系统正在运行

  @security @rate-limiting
  Scenario: 速率限制监控
    Given 速率限制规则已配置
    When 某个IP地址在1分钟内发送100个请求
    Then 系统应该开始限制该IP的请求
    And 应该记录速率限制事件
    And 如果持续1小时应该触发安全告警

  @security @authentication
  Scenario: 认证失败监控
    Given 用户认证系统正在运行
    When 某个IP地址在5分钟内有10次登录失败
    Then 应该记录安全事件
    And 应该暂时封禁该IP地址30分钟
    And 应该发送安全告警通知

  @security @sql-injection
  Scenario: SQL注入攻击检测
    Given SQL注入检测规则已启用
    When 请求包含SQL注入特征
      """
      POST /api/v1/bookmarks
      {
        "title": "'; DROP TABLE bookmarks; --",
        "url": "http://example.com"
      }
      """
    Then 请求应该被拒绝
    And 应该返回HTTP 400响应
    And 应该记录安全威胁事件
    And 应该标记该请求来源为可疑

  @security @xss-protection
  Scenario: XSS攻击防护
    Given XSS防护规则已启用
    When 请求包含恶意脚本
      """
      POST /api/v1/bookmarks
      {
        "title": "<script>alert('xss')</script>",
        "url": "http://example.com"
      }
      """
    Then 恶意脚本应该被清理或拒绝
    And 应该记录XSS攻击尝试
    And 书签标题应该被安全处理

  @security @data-access
  Scenario: 异常数据访问监控
    Given 数据访问监控已启用
    When 用户尝试访问不属于他们的数据
    Then 访问应该被拒绝
    And 应该记录未授权访问尝试
    And 如果同一用户多次尝试应该触发告警

  @security @api-abuse
  Scenario: API滥用检测
    Given API使用监控已启用
    When 某个用户在短时间内创建大量书签
      | action           | count | timeframe |
      | create_bookmark  | 1000  | 1 minute  |
    Then 应该触发API滥用告警
    And 应该临时限制该用户的API访问
    And 应该记录滥用行为

  @security @malicious-files
  Scenario: 恶意文件上传检测
    Given 文件上传安全检查已启用
    When 用户尝试上传可疑文件
    Then 文件应该经过安全扫描
    And 如果检测到威胁应该拒绝上传
    And 应该记录恶意文件上传尝试

  @security @encryption
  Scenario: 数据传输加密验证
    Given 系统要求HTTPS加密传输
    When 客户端尝试使用HTTP连接
    Then 连接应该被重定向到HTTPS
    And 敏感数据传输应该使用TLS 1.2+
    And 应该记录不安全连接尝试

  @security @compliance
  Scenario: 合规性监控
    Given 数据保护合规规则已配置
    When 处理用户个人数据
    Then 应该记录数据处理活动
    And 应该确保符合GDPR要求
    And 应该提供数据访问和删除机制

  @security @incident-response
  Scenario: 安全事件响应
    Given 安全事件响应流程已配置
    When 检测到严重安全威胁
    Then 应该自动执行响应措施:
      | action              | description        |
      | block_ip            | 封禁攻击源IP       |
      | alert_admins        | 通知管理员         |
      | log_incident        | 记录安全事件       |
      | backup_logs         | 备份相关日志       |
      | isolate_service     | 必要时隔离服务     |