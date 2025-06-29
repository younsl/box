!/bin/sh
# traffic_distribution_test.sh - Tests Service trafficDistribution Local AZ communication

SERVICE_NAME="<SERVICE>.<NAMESPACE>.svc.cluster.local"
TOTAL_CHECKS=20

echo "=== Traffic Distribution Test ==="
echo "Service: $SERVICE_NAME"
echo "Testing Local AZ communication..."
echo "Press Ctrl+C to stop"
echo ""

# 백그라운드에서 지속적으로 요청 생성
generate_traffic() {
   while true; do
       curl -s http://$SERVICE_NAME/actuator/health > /dev/null 2>&1
       sleep 0.5
   done
} &

TRAFFIC_PID=$!

# 컬럼명 표시 (정렬된 형태)
printf "%-8s %-8s %s\n" "CHECK" "[COUNT]" "IP_ADDRESS"
echo "----------------------------------------"

# 연결 상태 모니터링
for i in $(seq 1 $TOTAL_CHECKS); do
   # netstat로 현재 연결 확인
   if command -v netstat >/dev/null 2>&1; then
       connections=$(netstat -tn 2>/dev/null | grep ":80\|:8080" | \
       awk '{print $5}' | cut -d: -f1 | sort | uniq -c)
   elif command -v ss >/dev/null 2>&1; then
       connections=$(ss -tn 2>/dev/null | grep ":80\|:8080" | \
       awk '{print $4}' | cut -d: -f1 | sort | uniq -c)
   else
       connections=""
   fi

   if [ -n "$connections" ]; then
       echo "$connections" | while read count ip; do
           if [ -n "$count" ] && [ -n "$ip" ]; then
               printf "%-8s %-8s %s\n" "$i/$TOTAL_CHECKS" "[$count]" "$ip"
           fi
       done
   else
       printf "%-8s %-8s %s\n" "$i/$TOTAL_CHECKS" "[0]" "No connections"
   fi

   sleep 2
done

# 트래픽 생성 중지
kill $TRAFFIC_PID 2>/dev/null
echo ""
echo "Monitoring completed."
