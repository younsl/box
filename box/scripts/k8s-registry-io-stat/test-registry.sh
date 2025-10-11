#!/bin/bash

success_count=0
failure_count=0

request_count=${REQUEST_COUNT:-10}

for i in $(seq 1 $request_count)
do
  code=$(curl -o /dev/null -s -w "%{http_code}" https://registry.k8s.io/v2/)
  echo "Request #$i: https://registry.k8s.io/v2/ - Response Code: $code"
  
  if [ "$code" -eq 200 ]; then
    success_count=$((success_count + 1))
  else
    failure_count=$((failure_count + 1))
  fi
done

succ_rate=$(echo "scale=2; $success_count / $request_count * 100" | bc)
echo "Success: $success_count, Failure: $failure_count (Success Rate: $succ_rate%)"