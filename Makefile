swagger:
	@swag init -g app/main.go
test:
	@go test -coverprofile="unit_test/coverage.out" "./domain/..."
	@go tool cover -func="unit_test/coverage.out"
coverage:
	@go tool cover -html="unit_test/coverage.out"
run:
	@go run app/main.go
mock_filtering:
	@mockery --all --dir=domain/filtering/interfaces --output domain/filtering/interfaces/mocks --case underscore
mock_elaborate:
	@mockery --all --dir=domain/elaborate/interfaces --output domain/elaborate/interfaces/mocks --case underscore
mock_kmb:
	@mockery --all --dir=domain/kmb/interfaces --output domain/kmb/interfaces/mocks --case underscore
mock_filtering_new:
	@mockery --all --dir=domain/filtering_new/interfaces --output domain/filtering_new/interfaces/mocks --case underscore
mock_elaborate_ltv:
	@mockery --all --dir=domain/elaborate_ltv/interfaces --output domain/elaborate_ltv/interfaces/mocks --case underscore
mock_cms:
	@mockery --all --dir=domain/cms/interfaces --output domain/cms/mocks --case underscore
mock_json:
	@mockery --name JSON --dir=shared/common --output shared/common/json/mocks --case underscore
mock_platformlog:
	@mockery --name PlatformLogInterface --dir=shared/common/platformlog --output shared/common/platformlog/mocks --case underscore
mock_platformcache:
	@mockery --name PlatformCacheInterface --dir=shared/common/platformcache --output shared/common/platformcache/mocks --case underscore
