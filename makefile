build_dir:
	mkdir -p builds

build_lambda: build_dir
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o builds/sts-lambda lambda/*.go

build_local_bin:
	CGO_ENABLED=9 go build -o ./sts cmd/*.go	

build: build_lambda

deploy_lambda: build_lambda
	cd terraform/terragrunt/lambda && \
	terragrunt apply
	