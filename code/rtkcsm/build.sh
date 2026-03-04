(
    cd code/rtkcsm/src/web/
    npm install
    npx tsc
    npx rollup -c
)
(cd code/rtkcsm/src && go install .)