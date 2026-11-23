# Локальный read-only нагрузочный тест Users

## Сценарий

Тестируется только `GET /api/v1/users`. Обработчик читает список пользователей из БД; операции создания, изменения и удаления не вызываются. Перед прогоном проверьте HTTP 200 и число записей, не выводя содержимое ответа. Для текущей локальной БД ранее подтверждено 49 тестовых записей.

Генератор запускается внутри кластера и обращается к `users-service.selena-dev.svc.cluster.local:9065`, поэтому запросы проходят через DNS и ClusterIP Service и могут распределяться между репликами. Каждая операция — отдельный HTTP GET с закрытием соединения после ответа. Ответ читается в память генератора и не печатается.

## Повторение

Команды выполнять из корня репозитория `users-service`. ConfigMap со скриптом создаётся один раз; перед повторным запуском завершённый Pod удаляется.

```sh
kubectl --context=selena-local -n selena-dev create configmap users-service-load-test-script --from-file=loadtest.py=k8s/local/load-test/loadtest.py
kubectl --context=selena-local -n selena-dev scale deployment/users-service --replicas=1
kubectl --context=selena-local -n selena-dev rollout status deployment/users-service
kubectl --context=selena-local apply -f k8s/local/load-test/pod.yaml
kubectl --context=selena-local -n selena-dev wait --for=jsonpath='{.status.phase}'=Succeeded pod/users-service-load-test --timeout=180s
kubectl --context=selena-local -n selena-dev logs pod/users-service-load-test
kubectl --context=selena-local -n selena-dev delete pod users-service-load-test

kubectl --context=selena-local -n selena-dev scale deployment/users-service --replicas=2
kubectl --context=selena-local -n selena-dev rollout status deployment/users-service
kubectl --context=selena-local apply -f k8s/local/load-test/pod.yaml
kubectl --context=selena-local -n selena-dev wait --for=jsonpath='{.status.phase}'=Succeeded pod/users-service-load-test --timeout=180s
kubectl --context=selena-local -n selena-dev logs pod/users-service-load-test
kubectl --context=selena-local -n selena-dev delete pod users-service-load-test
kubectl --context=selena-local -n selena-dev delete configmap users-service-load-test-script
```

Генератор по умолчанию выполняет 200 GET-запросов при concurrency 10 с timeout 5 секунд. Он выводит число HTTP 200, HTTP/transport errors, общее время, requests/sec и min/p50/p95/p99/max latency. HTTP-тело и персональные данные не выводятся.

## Результаты прогона 2026-11-22

| Ready replicas | Requests | Concurrency | HTTP 200 | Errors | RPS | p50 | p95 | p99 | Max |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 1 | 200 | 10 | 200 | 0 | 118.72 | 76.79 ms | 198.03 ms | 300.68 ms | 390.44 ms |
| 2 | 200 | 10 | 200 | 0 | 142.40 | 35.29 ms | 284.93 ms | 553.52 ms | 635.00 ms |

При двух репликах измеренный RPS в этом единственном прогоне был выше примерно на 20%, но p95 и p99 задержки также были выше. Одного короткого прогона недостаточно для вывода о стабильном выигрыше от масштабирования.

Metrics Server в локальном кластере недоступен (`kubectl top pods` возвращает `Metrics API not available`), поэтому CPU/memory Pod не измерялись. Это один короткий тест на MacBook/Minikube с существующей внешней Docker PostgreSQL; приложение использует `linux/amd64` image на ARM64 узле. Результаты не являются прогнозом AWS или облачной производительности.
