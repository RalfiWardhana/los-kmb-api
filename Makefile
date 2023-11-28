test:
	@go test -coverprofile="unit_test/coverage.out" "./domain/..."
	@go tool cover -func="unit_test/coverage.out"
coverage:
	@go tool cover -html="unit_test/coverage.out"
run:
	@go run app/main.go
mock:
	@mockery --all --dir=domain/kmb/interfaces --output domain/kmb/interfaces/mocks --case underscore
	@mockery --all --dir=domain/filtering_new/interfaces --output domain/filtering_new/interfaces/mocks --case underscore
	@mockery --all --dir=domain/elaborate_ltv/interfaces --output domain/elaborate_ltv/interfaces/mocks --case underscore
	@mockery --all --dir=domain/cms/interfaces --output domain/cms/interfaces/mocks --case underscore
	@mockery --name JSON --dir=shared/common --output shared/common/json/mocks --case underscore
	@mockery --name PlatformLogInterface --dir=shared/common/platformlog --output shared/common/platformlog/mocks --case underscore
	@mockery --name UtilsInterface --dir=shared/utils --output shared/utils/mocks --case underscore