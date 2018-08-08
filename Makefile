azure-instance:
	./test/smoke.sh 1 1 2 sample

local-instance:
	./test/smoke.sh 0 0 2 sample

azure-instance-with-usr:
	./test/smoke.sh 1 1 2 usr

local-instance-with-usr:
	./test/smoke.sh 0 0 2 usr

user-function:
	./test/build.sh docker usr

.PHONY: azure-instance local-instance user-function azure-instance-with-usr local-instance-with-usr
