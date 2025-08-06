./decompress.sh

function custom_weights_measure_ranking {
    rtkcsm --file $1 --reader $2 --profile graph-ranking=graph-ranking.csv --profile-graph-ranking-id $3 --export graphs.json
}

function equal_weights_measure_ranking {
    rtkcsm --file $1 --reader $2 --stage-weight incoming=0.25 same-zone=0.25 different-zone=0.25 outgoing=0.25 --profile graph-ranking=graph-ranking.csv --profile-graph-ranking-id $3 --export graphs.json
}

function measure_performance {
    rtkcsm --file $1 --reader $2 --profile-log-resolution $3 --profile memory="memory.csv"
    rtkcsm --file $1 --reader $2 --profile-log-resolution $3 --profile alerts="alerts.csv"
}

CWD=$(pwd)

# CSE-CIC-IDS2018
echo "CSE-CIC-IDS2018"

mkdir -p data/results/ids2018/ && cd data/results/ids2018/
measure_performance "../../alerts/ids2018/eve.json" "suricata" "1000"

mkdir -p custom-weights && cd custom-weights
custom_weights_measure_ranking "../../../alerts/ids2018/eve.json" "suricata" "5486"
cd -

mkdir -p equal-weights && cd equal-weights
equal_weights_measure_ranking "../../../alerts/ids2018/eve.json" "suricata" "5486"
cd -

cd "${CWD}"
echo "----------------------------"
# CIC-IDS2018-APT
echo "CSE-CIC-IDS2018-APT"

mkdir -p data/results/ids2018-apt/ && cd data/results/ids2018-apt/
measure_performance "../../alerts/ids2018-apt/notice.json" "zeek" "1000"

# Graph id = 1 is the one containing the APT attack
mkdir -p custom-weights && cd custom-weights
custom_weights_measure_ranking "../../../alerts/ids2018-apt/notice.json" "zeek" "1"
cd -

mkdir -p equal-weights && cd equal-weights
equal_weights_measure_ranking "../../../alerts/ids2018-apt/notice.json" "zeek" "1"
cd -

cd "${CWD}"

echo "----------------------------"
# CIC-IDS2017
echo "CIC-IDS2017"

mkdir -p data/results/ids2017/ && cd data/results/ids2017/
measure_performance "../../alerts/ids2017/eve.json" "suricata" "100"

# Graph id = 2 is the one containing the multi-stage attack
mkdir -p custom-weights && cd custom-weights
custom_weights_measure_ranking "../../../alerts/ids2017/eve.json" "suricata" "2"
cd -

mkdir -p equal-weights && cd equal-weights
equal_weights_measure_ranking "../../../alerts/ids2017/eve.json" "suricata" "2"
cd -

cd "${CWD}"
echo "----------------------------"
# CIC-IDS2017-Perf
echo "CIC-IDS2017-Perf"

mkdir -p data/results/ids2017-perf/ && cd data/results/ids2017-perf/
measure_performance "../../alerts/ids2017-perf/eve.json" "suricata" "1000"

cd "${CWD}"