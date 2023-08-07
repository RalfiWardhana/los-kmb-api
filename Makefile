unit-test:
	@go test -coverprofile="unit_test/coverage.out" "./domain/.../usecase"
	@go tool cover -func="unit_test/coverage.out"
html-coverage:
	@go tool cover -html="unit_test/coverage.out"
run-app:
	@go run app/main.go