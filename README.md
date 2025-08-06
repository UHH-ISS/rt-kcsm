## Run experiments

Prerequisites:
- `Golang 1.24`
- `Node.js 23.11.0`
- `npm 10.9.2`
- `Python 3.13`
- `Jupyter Notebook`
- `bash` or `zsh`

### Install RT-KCSM from source with Golang

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

### Run the evaluation setup

```bash
cd evaluation
./run-evaluation.sh
```

Wait until it has finished. Now run all steps of the Jupyter Notebook(s):
- `evaluation/performance.ipynb`
- `evaluation/detection.ipynb`

The result figures are located in `evaluation/figures/`