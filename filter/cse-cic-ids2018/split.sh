#!/bin/bash
mkdir -p filters
mkdir -p splits
for file in filters/*; do
    if [ -f "$file" ]; then
        NAME=$(basename $file)
        echo "Start splitting: $file"; tshark -q -r all-merged.pcap -F libpcap -M 50000 -Y "$(cat $file | tr '\n' ' ' | tr '\"' ' ')" -w splits/$NAME.pcap --log-level critical 2>&1 | grep -v "resetting session." &
    fi
done

BENIGN_FILTERS=()

for file in filters/*; do
    FILTER=$(cat $file | tr '\n' ' ' | tr '\"' ' ')
    BENIGN_FILTERS+=("!(${FILTER})")
done

BENIGN_FILTER=$(printf " && %s" "${BENIGN_FILTERS[@]}")

echo ${BENIGN_FILTER:4}
echo "Start splitting: benign"; tshark -r all-merged.pcap -F libpcap -M 50000 -Y "${BENIGN_FILTER:4}" --log-level critical -w splits/benign.pcap 2>&1 | grep -v "resetting session." &

FAIL=0

for job in `jobs -p`
do
    wait $job || let "FAIL+=1"
done

echo "Failed: $FAIL"
