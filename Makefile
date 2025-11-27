.PHONY: mockdata-clean mockdata-dist

clean-mockdata:
	rm -rf tests/act/mockdata/dist/*

clean-node-modules:
	find tests -name node_modules -type d -prune -exec rm -rf '{}' +

mockdata-dist: clean-mockdata
	./scripts/mockdata-dist.sh simple-frontend
	./scripts/mockdata-dist.sh simple-frontend-yarn
	./scripts/mockdata-dist.sh simple-frontend-pnpm
	./scripts/mockdata-dist.sh simple-backend
	@echo All done!


