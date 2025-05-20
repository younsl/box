#!/usr/bin/env bash

SCRIPT_PATH="$(pwd)/$(basename $0)"

case $1 in
    'start')
        SCHEDULE='*/10 * * * *'
        
        if crontab -l 2>/dev/null | grep -q "$SCRIPT_PATH"; then
            echo "EIP rotator (${SCRIPT_PATH}) is already scheduled in crontab"
        else
            (crontab -l 2>/dev/null; echo "$SCHEDULE $SCRIPT_PATH") | crontab -
            echo "Scheduled EIP rotator (${SCRIPT_PATH}) to run every ten minutes"
        fi
        ;;
    'stop')
        crontab -l | grep -v $(basename $0) | crontab -
        echo "Stopped EIP rotator (${SCRIPT_PATH}) from crontab"
        ;;
    *)
        REGION=$(curl -s http://169.254.169.254/latest/meta-data/placement/region)
        INSTANCE_ID=$(curl -s http://169.254.169.254/latest/meta-data/instance-id)
        OLD_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)
        OLD_ALLOCATION=$(aws ec2 describe-addresses --public-ips ${OLD_IP} --query Addresses[0].AllocationId --output text)
        NEW_IP=$(aws ec2 allocate-address --query PublicIp --output text)

        echo ''
        echo "[~] Change EIP address from ${OLD_IP} to ${NEW_IP}"
        echo ''
        echo "[+] Associating new EIP ${NEW_IP} to instance ${INSTANCE_ID} ..."
        aws ec2 associate-address --instance-id ${INSTANCE_ID} --public-ip ${NEW_IP} --region ${REGION}
        echo "[-] Releasing old EIP ${OLD_IP} ..."
        aws ec2 release-address --allocation-id ${OLD_ALLOCATION} --region ${REGION}
esac