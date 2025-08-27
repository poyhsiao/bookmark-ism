# Language: zh-CN
# 功能: 应用性能监控和可观测性
# 作为系统管理员，我需要全面的应用监控，以确保系统性能和健康状态

Feature: 应用性能监控和可观测性
  为了确保系统稳定运行和快速问题诊断
  作为系统管理员
  我需要全面的应用监控和可观测性功能

  Background:
    Given 监控系统已经部署并运行
    And Prometheus服务器正在收集metrics
    And Grafana dashboard已经配置

  @monitoring @metrics
  Scenario: Prometheus metrics收集
    Given 应用服务正在运行
    When 我访问"/metrics"端点
    Then 我应该看到Prometheus格式的metrics
    And metrics应该包含以下指标:
      | metric_name              | metric_type | description          |
      | http_requests_total      | counter     | HTTP请求总数          |
      | http_request_duration    | histogram   | HTTP请求响应时间      |
      | active_connections       | gauge       | 当前活跃连接数        |
      | database_connections     | gauge       | 数据库连接池状态      |
      | redis_connections        | gauge       | Redis连接状态         |
      | go_memstats_alloc_bytes  | gauge       | Go内存分配统计        |

  @monitoring @health-check
  Scenario: 服务健康检查端点
    Given 所有依赖服务正常运行
    When 我访问"/health"端点
    Then 我应该收到HTTP 200响应
    And 响应应该包含:
      """
      {
        "status": "healthy",
        "timestamp": "*",
        "services": {
          "database": {
            "status": "healthy",
            "response_time": "*"
          },
          "redis": {
            "status": "healthy",
            "response_time": "*"
          },
          "storage": {
            "status": "healthy",
            "response_time": "*"
          }
        }
      }
      """

  @monitoring @health-check @failure
  Scenario: 依赖服务故障时的健康检查
    Given 数据库服务不可用
    When 我访问"/health"端点
    Then 我应该收到HTTP 503响应
    And 响应应该包含:
      """
      {
        "status": "unhealthy",
        "timestamp": "*",
        "services": {
          "database": {
            "status": "unhealthy",
            "error": "*"
          }
        }
      }
      """

  @monitoring @performance
  Scenario: 应用性能指标监控
    Given 应用正在处理用户请求
    When 系统运行30秒
    Then Prometheus应该记录以下性能指标:
      | metric                   | condition        |
      | http_request_duration_p95| < 200ms         |
      | http_request_duration_p99| < 500ms         |
      | active_connections       | > 0             |
      | go_memstats_heap_inuse   | < 100MB         |
      | database_query_duration  | < 50ms (avg)    |

  @monitoring @alerts
  Scenario: 性能告警触发
    Given Grafana告警规则已配置
    When 应用响应时间超过1秒持续2分钟
    Then 应该触发"高延迟"告警
    And 告警应该发送到通知渠道
    And 告警应该包含详细的错误信息

  @monitoring @dashboard
  Scenario: Grafana dashboard显示实时指标
    Given Grafana dashboard已加载
    When 我查看"Bookmark Service Overview"dashboard
    Then 我应该看到以下面板:
      | panel_name           | data_source | metric                |
      | HTTP请求率           | Prometheus  | http_requests_total   |
      | 响应时间分布         | Prometheus  | http_request_duration |
      | 错误率               | Prometheus  | http_requests_total   |
      | 活跃连接数           | Prometheus  | active_connections    |
      | 数据库性能           | Prometheus  | database_*            |
      | Go运行时指标         | Prometheus  | go_*                  |

  @monitoring @logs
  Scenario: 结构化日志聚合
    Given 日志聚合系统已配置
    When 应用生成日志事件
    Then 日志应该以JSON格式结构化
    And 日志应该包含以下字段:
      | field_name   | description      |
      | timestamp    | 时间戳           |
      | level        | 日志级别         |
      | message      | 日志消息         |
      | service      | 服务名称         |
      | request_id   | 请求ID           |
      | user_id      | 用户ID (如适用)  |
      | trace_id     | 追踪ID           |

  @monitoring @tracing
  Scenario: 分布式追踪
    Given 分布式追踪已启用
    When 用户创建一个书签
    Then 应该生成完整的追踪链路
    And 追踪应该包含以下span:
      | service_name      | operation_name     |
      | api-server        | POST /bookmarks    |
      | bookmark-service  | CreateBookmark     |
      | database          | INSERT bookmark    |
      | search-service    | IndexBookmark      |
      | storage-service   | CaptureScreenshot  |

  @monitoring @resource-usage
  Scenario: 资源使用监控
    Given 系统正在正常运行
    When 我检查资源使用情况
    Then CPU使用率应该低于80%
    And 内存使用率应该低于85%
    And 磁盘使用率应该低于90%
    And 网络连接数应该在正常范围内

  @monitoring @business-metrics
  Scenario: 业务指标监控
    Given 应用正在服务用户请求
    When 我查看业务指标
    Then 应该显示以下业务数据:
      | metric_name           | description          |
      | total_users           | 总用户数             |
      | active_users_24h      | 24小时活跃用户       |
      | bookmarks_created_24h | 24小时创建书签数     |
      | sync_operations_24h   | 24小时同步操作数     |
      | search_queries_24h    | 24小时搜索查询数     |