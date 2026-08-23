# EnvMonitor

EnvMonitor 是环境水质在线监测平台。流域下的监测站点按周期采样，读数以文件
形式持久化；平台按指标阈值判定超标并生成告警，告警上报下游数据中心并可
确认；数据按月聚合生成月报，站点采样受容量配额约束，采样与告警事件全部
留审计。

## 构建

```bash
docker build -t envmonitor .
```

## 运行

```bash
ENVMONITOR_DATA_DIR=./data ENVMONITOR_ADDR=127.0.0.1:7790 go run ./cmd/envmonitor
```

启动后打开 http://127.0.0.1:7790 进入控制台，页面包含站点、读数、告警和
审计四个视图，JSON API 位于 /api 下。

## 数据目录

流域、站点、读数、超标记录、确认记录、月报汇总、告警、月报窗口、中心
outbox、配额账本与审计日志都以文件形式保存在数据目录下，可以直接整体
备份。
