.PHONY: clean-mockdata clean-node-modules clean-act-tmp mockdata-dist mockdata-dist-artifacts mockdata

clean-node-modules:
	find tests -name node_modules -type d -prune -exec rm -rf '{}' +

clean-dist:
	find tests ! -path '*/act/*' -name dist -type d -prune -exec rm -rf '{}' +

clean-act-tmp:
	rm -rf /tmp/act-artifacts
	rm -rf /tmp/act-cache
	rm -rf /tmp/act-gcs

clean-act-toolcache-volumes:
	docker volume ls -q | grep "^act-toolcache-" | xargs docker volume rm

clean-mockdata:
	rm -rf tests/act/mockdata/*

clean: clean-node-modules clean-dist clean-act-tmp clean-act-toolcache-volumes

mockdata-dist: clean-mockdata
	for tc in $$(./scripts/find-tests.sh); do \
		./scripts/mockdata-dist.sh $$tc; \
	done
	@echo All done!

mockdata-dist-artifacts: mockdata-dist
	for tc in $$(./scripts/find-tests.sh); do \
		./scripts/mockdata-dist-artifacts.sh $$tc; \
	done
	@echo All done!

mockdata: mockdata-dist-artifacts
