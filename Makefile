.PHONY: clean-mockdata clean-node-modules clean-act-tmp mockdata-dist

clean-node-modules:
	find tests -name node_modules -type d -prune -exec rm -rf '{}' +

clean-act-tmp:
	rm -rf /tmp/act-artifacts
	rm -rf /tmp/act-cache

clean-act-toolcache-volumes:
	docker volume ls -q | grep "^act-toolcache-" | xargs docker volume rm

clean: clean-node-modules clean-act-tmp clean-act-toolcache-volumes

clean-mockdata:
	rm -rf tests/act/mockdata/dist/*

mockdata-dist: clean-mockdata
	./scripts/mockdata-dist.sh simple-frontend
	./scripts/mockdata-dist.sh simple-frontend-yarn
	./scripts/mockdata-dist.sh simple-frontend-pnpm
	./scripts/mockdata-dist.sh simple-backend
	@echo All done!


