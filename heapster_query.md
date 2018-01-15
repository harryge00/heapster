## heapster查询
curl -G http://localhost:8082/api/v1/model/namespaces/default/pods/debian/metrics/network/rx

curl -G http://localhost:8082/api/v1/model/namespaces/default/pods/debian/containers/testcontainer/metrics/network/rx

curl -G http://localhost:8082/api/v1/model/namespaces/default/pods/debian/metrics/memory/usage

无状态应用的metrics查询 api 为：
```
GET http://localhost:8082/api/v1/model/namespaces/default/rcs/{replicationcontroller-example}/metrics/{metric-name}
```
有状态应用的metrics查询 api 为：
```
GET http://localhost:8082/api/v1/model/namespaces/default/rcs/{replicationcontroller-example}/metrics/{metric-name}
```

1. memory-usage:
```
curl -G http://localhost:8082/api/v1/model/namespaces/default/statefulsets/statefulset-example/metrics/memory/usage
curl -G http://localhost:8082/api/v1/model/namespaces/default/rcs/replicationcontroller-example/metrics/memory/usage
curl -G http://localhost:8082/api/v1/model/namespaces/default/pods/replicationcontroller-example-ft3jr/metrics/memory/usage
curl -G http://localhost:8082/api/v1/model/namespaces/default/pods/replicationcontroller-example-jx9pp/metrics/memory/usage
```

2. uptime:
```
curl -G http://localhost:8082/api/v1/model/namespaces/default/statefulsets/statefulset-example/metrics/uptime;
curl -G http://localhost:8082/api/v1/model/namespaces/default/rcs/replicationcontroller-example/metrics/uptime;
curl -G http://localhost:8082/api/v1/model/namespaces/default/pods/replicationcontroller-example-ft3jr/metrics/uptime;
curl -G http://localhost:8082/api/v1/model/namespaces/default/pods/replicationcontroller-example-jx9pp/metrics/uptime;
```

3. cpu-usage:
```
curl -G http://localhost:8082/api/v1/model/namespaces/default/statefulsets/statefulset-example/metrics/cpu/usage;
curl -G http://localhost:8082/api/v1/model/namespaces/default/rcs/replicationcontroller-example/metrics/cpu/usage;
curl -G http://localhost:8082/api/v1/model/namespaces/default/pods/replicationcontroller-example-ft3jr/metrics/cpu/usage;
curl -G http://localhost:8082/api/v1/model/namespaces/default/pods/replicationcontroller-example-jx9pp/metrics/cpu/usage;
```

4. network-rx:
```
curl -G http://localhost:8082/api/v1/model/namespaces/default/statefulsets/statefulset-example/metrics/network/rx;
curl -G http://localhost:8082/api/v1/model/namespaces/default/rcs/replicationcontroller-example/metrics/network/rx;
curl -G http://localhost:8082/api/v1/model/namespaces/default/pods/replicationcontroller-example-ft3jr/metrics/network/rx;
curl -G http://localhost:8082/api/v1/model/namespaces/default/pods/replicationcontroller-example-jx9pp/metrics/network/rx;
```

5. filesystem-usage:
```
curl -G http://localhost:8082/api/v1/model/namespaces/default/statefulsets/statefulset-example/metrics/filesystem/usage;
curl -G http://localhost:8082/api/v1/model/namespaces/default/rcs/replicationcontroller-example/metrics/filesystem/usage;
curl -G http://localhost:8082/api/v1/model/namespaces/default/pods/replicationcontroller-example-ft3jr/metrics/filesystem/usage;
curl -G http://localhost:8082/api/v1/model/namespaces/default/pods/replicationcontroller-example-jx9pp/metrics/filesystem/usage;
```

6. cpu/usage_rate:
```
curl -G http://localhost:8082/api/v1/model/namespaces/default/statefulsets/statefulset-example/metrics/cpu/usage_rate;
curl -G http://localhost:8082/api/v1/model/namespaces/default/rcs/replicationcontroller-example/metrics/cpu/usage_rate;
curl -G http://localhost:8082/api/v1/model/namespaces/default/pods/replicationcontroller-example-ft3jr/metrics/cpu/usage_rate;
curl -G http://localhost:8082/api/v1/model/namespaces/default/pods/replicationcontroller-example-jx9pp/metrics/cpu/usage_rate;
curl -G http://10.252.10.4:8082/api/v1/model/namespaces/wangzhuzhen-2/pods/wong0829-v0492/metrics/memory/usage
```

### heapster调用kubelet API命令：
```
curl -XPOST -H "Content-Type:application/json" -d '{"containerName":"/","num_stats":1,"start":"2017-09-13T10:10:00+08:00","end":"2017-09-13T21:31:00+08:00","subcontainers":true}' http://10.30.21.136:10255/stats/container/
```

**通过 heapster 查询容器性能数据**
```
curl -G http://172.25.3.194:8082/api/v1/model/namespaces/zhangyuandao-24/pods/vsl21-flnm2/metrics/memorusage
```

**获取cluster支持的metrics**
`curl -L http://<heapster-IP>:8082/api/v1/model/metrics`

**列出Nodes支持的metrics**
`curl -L http://<heapster-IP>:8082/api/v1/model/nodes/metrics`

**查看对应Pod的cpu使用率**
```
curl -L http://<heapster-IP>:8082/api/v1/model/namespaces/<namespace-name>/pods/<pod-name>/metrics/cpu-usage
curl -L http://10.30.100.9:8182/api/v1/model/namespaces/zhangneng-72/pods/node-swnhm/metrics/cpu-usage
```

## Influxdb 查询

1. 获取对应node的信息：
```
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"cpu/limit\" WHERE \"nodename\"='172.25.3.194'"
```

2. 获取对应Pod的信息：
```
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"cpu/usage\" WHERE \"pod_name\"='replicationcontroller-example-1sx60'"
```

3. 获取对应容器的信息：
```
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"cpu/usage\" WHERE \"container_name\"='testcontainer'"
```

4. 获取RC的信息：
```
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"cpu/usage\" WHERE \"rc_name\"='replicationcontroller-example' and time > now() - 1m"
```

5. 获取statefulset的信息
```
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"cpu/usage\" WHERE \"ss_name\"='statefulset-example' and time > now() - 5m"
```

6. 其他：
```
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"cpu/usage\" WHERE \"ss_name\"='statefulset-example' and time > now() - 5m";
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"cpu/usage\" WHERE \"rc_name\"='replicationcontroller-example' and time > now() - 5m";
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"cpu/usage\" WHERE \"pod_name\"='replicationcontroller-example-ft3jr' and time > now() - 5m";
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"cpu/usage\" WHERE \"pod_name\"='replicationcontroller-example-jx9pp' and time > now() - 5m";
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"cpu/usage\" WHERE \"container_name\"='testcontainer' and time > now() - 1m"
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"cpu/usage\" WHERE \"pod_name\"='debian' and time > now() - 1m"
```

### Influxdb具体指标查询

1. filesystem-usage:
```
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"filesystem/usage\" WHERE \"ss_name\"='statefulset-example' and time > now() - 5m";
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"filesystem/usage\" WHERE \"rc_name\"='replicationcontroller-example' and time > now() - 5m";
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"filesystem/usage\" WHERE \"pod_name\"='replicationcontroller-example-ft3jr' and time > now() - 5m";
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"filesystem/usage\" WHERE \"pod_name\"='replicationcontroller-example-jx9pp' and time > now() - 5m";
```

2. cpu/usage:
```
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"cpu/usage\" WHERE \"pod_name\"='replicationcontroller-example-jx9pp' and time > now() - 5m";
```

3. cpu/usage_rate:
```
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"cpu/usage_rate\" WHERE \"pod_name\"='replicationcontroller-example-jx9pp' and time > now() - 5m";
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"cpu/usage_rate\" WHERE \"pod_name\"='replicationcontroller-example-8mf50' and \"pod_namespace\"='default' and  time > now() - 10m LIMIT 300";
```

4. Network tx
```
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"network/rx_rate\",\"network/tx_rate\" WHERE \"pod_name\"='ttttt-02z9b' and \"pod_namespace\"='zhangyuandao-testnamespace' and  time > now() - 3h LIMIT 300"
```

5. memory usage
```
curl -G 'http://172.25.3.194:8086/query?pretty=true' --data-urlencode "db=k8s" --data-urlencode "q=SELECT \"value\" FROM \"memory/usage\" WHERE \"pod_name\"='ttttt-02z9b' and \"pod_namespace\"='zhangyuandao-testnamespace' and  time > now() - 10m LIMIT 300"
```
