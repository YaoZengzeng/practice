## Setup kafka develop environment

```bash
kubectl apply -f manifest/zookeeper.yaml
kubectl apply -f manifest/kafka.yaml
```

### producer

Simple golang kafka producer client that could be used to produce messages to kafka cluster.

### consumer

Simple golang kafka consumer client that could be used to consume messages from kafka cluster.
