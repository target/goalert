build: while true; do make -qs bin/goalert || make bin/goalert || (echo '\033[0;31mBuild Failure'; sleep 3); sleep 0.1; done

@watch-file=./bin/goalert
goalert: ./bin/goalert -l=localhost:3030 --ui-dir=web/src/build --db-url=postgres://goalert@localhost --listen-sysapi=localhost:1234 --listen-prometheus=localhost:2112 --smtp-listen=localhost:9025 --email-integration-domain=localhost --listen-pprof=localhost:6060 --pprof-mutex-profile-fraction=1 --pprof-block-profile-rate=1000

smtp: ./bin/tools/mailpit -s localhost:1025 -l localhost:8025
prom: bin/tools/prometheus --log.level=warn --config.file=devtools/prometheus/prometheus.yml --storage.tsdb.path=bin/prom-data/ --web.listen-address=localhost:9090

@watch-file=./web/src/esbuild.config.js
ui: ./bin/tools/bun run esbuild --watch

grpcui: go run ./devtools/waitfor tcp://localhost:1234 && go run github.com/fullstorydev/grpcui/cmd/grpcui -plaintext -open-browser=false -port 8234 localhost:1234

oidc: ./bin/mockoidc
