# Real-Time Kill Chain State Machine (RT-KCSM)

Detect multi-stage attacks by correlating alerts from Intrusion Detection Systems (IDS) to generate scenario graphs.

Implementation of [our paper](https://doi.org/10.1109/CNS66487.2025.11194951):
```bibtex
@inproceedings{kistenmacher2025rtkcsm,
  title={Real-Time Detection of Multi-Stage Attacks Using Kill Chain State Machines},
  author={Kistenmacher, Liliana and Talpur, Anum and Fischer, Mathias},
  booktitle={2025 IEEE Conference on Communications and Network Security (CNS)},
  pages={1--9},
  year={2025},
  organization={IEEE},
  doi={10.1109/CNS66487.2025.11194951}
}
```

## Run using Docker

```bash
docker run --rm -v ./evaluation/data/alerts/:/data/ -p 8080:8080 ghcr.io/uhh-iss/rt-kcsm:latest --server :8080 --file /data/ids2018-apt/notice.json --reader zeek
```

CLI options:
```bash
$ rtkcsm -h
Usage: rtkcsm [--file FILE] [--listen LISTEN] [--server SERVER] [--import IMPORT] [--reader READER] [--transport TRANSPORT] [--export EXPORT] [--risk RISK] [--profile PROFILE] [--profile-graph-ranking-id PROFILE-GRAPH-RANKING-ID] [--stage-weight STAGE-WEIGHT] [--profile-log-resolution PROFILE-LOG-RESOLUTION]

Options:
  --file FILE            filepath of logs from suricata (eve.json) or zeek (JSON format)
  --listen LISTEN        TCP port to listen on for alerts
  --server SERVER        web interface port for visualization
  --import IMPORT        Import existing graphs
  --reader READER        format for reading from transport: 'zeek', 'suricata', 'ocsf', 'suricata-tenzir' [default: suricata]
  --transport TRANSPORT
                         'file', 'stdin', or 'tcp' for ingesting alerts [default: file]
  --export EXPORT        file name of exported graphs from RT-KCSM
  --risk RISK            set risk score (low=0.5,default=1.0,high=1.5) of an IP address for a host/asset: --risk 10.0.0.1=1.5
  --profile PROFILE      performance profile options: memory=/path/to/file, cpu=/path/to/file, alerts=/path/to/file, graphs=/path/to/file, graph-ranking=/path/to/file, progress=true
  --profile-graph-ranking-id PROFILE-GRAPH-RANKING-ID
                         graph id for profiling ranking
  --stage-weight STAGE-WEIGHT
                         set custom stage weights (incoming, same-zone, different-zone, outgoing): --stage-weight incoming=0.1
  --profile-log-resolution PROFILE-LOG-RESOLUTION
                         resolution of updating alert count [default: 1000]
  --help, -h             display this help and exit
```

## Run experiments

Prerequisites:
- `Golang 1.24`
- `Node.js 23.11.0`
- `npm 10.9.2`
- `Python 3.13`
- `Jupyter Notebook`
- `bash` or `zsh`

### 1. Install RT-KCSM from source with Golang

```bash
cd src/web/
npm install
tsc
rollup -c
cd ..
go install .
cd ..
```

Make sure you have the go binaries folder in your `$PATH` variable.

### 2. Run the evaluation setup

```bash
cd evaluation
./run-evaluation.sh
```

Wait until it has finished. Now run all steps of the Jupyter Notebook(s):
- `evaluation/performance.ipynb`
- `evaluation/detection.ipynb`

The result figures are located in `evaluation/figures/`