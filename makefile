build_dir:
	mkdir -p builds

build_lambda: build_dir
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o builds/sts-lambda lambda/*.go

build_local_bin: build_dir
	CGO_ENABLED=0 go build -o ./builds/sts_local cmd/*.go
	sudo mv ./builds/sts_local /usr/local/bin/sts	

build: build_lambda

deploy_lambda: build_lambda
	cd terraform/terragrunt/lambda && \
	terragrunt apply

deploy: build_lambda
	cd terraform/terragrunt && \
	terragrunt run-all apply

clean_builds: 
	rm -rf builds

clean_terraform:
	cd terraform/terragrunt && \
	terragrunt run-all destroy

clean: clean_terraform clean_builds