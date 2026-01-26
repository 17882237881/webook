# Webook 帖子互动统计（点赞/收藏/阅读量）实现文档

## 目录

1. [功能概述](#功能概述)
2. [设计目标与取舍](#设计目标与取舍)
3. [技术选型引导（为什么用 MQ / Redis）](#技术选型引导为什么用-mq--redis)
4. [架构分层与流程总览](#架构分层与流程总览)
5. [数据模型设计](#数据模型设计)
6. [Redis Key 设计](#redis-key-设计)
7. [RabbitMQ 事件设计](#rabbitmq-事件设计)
8. [核心流程详解](#核心流程详解)
9. [一致性与幂等性策略](#一致性与幂等性策略)
10. [高并发下的风险点与处理](#高并发下的风险点与处理)
11. [部署与配置](#部署与配置)
12. [API 接口说明](#api-接口说明)
13. [常见问题与排查](#常见问题与排查)
14. [可扩展方向](#可扩展方向)
---

## 功能概述

本模块为帖子增加三类互动统计：

- 点赞（Like）
- 收藏（Collect）
- 阅读量（Read / PV）

并支持以下能力：

- 点赞/取消点赞
- 收藏/取消收藏
- 阅读量统计（带短时去重）
- 列表与详情页展示计数和用户状态（是否点赞/收藏）

---

## 设计目标与取舍

### 学习目标

- 学会 RabbitMQ 的基础用法（生产者/消费者/ACK/至少一次投递）
- 学会 Redis 计数、去重、分布式锁的常见使用方式
- 学会 MySQL 关系表的幂等写入与状态切换
- 学会“最终一致性”的落库策略

### 高并发目标

1. **读快**：列表和详情页主要读取 Redis 计数
2. **写稳**：点赞/收藏关系表仍同步写 MySQL，保证幂等
3. **异步统计**：计数变更走 MQ -> Redis -> 批量落库
4. **允许延迟**：计数展示允许延迟 1~5 秒

### 关键取舍

- **计数不直接写数据库**：避免热点行更新压力
- **统计使用 MQ 异步**：将高频写变成异步消息
- **关系表同步写**：确保用户态正确（能查询是否点赞/收藏）
- **最终一致**：计数可能短时偏差，但不影响业务主流程

---

## 技术选型引导（为什么用 MQ / Redis）

这一节专门回答“如果不用 MQ/Redis 会怎样、为什么要用它们”。这也是高并发设计里最核心的取舍依据。

### 为什么要用消息队列（RabbitMQ）

**不用 MQ 会发生什么问题：**

1. **请求链路变长，接口容易超时**  

   点赞/收藏/阅读都要同步更新统计表，如果直接写 MySQL，会把“业务主流程”绑死在统计更新上。一旦数据库写入变慢，用户接口会直接超时。
   例子：用户打开帖子详情时，后台不仅查帖子，还要同步更新统计表。原本 20ms 的请求变成 500ms，前端出现明显卡顿甚至超时。

2. **数据库写放大，热点行锁竞争**  

   阅读量、点赞数是典型热点计数，直接 `UPDATE post_stats SET read_cnt = read_cnt + 1` 会形成高频写热点行，InnoDB 行锁竞争严重，吞吐急剧下降。
   例子：一篇爆款文章每秒 5000 次阅读，数据库不断对同一行做 UPDATE，行锁等待堆积，写入吞吐大幅下降。

3. **流量峰值会拖垮整个业务**  

   没有 MQ，峰值流量直接冲击数据库写能力，容易引发连锁故障（接口变慢 -> 重试增多 -> 进一步放大流量）。
   例子：活动期间阅读 QPS 2 万，但数据库写能力只有 1000/s，大量请求失败，应用日志出现超时与重试雪崩。

**用了 MQ 之后的收益：**

- **异步解耦**：业务请求只负责“产生事件”，统计更新放到后台消费
  例子：点赞接口写完关系表就返回，统计消费者慢一点也不会拖慢用户响应。
- **削峰填谷**：MQ 吸收突发流量，消费端按可控速度处理
  例子：突发 1 万次阅读请求进入队列，消费者每秒处理 1000 条，用户仍能快速返回。
- **可扩展性强**：未来新增“积分”“通知”等系统，只需订阅事件
  例子：新增“通知作者”功能，只需增加一个新消费者，不改点赞接口。
- **可靠性提升**：至少一次投递 + 消费端幂等，降低数据丢失风险
  例子：消费者宕机重启后还能继续消费，重复消息也会被去重，计数不乱。

> 结论：MQ 让“高频统计更新”从同步路径移走，保护核心业务接口的稳定性。

### 为什么要用 Redis

**不用 Redis 会发生什么问题：**

1. **计数查询慢**  

   每次列表/详情都要查 MySQL 统计表或 join，会明显拖慢读接口响应。
   例子：列表页一次展示 20 篇文章，如果每篇都查一次统计表，就会产生 20 次数据库查询，接口响应明显变慢。

2. **计数更新慢**  

   高频写 MySQL 不仅慢，还会产生锁竞争，尤其是阅读量写入。
   例子：阅读量每次都 UPDATE 数据库，QPS 1 万时数据库 CPU 飙升，最终导致整体吞吐下降。

3. **去重/幂等成本高**  

   阅读量去重如果用 MySQL，需要额外表或复杂查询；点赞/收藏的状态判断也会频繁 hit DB。
   例子：为防刷阅读量，需要建一张阅读记录表并查询/插入，每天可能产生千万级记录，查询和存储成本都很高。

**用了 Redis 之后的收益：**

- **高并发写入**：`HINCRBY` 是原子操作，支持极高 QPS
  例子：每秒几万次阅读量累加也能稳定写入，不会出现并发丢失。
- **高并发读取**：计数从缓存直接读，响应快
  例子：列表页 20 篇文章的计数一次 Redis 管道即可拿到，接口延迟很低。
- **去重更简单**：`SETNX + TTL` 轻松实现“30 秒内不重复计数”
  例子：用户 10 秒内刷新 5 次，只会被计 1 次阅读。
- **批量刷库**：Dirty Set + 定时任务，减少数据库写压力
  例子：1 万次阅读只触发几次批量刷库，比逐条 UPDATE 轻得多。

> 结论：Redis 负责“高频写”和“快速读”，是统计系统的性能核心。

### 为什么要“Redis + MQ”一起用？

如果只用 Redis 而不用 MQ：  
- 业务接口仍要同步更新 Redis，接口仍可能被阻塞
  例子：Redis 短暂抖动时，所有点赞请求都会失败，用户体验受影响。

- 统计逻辑耦合在请求路径上，不易扩展
  例子：想在点赞时新增“通知作者”，需要在接口里再调用通知服务，链路更长、失败点更多。


如果只用 MQ 而不用 Redis：  
- 消费端只能直接写数据库，无法解决热点行写入瓶颈
  例子：消费端收到大量阅读事件仍要 UPDATE 同一行统计数据，热点锁竞争问题依旧存在。


**组合方案的价值：**

```
业务请求 -> MQ（异步） -> Redis（高性能计数） -> MySQL（批量落库）
```

这样既能**异步解耦**，又能**高并发计数**，同时保留持久化保障。

---

## 架构分层与流程总览

```
HTTP Handler
  -> Application Service
    -> Repository (MySQL)
    -> Cache (Redis)
    -> MQ Publisher (RabbitMQ)

RabbitMQ Consumer
  -> Redis 计数
  -> Dirty Set

Flush Worker (定时任务)
  -> Redis 批量读取计数
  -> MySQL upsert
```

分层位置：

- HTTP: `internal/adapters/inbound/http/post.go`
- Service: `internal/application/post_stats.go`
- MQ: `internal/adapters/outbound/mq/*`
- Redis: `internal/adapters/outbound/persistence/redis/post_stats.go`
- MySQL: `internal/adapters/outbound/persistence/mysql/post_stats.go`
- 定时刷盘：`internal/application/post_stats_flusher.go`

---

## 数据模型设计

### 1) 统计表：`post_stats`

| 字段 | 类型 | 说明 |
|------|------|------|
| post_id | bigint | 主键，帖子ID |
| like_cnt | bigint | 点赞数 |
| collect_cnt | bigint | 收藏数 |
| read_cnt | bigint | 阅读数 |
| ctime | bigint | 创建时间（毫秒） |
| utime | bigint | 更新时间（毫秒） |

**设计理由**：
- 避免在 `posts` 表上频繁更新，降低锁竞争
- 单独做统计可按需扩展（后续加评论数等）

### 2) 点赞关系表：`post_like_rel`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint | 自增主键 |
| post_id | bigint | 帖子ID |
| user_id | bigint | 用户ID |
| status | tinyint | 1=点赞 0=取消 |
| ctime | bigint | 创建时间 |
| utime | bigint | 更新时间 |

**关键点**：`(post_id, user_id)` 唯一索引

**设计理由**：
- 用 `status` 做软切换，避免频繁插删
- 唯一索引保证幂等

### 3) 收藏关系表：`post_collect_rel`

结构同点赞表，仅语义不同。

---

## Redis Key 设计

| Key | 类型 | 说明 |
|-----|------|------|
| `post:stats:{postId}` | Hash | 存储 `like_cnt/collect_cnt/read_cnt` |
| `post:stats:dirty` | Set | 记录需要刷库的 postId |
| `post:read:dedupe:uid:{userId}:pid:{postId}` | String | 登录用户阅读去重 |
| `post:read:dedupe:anon:{hash}:pid:{postId}` | String | 匿名阅读去重 |
| `post:stats:event:{eventId}` | String | MQ 事件去重 |
| `post:stats:flush:lock` | String | 刷库分布式锁 |

**设计理由**：
- Hash 适合聚合计数
- Set 适合去重和脏数据集合
- SetNX 适合去重和锁

---

## RabbitMQ 事件设计

### 交换机 / 队列

- Exchange: `post.stats.exchange`
- Queue: `post.stats.queue`
- RoutingKey: `post.stats`
- 类型: direct

### 消息结构（JSON）

```
{
  "event_id": "uuid",
  "type": "like|unlike|collect|uncollect|read",
  "post_id": 123,
  "user_id": 456,
  "ts": 1730000000
}
```

**为什么这样设计**：
- `event_id` 用于幂等（消费端去重）
- `type` 区分行为
- `post_id` 是计数的主维度

---

## 核心流程详解

### 1) 点赞/取消点赞

**写路径：**

1. API 进入 `PostInteractionService.Like/Unlike`
2. MySQL 写关系表（幂等：唯一索引 + 状态切换）
3. 如果状态发生变化，发布 MQ 事件
4. MQ 消费者更新 Redis 计数 + dirty 集合
5. 定时刷库把 Redis 计数同步到 MySQL

**为什么这样做**：
- 用户态必须立即正确 -> MySQL 关系表同步写
- 计数延迟无所谓 -> 异步统计

### 2) 收藏/取消收藏

流程与点赞相同。

### 3) 阅读量

**写路径：**

1. API 进入 `Read`
2. Redis `SETNX` 做 30 秒去重
3. 通过后发布 MQ `read` 事件
4. 消费端自增 Redis `read_cnt`
5. dirty set 等待刷库

**为什么这样做**：
- 阅读量写频率最高，必须用 Redis 抗高并发
- 30 秒去重防止刷接口造成暴涨

---

## 一致性与幂等性策略

### 幂等策略

1. **点赞/收藏**：
   - `(post_id, user_id)` 唯一索引
   - `status` 字段确保多次请求不会重复计数

2. **MQ 消费**：
   - Redis `post:stats:event:{eventId}` 做幂等
   - 同一个事件只会处理一次

3. **阅读去重**：
   - Redis `SETNX` + TTL 30 秒

### 最终一致

- 计数更新：Redis 立即变化，MySQL 延迟刷库
- 展示层：默认读 Redis
- 允许 1~5 秒延迟

---

## 高并发下的风险点与处理

### 1) 热点行更新

**问题**：大量点赞/阅读写 MySQL 会锁表/锁行

**处理**：所有计数更新改为 Redis，再批量落库

### 2) MQ 重复消费

**问题**：RabbitMQ 至少一次投递，可能重复

**处理**：消费端 Redis `SETNX` 去重

### 3) MQ 异常导致计数缺失

**问题**：MySQL 写成功但 MQ 发布失败

**处理**：当前实现未引入 Outbox（生产场景需改造）

**为什么不做 Outbox**：
- 学习阶段优先理解 MQ 基本模型
- Outbox 属于下一阶段优化点

### 4) 多实例刷库冲突

**问题**：多个服务同时刷库会重复写

**处理**：Redis `TryLock` 做分布式锁

### 5) 计数变成负数

**可能原因**：
- 多消费者乱序（like/unlike 顺序被打乱）

**当前策略**：
- 单消费者 + 顺序消息可避免
- 如果要严格避免，需在消费端 clamp >=0 或用数据库对账

---

## 部署与配置

### Docker Compose

新增 RabbitMQ 服务：`deploy/docker/docker-compose.yaml`

```
services:
  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - 5672:5672
      - 15672:15672
```

访问管理台：`http://localhost:15672` (guest/guest)

### 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| AMQP_URL | amqp://guest:guest@localhost:5672/ | RabbitMQ 地址 |
| MQ_EXCHANGE | post.stats.exchange | 交换机 |
| MQ_QUEUE | post.stats.queue | 队列 |
| MQ_ROUTING_KEY | post.stats | 路由键 |
| MQ_PREFETCH | 50 | 消费者预取数 |

---

## API 接口说明

### 点赞 / 取消点赞

- `POST /posts/:id/like`
- `POST /posts/:id/unlike`

### 收藏 / 取消收藏

- `POST /posts/:id/collect`
- `POST /posts/:id/uncollect`

### 阅读计数

- `POST /posts/:id/read`
- `GET /posts/:id` 会自动触发阅读计数

### 返回数据字段（列表/详情）

```
likeCnt
collectCnt
readCnt
liked
collected
```

---

## 常见问题与排查

### 1) RabbitMQ 连接失败

**现象**：启动 panic

**原因**：RabbitMQ 未启动或 AMQP_URL 配置错误

**排查**：
- 查看 `docker ps` 是否有 rabbitmq
- 检查 `AMQP_URL`

### 2) 计数不更新

**原因**：消费者未启动或消息未投递

**排查**：
- 管理台查看 queue 是否有堆积
- 检查日志：consumer 是否收到消息

### 3) 计数刷新不到 MySQL

**原因**：flusher 未启动 or Redis 锁竞争

**排查**：
- 确认 `InitPostStatsWorker` 是否启动
- Redis 查看 `post:stats:dirty` 是否有数据

### 4) go mod 缺依赖

**现象**：构建报错 `amqp091-go` 不存在

**解决**：
- `go mod tidy`

### 5) 读接口需要登录

**现象**：未登录访问 `/posts` 或 `/posts/:id` 返回 401

**原因**：JWT 中间件默认保护所有路径，仅放行 `/users`、`/users/login` 等少量接口

**解决**：
- 如果希望“未登录可阅读”，在 `internal/adapters/inbound/http/middleware/jwt_middleware.go` 的 `IgnorePaths` 中放行 GET `/posts` 和 `/posts/:id`

---

## 可扩展方向

1. **Outbox 模式**：保证 DB 与 MQ 一致性
2. **DLQ 死信队列**：失败消息不丢失
3. **批量消息**：读取量更大时可合并事件
4. **监控告警**：消费堆积、刷库延迟监控
5. **数据对账**：定期用关系表回算计数

---

## 学习要点总结

- RabbitMQ 基础：exchange/queue/routing-key/prefetch/ack
- Redis 去重：SETNX + TTL
- 计数聚合：Hash + HINCRBY
- 最终一致：缓存优先 + 异步刷库
- 幂等设计：唯一索引 + 状态字段 + 事件去重

---

> 文档对应实现文件：
> - `internal/application/post_stats.go`
> - `internal/adapters/outbound/mq/*`
> - `internal/adapters/outbound/persistence/redis/post_stats.go`
> - `internal/adapters/outbound/persistence/mysql/post_stats.go`
> - `internal/adapters/inbound/http/post.go`
> - `deploy/docker/docker-compose.yaml`
