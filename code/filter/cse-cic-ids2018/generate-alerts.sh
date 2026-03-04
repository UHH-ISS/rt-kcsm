for file in splits/*; do
    if [ -f "$file" ]; then
        NAME=$(basename $file)
        suricata -r $file --set outputs.1.eve-log.filename=/dev/stdout --set logging.outputs.0.console.enabled=no | jq -c '.label += "'"$file"'"' > "alerts/$NAME.json" &
    fi
done

FAIL=0

for job in `jobs -p`
do
    wait $job || let "FAIL+=1"
done
