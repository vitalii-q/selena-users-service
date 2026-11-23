import argparse
import concurrent.futures
import http.client
import math
import time
from urllib.parse import urlsplit


def request_once(target, timeout):
    started = time.perf_counter()
    connection = None
    try:
        connection = http.client.HTTPConnection(target.hostname, target.port, timeout=timeout)
        path = target.path or "/"
        if target.query:
            path = f"{path}?{target.query}"
        connection.request("GET", path, headers={"Connection": "close", "Accept": "application/json"})
        response = connection.getresponse()
        response.read()
        elapsed_ms = (time.perf_counter() - started) * 1000
        return response.status, elapsed_ms, None
    except Exception as error:
        elapsed_ms = (time.perf_counter() - started) * 1000
        return None, elapsed_ms, type(error).__name__
    finally:
        if connection is not None:
            connection.close()


def percentile(values, fraction):
    ordered = sorted(values)
    if not ordered:
        return float("nan")
    index = max(0, math.ceil(fraction * len(ordered)) - 1)
    return ordered[index]


def main():
    parser = argparse.ArgumentParser(description="Concurrent read-only HTTP GET load test")
    parser.add_argument("--url", required=True)
    parser.add_argument("--requests", type=int, default=200)
    parser.add_argument("--concurrency", type=int, default=10)
    parser.add_argument("--timeout", type=float, default=5.0)
    args = parser.parse_args()
    target = urlsplit(args.url)
    if target.scheme != "http" or not target.hostname or not target.port:
        parser.error("--url must be an HTTP URL with an explicit port")
    if args.requests < 1 or args.concurrency < 1:
        parser.error("--requests and --concurrency must be positive")

    wall_started = time.perf_counter()
    with concurrent.futures.ThreadPoolExecutor(max_workers=args.concurrency) as executor:
        results = list(executor.map(lambda _: request_once(target, args.timeout), range(args.requests)))
    total_seconds = time.perf_counter() - wall_started

    successful_latencies = [elapsed for status, elapsed, _ in results if status == 200]
    http_errors = sum(status is not None and status != 200 for status, _, _ in results)
    transport_errors = sum(error is not None for _, _, error in results)
    successful = len(successful_latencies)

    print(f"url_path={target.path or '/'}")
    print(f"requests={args.requests} concurrency={args.concurrency} timeout_seconds={args.timeout:g}")
    print(f"success_200={successful} http_errors={http_errors} transport_errors={transport_errors}")
    print(f"total_seconds={total_seconds:.3f} requests_per_second={args.requests / total_seconds:.2f}")
    if successful_latencies:
        print(
            "latency_ms "
            f"min={min(successful_latencies):.2f} "
            f"p50={percentile(successful_latencies, 0.50):.2f} "
            f"p95={percentile(successful_latencies, 0.95):.2f} "
            f"p99={percentile(successful_latencies, 0.99):.2f} "
            f"max={max(successful_latencies):.2f}"
        )
    raise SystemExit(0 if successful == args.requests else 1)


if __name__ == "__main__":
    main()
