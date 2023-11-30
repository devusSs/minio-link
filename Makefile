# DO NOT CHANGE.
build:
	@chmod +x scripts/build.sh
	@./scripts/build.sh

# DO NOT CHANGE.
check:
	@chmod +x scripts/check.sh
	@./scripts/check.sh

# DO NOT CHANGE.
clean: check
	@chmod +x scripts/clean.sh
	@./scripts/clean.sh
