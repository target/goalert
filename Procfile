build: while true; do make -qs bin/goalert || make bin/goalert || (echo '\033[0;31mBuild Failure'; sleep 3); sleep 0.1; done

@watch-file=./bin/goalert
goalert: ./bin/goalert -l=localhost:3030 --ui-dir=web/src/build --db-url=postgres://goalert@localhost --listen-sysapi=localhost:1234 --listen-prometheus=localhost:2112

smtp: go run github.com/mailhog/MailHog -ui-bind-addr=localhost:8025 -api-bind-addr=localhost:8025 -smtp-bind-addr=localhost:1025 | grep -v KEEPALIVE
prom: bin/tools/prometheus --log.level=warn --config.file=devtools/prometheus/prometheus.yml --storage.tsdb.path=bin/prom-data/ --web.listen-address=localhost:9090

@watch-file=./web/src/esbuild.config.js
ui: yarn run esbuild --watch

grpcui: go run github.com/fullstorydev/grpcui/cmd/grpcui -plaintext -open-browser=false -port 8234 localhost:1234
