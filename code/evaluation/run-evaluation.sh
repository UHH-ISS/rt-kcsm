bash code/evaluation/decompress.sh
bash code/rtkcsm/build.sh

function custom_weights_measure_ranking {
    rtkcsm --file $1 --reader $2 --profile graph-ranking=$4/graph-ranking.csv --profile-graph-ranking-id $3 --export $4/graphs.json
}

function equal_weights_measure_ranking {
    rtkcsm --file $1 --reader $2 --stage-weight incoming=0.25 same-zone=0.25 different-zone=0.25 outgoing=0.25 --profile graph-ranking=$4/graph-ranking.csv --profile-graph-ranking-id $3 --export $4/graphs.json
}

function measure_performance {
    rtkcsm --file $1 --reader $2 --profile-log-resolution $3 --profile memory="$4/memory.csv"
    rtkcsm --file $1 --reader $2 --profile-log-resolution $3 --profile alerts="$4/alerts.csv"
}

# CSE-CIC-IDS2018
echo "CSE-CIC-IDS2018"
ALERT_FILE="data/alerts/ids2018/eve.json"

RESULT_DIR="results/alerts/ids2018"
mkdir -p ${RESULT_DIR}
measure_performance "${ALERT_FILE}" "suricata" "1000" "${RESULT_DIR}"

RESULT_DIR="results/alerts/ids2018/custom-weights"
mkdir -p ${RESULT_DIR}
custom_weights_measure_ranking "${ALERT_FILE}" "suricata" "5486" "${RESULT_DIR}"

RESULT_DIR="results/alerts/ids2018/equal-weights"
mkdir -p ${RESULT_DIR}
equal_weights_measure_ranking "${ALERT_FILE=}" "suricata" "5486" "${RESULT_DIR}"

echo "----------------------------"
# CIC-IDS2018-APT
echo "CSE-CIC-IDS2018-APT"
ALERT_FILE="data/alerts/ids2018-apt/notice.json"

RESULT_DIR="results/alerts/ids2018-apt"
mkdir -p ${RESULT_DIR}
measure_performance "${ALERT_FILE}" "zeek" "1000" "${RESULT_DIR}"

# Graph id = 1 is the one containing the APT attack
RESULT_DIR="results/alerts/ids2018-apt/custom-weights"
mkdir -p ${RESULT_DIR}
custom_weights_measure_ranking "${ALERT_FILE}" "zeek" "1" "${RESULT_DIR}"

RESULT_DIR="results/alerts/ids2018-apt/equal-weights"
mkdir -p ${RESULT_DIR}
equal_weights_measure_ranking "${ALERT_FILE}" "zeek" "1" "${RESULT_DIR}"

echo "----------------------------"
# CIC-IDS2017
echo "CIC-IDS2017"
ALERT_FILE="data/alerts/ids2017/eve.json"

RESULT_DIR="results/alerts/ids2017"
mkdir -p ${RESULT_DIR}
measure_performance "${ALERT_FILE}" "suricata" "100" "${RESULT_DIR}"

# Graph id = 2 is the one containing the multi-stage attack
RESULT_DIR="results/alerts/ids2017/custom-weights"
mkdir -p ${RESULT_DIR}
custom_weights_measure_ranking "${ALERT_FILE}" "suricata" "2" "${RESULT_DIR}"

RESULT_DIR="results/alerts/ids2017/equal-weights"
mkdir -p ${RESULT_DIR}
equal_weights_measure_ranking "${ALERT_FILE}" "suricata" "2" "${RESULT_DIR}"

echo "----------------------------"
# CIC-IDS2017-Perf
echo "CIC-IDS2017-Perf"
ALERT_FILE="data/alerts/ids2017-perf/eve.json"

RESULT_DIR="results/alerts/ids2017-perf"
mkdir -p ${RESULT_DIR}
measure_performance "${ALERT_FILE}" "suricata" "1000" "${RESULT_DIR}"

echo "----------------------------"
echo "Generating figures..."

export PIPENV_PIPFILE=code/evaluation/Pipfile
pipenv install

mkdir -p results/figures
pipenv run jupyter nbconvert --to notebook --inplace --execute code/evaluation/detection.ipynb
pipenv run jupyter nbconvert --to notebook --inplace --execute code/evaluation/performance.ipynb