PROJECT = github.com/chromium/hstspreload.appspot.com/...

.PHONY: build
build:
	go build ${PROJECT}

.PHONY: format
format:
	go fmt ${PROJECT}
	# Need to specify non-default clang-format: https://crbug.com/558447
	/usr/local/bin/clang-format -i -style=Google files/static/js/*.js

.PHONY: lint
lint:
	go vet ${PROJECT}
	golint -set_exit_status ${PROJECT}

.PHONY: pre-commit
pre-commit: lint build test

.PHONY: travis
travis: pre-commit

.PHONY: deploy
deploy: check version
	date
	time aedeploy gcloud preview app deploy app.yaml --promote
	date

CURRENT_DIR = "$(shell pwd)"
EXPECTED_DIR = "${GOPATH}/src/github.com/chromium/hstspreload.appspot.com"

.PHONY: check
check:
ifeq (${CURRENT_DIR}, ${EXPECTED_DIR})
	@echo "PASS: Current directory is in \$$GOPATH."
else
	@echo "FAIL"
	@echo "Expected: ${EXPECTED_DIR}"
	@echo "Actual:   ${CURRENT_DIR}"
endif

# Version file.

.PHONY: version
version:
	git rev-parse HEAD > ./files/version

# Google Cloud Datastore Emulator

GCD_NAME = gcd-grpc-1.0.0
DATABASE_TESTING_FOLDER = ${HOME}/.datastore-emulator

.PHONY: get-datastore-emulator
get-datastore-emulator: ${DATABASE_TESTING_FOLDER}/gcd/gcd.sh
${DATABASE_TESTING_FOLDER}/gcd/gcd.sh:
	mkdir -p "${DATABASE_TESTING_FOLDER}"
	curl "https://storage.googleapis.com/gcd/tools/${GCD_NAME}.zip" -o "${DATABASE_TESTING_FOLDER}/${GCD_NAME}.zip"
	unzip "${DATABASE_TESTING_FOLDER}/${GCD_NAME}.zip" -d "${DATABASE_TESTING_FOLDER}"

# Testing

.PHONY: serve
serve: check get-datastore-emulator version
	go run *.go -local

.PHONY: test
test: get-datastore-emulator
	go test -v -cover "${PROJECT}"
