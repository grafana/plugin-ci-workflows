# Changelog

## [1.1.0](https://github.com/grafana/plugin-ci-workflows/compare/ci-cd-workflows/v1.0.0...ci-cd-workflows/v1.1.0) (2025-07-08)


### 🎉 Features

* allow custom working directory for plugins ([#110](https://github.com/grafana/plugin-ci-workflows/issues/110)) ([0e6a972](https://github.com/grafana/plugin-ci-workflows/commit/0e6a972a3dde0c1a7a50deabaa3e4fa29f353aa1))
* **cd:** Pass Playwright inputs to the ci workflow ([#70](https://github.com/grafana/plugin-ci-workflows/issues/70)) ([9afb3f7](https://github.com/grafana/plugin-ci-workflows/commit/9afb3f7acb0b343a6012d25465628b116395afaf))
* dockerized playwright tests ([#67](https://github.com/grafana/plugin-ci-workflows/issues/67)) ([66de69f](https://github.com/grafana/plugin-ci-workflows/commit/66de69feb58f1396501be4a1459f5a93de35174b))
* **playwright:** allow custom plugin directories ([#119](https://github.com/grafana/plugin-ci-workflows/issues/119)) ([8bfca2e](https://github.com/grafana/plugin-ci-workflows/commit/8bfca2e92d8969217b8ddd68d99f5ab728496843))
* simplify playwright docker e2e with profiles ([#103](https://github.com/grafana/plugin-ci-workflows/issues/103)) ([fb156a0](https://github.com/grafana/plugin-ci-workflows/commit/fb156a03982a2a5c31ac7c74e873a25675ce6e34))


### 🐛 Bug Fixes

* CD action to avoid referring to non-existent input ([#106](https://github.com/grafana/plugin-ci-workflows/issues/106)) ([1114ffd](https://github.com/grafana/plugin-ci-workflows/commit/1114ffdf4bac5c3cf79b0db015852fd0842fccd9))
* remove accidental reference to nonexistent input in playwright step ([#111](https://github.com/grafana/plugin-ci-workflows/issues/111)) ([34e7433](https://github.com/grafana/plugin-ci-workflows/commit/34e743382725933117deb4f2e9219cb848220358))
* remove vaultt instance param ([265f2aa](https://github.com/grafana/plugin-ci-workflows/commit/265f2aa570c6959bc0ef85f0b92adc6c5d4432fc))


### 🤖 Continuous Integration

* Add final job to check for E2E tests matrix status ([#94](https://github.com/grafana/plugin-ci-workflows/issues/94)) ([d6962c2](https://github.com/grafana/plugin-ci-workflows/commit/d6962c2352b969a252355978092dbcfbfd90a643))
* Add Trufflehog secrets scanning for packaged plugin ZIP files ([#12](https://github.com/grafana/plugin-ci-workflows/issues/12)) ([97c03bc](https://github.com/grafana/plugin-ci-workflows/commit/97c03bcccd8dc75490b418c2a43ca4284dcf4a1e))
* Allow passing secrets to Playwright ([#141](https://github.com/grafana/plugin-ci-workflows/issues/141)) ([dfa6b3e](https://github.com/grafana/plugin-ci-workflows/commit/dfa6b3e6f5357a9c73a0f91bf8b70aeb5a6ebc53))
* Fix checkout for forks ([#96](https://github.com/grafana/plugin-ci-workflows/issues/96)) ([afa8eb2](https://github.com/grafana/plugin-ci-workflows/commit/afa8eb23b0520f0f5d08b78ef2104bbee535ff93))
* Fix docs exist check ([#29](https://github.com/grafana/plugin-ci-workflows/issues/29)) ([d3f2a6c](https://github.com/grafana/plugin-ci-workflows/commit/d3f2a6c8101f5d0d53ff242bd01217f64d309855))
* Fix error when packaging plugin for fork PRs ([#143](https://github.com/grafana/plugin-ci-workflows/issues/143)) ([63de1b3](https://github.com/grafana/plugin-ci-workflows/commit/63de1b32f52af50ed2a4921a29f72e1a769389fa))
* Fix GCS upload path when targeting non-main branches ([#25](https://github.com/grafana/plugin-ci-workflows/issues/25)) ([151411c](https://github.com/grafana/plugin-ci-workflows/commit/151411cb013932c696297fe206963e8575145e9d))
* Fix GCS upload skipped for push events ([#43](https://github.com/grafana/plugin-ci-workflows/issues/43)) ([fa5e7de](https://github.com/grafana/plugin-ci-workflows/commit/fa5e7ded40b90ff05dde3ea6410c9dbe1a1da33d))
* Fix IS_FORK for push events ([#44](https://github.com/grafana/plugin-ci-workflows/issues/44)) ([99e00b7](https://github.com/grafana/plugin-ci-workflows/commit/99e00b79cbedb46d4034119d1b7d91de38cb0fe1))
* Trufflehog: Only report verified and unknown secrets ([#22](https://github.com/grafana/plugin-ci-workflows/issues/22)) ([aa4c703](https://github.com/grafana/plugin-ci-workflows/commit/aa4c703a6a7d3eec99d36a5e77e2d586435d6ff6))


### 🔧 Miscellaneous Chores

* Bump default Go version to 1.23 ([#27](https://github.com/grafana/plugin-ci-workflows/issues/27)) ([24c53b5](https://github.com/grafana/plugin-ci-workflows/commit/24c53b5bf16237ef0b863a6a7f18c46374728d7d))
* Bump softprops/action-gh-release to v2.2.1 ([#23](https://github.com/grafana/plugin-ci-workflows/issues/23)) ([9ca12b0](https://github.com/grafana/plugin-ci-workflows/commit/9ca12b0e1badfbe0c4ee8e4af6bcae6af5cdb552))
* **deps:** Bump actions/attest-build-provenance from 2.2.3 to 2.3.0 ([#79](https://github.com/grafana/plugin-ci-workflows/issues/79)) ([c1fed14](https://github.com/grafana/plugin-ci-workflows/commit/c1fed14d01f040f9ad828bc32366c1f5e99399f9))
* **deps:** Bump actions/attest-build-provenance from 2.3.0 to 2.4.0 ([#137](https://github.com/grafana/plugin-ci-workflows/issues/137)) ([acb4a6b](https://github.com/grafana/plugin-ci-workflows/commit/acb4a6b6d26e0e978386752a31a6340bfc01b445))
* **deps:** Bump actions/create-github-app-token from 1.12.0 to 2.0.2 ([#77](https://github.com/grafana/plugin-ci-workflows/issues/77)) ([9132940](https://github.com/grafana/plugin-ci-workflows/commit/91329403f85b22e31ac6cbe83910352b593fd090))
* **deps:** Bump actions/create-github-app-token from 2.0.2 to 2.0.6 ([#108](https://github.com/grafana/plugin-ci-workflows/issues/108)) ([81fe284](https://github.com/grafana/plugin-ci-workflows/commit/81fe284f10a1b56fba6e44027ca531e17b66ea71))
* **deps:** Bump actions/download-artifact from 4.2.1 to 4.3.0 ([#76](https://github.com/grafana/plugin-ci-workflows/issues/76)) ([1d59387](https://github.com/grafana/plugin-ci-workflows/commit/1d59387e7c96310085602f880d99f37c6d4f7649))
* **deps:** Bump actions/setup-node from 4.3.0 to 4.4.0 ([#81](https://github.com/grafana/plugin-ci-workflows/issues/81)) ([e94caab](https://github.com/grafana/plugin-ci-workflows/commit/e94caab2af542079166419080a112aebd40e25e5))
* **deps:** Bump google-github-actions/auth from 2.1.8 to 2.1.10 ([#78](https://github.com/grafana/plugin-ci-workflows/issues/78)) ([78b05bb](https://github.com/grafana/plugin-ci-workflows/commit/78b05bb72822848b39f72f91f47f12d9057666f2))
* **deps:** Bump softprops/action-gh-release from 2.2.1 to 2.2.2 ([#80](https://github.com/grafana/plugin-ci-workflows/issues/80)) ([6156f9c](https://github.com/grafana/plugin-ci-workflows/commit/6156f9c7efac2d99f180260dcbff8c6221aaa7e2))
* **deps:** Bump softprops/action-gh-release from 2.2.2 to 2.3.2 ([#136](https://github.com/grafana/plugin-ci-workflows/issues/136)) ([669131a](https://github.com/grafana/plugin-ci-workflows/commit/669131a0b4fb6c35fe8e20f16d149c375dbde80a))
* Fix actionlint warnings ([#83](https://github.com/grafana/plugin-ci-workflows/issues/83)) ([9bacf72](https://github.com/grafana/plugin-ci-workflows/commit/9bacf72016cc3cd1b0fdd0eacf10baccdaf4f87c))
* setup release please ([ad8a863](https://github.com/grafana/plugin-ci-workflows/commit/ad8a863cf57a9c50faa1f9252cee2ea810d957f1))
* some nitpicks ([#8](https://github.com/grafana/plugin-ci-workflows/issues/8)) ([4cf907a](https://github.com/grafana/plugin-ci-workflows/commit/4cf907a5633af8a47eb4e549135b18b1604a001e))
* Use get-vault-secrets without exporting env variables ([#130](https://github.com/grafana/plugin-ci-workflows/issues/130)) ([0ff10ef](https://github.com/grafana/plugin-ci-workflows/commit/0ff10ef11ee73912d45684a3820acdce88dd20ee))
