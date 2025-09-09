# Changelog

## [2.0.0](https://github.com/grafana/plugin-ci-workflows/compare/ci-cd-workflows/v1.2.0...ci-cd-workflows/v2.0.0) (2025-09-09)


### ⚠ BREAKING CHANGES

* Add support for tagged releases ([#282](https://github.com/grafana/plugin-ci-workflows/issues/282))
* **deps:** update actions/setup-go action to v6 ([#274](https://github.com/grafana/plugin-ci-workflows/issues/274))
* Add support for pinning plugin-ci-workflows to semver tags ([#216](https://github.com/grafana/plugin-ci-workflows/issues/216))

### 🎉 Features

* add step to cache playwright ([#253](https://github.com/grafana/plugin-ci-workflows/issues/253)) ([061c113](https://github.com/grafana/plugin-ci-workflows/commit/061c113e379efa0e0196eb9cad3e727ed0f2986d))
* add support for backend secrets ([#255](https://github.com/grafana/plugin-ci-workflows/issues/255)) ([1cb7ddc](https://github.com/grafana/plugin-ci-workflows/commit/1cb7ddcba0978c03806d5ae0a3c4eda7b1006f21))
* Add support for pinning plugin-ci-workflows to semver tags ([#216](https://github.com/grafana/plugin-ci-workflows/issues/216)) ([8458e50](https://github.com/grafana/plugin-ci-workflows/commit/8458e5069ef67cdd50c2afa94ddac1b99545c5d0))


### 📝 Documentation

* add example workflow for change scope plugins action ([#169](https://github.com/grafana/plugin-ci-workflows/issues/169)) ([eba74c1](https://github.com/grafana/plugin-ci-workflows/commit/eba74c1ae47e818595ed5fefbc5550ce758241d7))


### 🤖 Continuous Integrations

* Add support for tagged releases ([#282](https://github.com/grafana/plugin-ci-workflows/issues/282)) ([ccd1f75](https://github.com/grafana/plugin-ci-workflows/commit/ccd1f75486be7f7290a6467d1c2c4f0fc343f898))


### 🔧 Chores

* **deps:** update actions/attest-build-provenance action to v3 ([#268](https://github.com/grafana/plugin-ci-workflows/issues/268)) ([1a170e5](https://github.com/grafana/plugin-ci-workflows/commit/1a170e5d94d91adaf3a0c2f79267df02c4657851))
* **deps:** update actions/checkout action to v5 ([#230](https://github.com/grafana/plugin-ci-workflows/issues/230)) ([d4ad142](https://github.com/grafana/plugin-ci-workflows/commit/d4ad142587ec6383f1d03f6998b675d4f713e3d4))
* **deps:** update actions/github-script action to v8 ([#277](https://github.com/grafana/plugin-ci-workflows/issues/277)) ([9a182f0](https://github.com/grafana/plugin-ci-workflows/commit/9a182f0875c7245f28b68477be1b6ffa29c654f0))
* **deps:** update actions/setup-go action to v6 ([#274](https://github.com/grafana/plugin-ci-workflows/issues/274)) ([66cf8ba](https://github.com/grafana/plugin-ci-workflows/commit/66cf8bad1ea05b14659adc0d592e4732a16470d9))
* **deps:** update amannn/action-semantic-pull-request action to v6 ([#234](https://github.com/grafana/plugin-ci-workflows/issues/234)) ([4179d5a](https://github.com/grafana/plugin-ci-workflows/commit/4179d5a711f2a43297a318b0c64d68e2b88aa009))
* **deps:** update google-github-actions/auth action to v2.1.13 ([#260](https://github.com/grafana/plugin-ci-workflows/issues/260)) ([7d4bfda](https://github.com/grafana/plugin-ci-workflows/commit/7d4bfdafad4b0900c1b91cb4628b6ef8618fb30f))
* **deps:** update google-github-actions/auth action to v3 ([#269](https://github.com/grafana/plugin-ci-workflows/issues/269)) ([4415c03](https://github.com/grafana/plugin-ci-workflows/commit/4415c03ca3eec53db5b1e5470ddea31ba1e02b2e))
* **deps:** update google-github-actions/setup-gcloud action to v2.2.1 ([#250](https://github.com/grafana/plugin-ci-workflows/issues/250)) ([ebe7e7b](https://github.com/grafana/plugin-ci-workflows/commit/ebe7e7beb9c816edea862e524a182310d114c721))
* **deps:** update google-github-actions/setup-gcloud action to v3 ([#251](https://github.com/grafana/plugin-ci-workflows/issues/251)) ([429f118](https://github.com/grafana/plugin-ci-workflows/commit/429f118cccf4b92c9038c41c3d9420a492edc74b))
* **deps:** update google-github-actions/upload-cloud-storage action to v2.2.4 ([#271](https://github.com/grafana/plugin-ci-workflows/issues/271)) ([2103538](https://github.com/grafana/plugin-ci-workflows/commit/21035385f0b98e644cc5f3262c5206ccfbdedad3))
* **deps:** update google-github-actions/upload-cloud-storage action to v3 ([#272](https://github.com/grafana/plugin-ci-workflows/issues/272)) ([7634080](https://github.com/grafana/plugin-ci-workflows/commit/76340806b625213d3219c0ad4385ffb74c16e3b3))
* **deps:** update grafana/shared-workflows/get-vault-secrets action to v1.3.0 ([#243](https://github.com/grafana/plugin-ci-workflows/issues/243)) ([929a251](https://github.com/grafana/plugin-ci-workflows/commit/929a2518a32ea0d20c6c1e07fd65fa540c2e7882))
* **deps:** update grafana/shared-workflows/trigger-argo-workflow action to v1.2.0 ([#247](https://github.com/grafana/plugin-ci-workflows/issues/247)) ([d613f64](https://github.com/grafana/plugin-ci-workflows/commit/d613f64e77f3b7da62a0a16156235ff2b373cebd))
* **deps:** update softprops/action-gh-release action to v2.3.3 ([#281](https://github.com/grafana/plugin-ci-workflows/issues/281)) ([e010fd3](https://github.com/grafana/plugin-ci-workflows/commit/e010fd3d9696478e61601f7d744b21b4ede91f1a))

## [1.2.0](https://github.com/grafana/plugin-ci-workflows/compare/ci-cd-workflows/v1.1.2...ci-cd-workflows/v1.2.0) (2025-08-21)


### 🎉 Features

* add option to disable docs publishing alltogether ([#237](https://github.com/grafana/plugin-ci-workflows/issues/237)) ([817b34f](https://github.com/grafana/plugin-ci-workflows/commit/817b34fe61d0f1be7bf6e76804d6e9b98108b0e3))
* add publish to catalog as pending option in ci and publish action ([#240](https://github.com/grafana/plugin-ci-workflows/issues/240)) ([0086720](https://github.com/grafana/plugin-ci-workflows/commit/00867201b23dd72abcc90b4746af139c90e0cebd))
* allow all workflows to set both node-version-file and node-version ([#208](https://github.com/grafana/plugin-ci-workflows/issues/208)) ([29a45db](https://github.com/grafana/plugin-ci-workflows/commit/29a45dbcdf7fded1f92f5c07495684595db2ac4d))
* allow passing of secrets to frontend build steps ([#222](https://github.com/grafana/plugin-ci-workflows/issues/222)) ([aebe053](https://github.com/grafana/plugin-ci-workflows/commit/aebe053353b0e1382be7538ab6e02236861358ac))
* **ci:** Add support for authenticating to NPM Google Artifact Registry ([#224](https://github.com/grafana/plugin-ci-workflows/issues/224)) ([87549ef](https://github.com/grafana/plugin-ci-workflows/commit/87549ef5afcb00132116658770a47b9229acdba1))
* **playwright:** env for GRAFANA_VERSION to allow the playwright config access ([#206](https://github.com/grafana/plugin-ci-workflows/issues/206)) ([ab20c6a](https://github.com/grafana/plugin-ci-workflows/commit/ab20c6a21e6da6722952759a57d1aa4057a2e118))
* **publish:** Sanity check ZIP files before publishing to catalog ([#199](https://github.com/grafana/plugin-ci-workflows/issues/199)) ([f799b45](https://github.com/grafana/plugin-ci-workflows/commit/f799b45c37748ac4c1426a79256d0c40c9c07648))
* Switch runners for exported workflows to self-hosted ones ([#217](https://github.com/grafana/plugin-ci-workflows/issues/217)) ([2851e10](https://github.com/grafana/plugin-ci-workflows/commit/2851e1098cbdac81eeaf8664d774be08bb84459a))


### 🐛 Bug Fixes

* **cd:** github draft optional, omit commits ([#202](https://github.com/grafana/plugin-ci-workflows/issues/202)) ([200078a](https://github.com/grafana/plugin-ci-workflows/commit/200078afee7387a1bb615a0961fb530297ac6451))
* **ci:** Fix NPM auth when running Playwright E2E tests ([#231](https://github.com/grafana/plugin-ci-workflows/issues/231)) ([8da89c2](https://github.com/grafana/plugin-ci-workflows/commit/8da89c22b1e96fe930b9b23be909d4e75bc341a9))
* Fix test-docs and publish-docs steps failing ([#225](https://github.com/grafana/plugin-ci-workflows/issues/225)) ([a17fdd6](https://github.com/grafana/plugin-ci-workflows/commit/a17fdd6e0e71870b665f94ca4132e06a64417aff))
* pass node-version down from ci to playwright ([#219](https://github.com/grafana/plugin-ci-workflows/issues/219)) ([840c01b](https://github.com/grafana/plugin-ci-workflows/commit/840c01bf99d7d9135c327e528abe1673072d90f5))


### 🤖 Continuous Integrations

* Switch runners for internal workflows to self-hosted ones ([#218](https://github.com/grafana/plugin-ci-workflows/issues/218)) ([d049b9d](https://github.com/grafana/plugin-ci-workflows/commit/d049b9d62f92cea820ff6b5a9c9a7b8db0bc908e))


### 🔧 Chores

* **deps:** update actions/checkout action to v4.3.0 ([#229](https://github.com/grafana/plugin-ci-workflows/issues/229)) ([7e02ef2](https://github.com/grafana/plugin-ci-workflows/commit/7e02ef237cad8a878b019502c854c0087f917c71))
* **deps:** update actions/create-github-app-token action to v2.1.0 ([#226](https://github.com/grafana/plugin-ci-workflows/issues/226)) ([486d5d6](https://github.com/grafana/plugin-ci-workflows/commit/486d5d665b8ca641b397fa2db72d4069f53d7b74))
* **deps:** update actions/create-github-app-token action to v2.1.1 ([#228](https://github.com/grafana/plugin-ci-workflows/issues/228)) ([d74d260](https://github.com/grafana/plugin-ci-workflows/commit/d74d260b1cac3551354981a6991e402fce04717f))
* **deps:** update actions/download-artifact action to v5 ([#210](https://github.com/grafana/plugin-ci-workflows/issues/210)) ([3581472](https://github.com/grafana/plugin-ci-workflows/commit/358147235f8c7770ce7ad64fdf878f0b12a8a181))
* **deps:** update google-github-actions/setup-gcloud action to v2.2.0 ([#227](https://github.com/grafana/plugin-ci-workflows/issues/227)) ([6fe0d0b](https://github.com/grafana/plugin-ci-workflows/commit/6fe0d0b17e6ec4cf9ea7d64b95f2e44fd057d561))
* **deps:** update googleapis/release-please-action action to v4.3.0 ([#244](https://github.com/grafana/plugin-ci-workflows/issues/244)) ([b148a5a](https://github.com/grafana/plugin-ci-workflows/commit/b148a5a472e0d821598970265c8f19dbc9706939))
* Fix some Zizmor issues ([#211](https://github.com/grafana/plugin-ci-workflows/issues/211)) ([5007dce](https://github.com/grafana/plugin-ci-workflows/commit/5007dce7256b91109fad412e29cd9bdca1078e9b))

## [1.1.2](https://github.com/grafana/plugin-ci-workflows/compare/ci-cd-workflows/v1.1.1...ci-cd-workflows/v1.1.2) (2025-08-05)


### 🐛 Bug Fixes

* moving secret fetching up ([#163](https://github.com/grafana/plugin-ci-workflows/issues/163)) ([e478422](https://github.com/grafana/plugin-ci-workflows/commit/e478422847f9e1c0df5535b79f5182c58e68efca))
* **publish:** Fix cd-style push events not being considered trusted ([#198](https://github.com/grafana/plugin-ci-workflows/issues/198)) ([34e0b28](https://github.com/grafana/plugin-ci-workflows/commit/34e0b281238a74d4dd7eb8886ea1c331db616ee4))
* **publish:** Fix secrets not being passed to Playwright when publishing ([#196](https://github.com/grafana/plugin-ci-workflows/issues/196)) ([71132a3](https://github.com/grafana/plugin-ci-workflows/commit/71132a38602f14d6f3cf50c26bc34071679bdf3b))


### 📝 Documentation

* add more about scopes input ([#171](https://github.com/grafana/plugin-ci-workflows/issues/171)) ([0b5c370](https://github.com/grafana/plugin-ci-workflows/commit/0b5c37064531d725ec257153a27de03107352201))


### 🤖 Continuous Integrations

* Add workflow for ensuring examples/base/README.md is up-to-date ([#191](https://github.com/grafana/plugin-ci-workflows/issues/191)) ([1ce6526](https://github.com/grafana/plugin-ci-workflows/commit/1ce652685b2b04de2b516f4e438d6fcaaf8b1bc8))


### 🔧 Chores

* **deps:** Bump google-github-actions/upload-cloud-storage ([#189](https://github.com/grafana/plugin-ci-workflows/issues/189)) ([4a55c8d](https://github.com/grafana/plugin-ci-workflows/commit/4a55c8d55981e2ab965acedfabd2df260c74dea9))
* **deps:** update google-github-actions/auth action to v2.1.11 ([#176](https://github.com/grafana/plugin-ci-workflows/issues/176)) ([f0c05fe](https://github.com/grafana/plugin-ci-workflows/commit/f0c05fe0b70002c6831a98f1f1ac9770952d74aa))
* **deps:** update google-github-actions/auth action to v2.1.12 ([#201](https://github.com/grafana/plugin-ci-workflows/issues/201)) ([de83d72](https://github.com/grafana/plugin-ci-workflows/commit/de83d728fef7e0b569b36f179d54147f2b29c30c))
* **deps:** update google-github-actions/setup-gcloud action to v2.1.5 ([#177](https://github.com/grafana/plugin-ci-workflows/issues/177)) ([3aa265b](https://github.com/grafana/plugin-ci-workflows/commit/3aa265b3fb17ec192b912b8cd4f50f2510d6ece2))
* **deps:** update step-security/harden-runner action to v2.13.0 ([#179](https://github.com/grafana/plugin-ci-workflows/issues/179)) ([bb88fd6](https://github.com/grafana/plugin-ci-workflows/commit/bb88fd6043e0170b79f1cbb629d9cf48442384a6))

## [1.1.1](https://github.com/grafana/plugin-ci-workflows/compare/ci-cd-workflows/v1.1.0...ci-cd-workflows/v1.1.1) (2025-07-11)


### 📝 Documentation

* Improvements to documentation and examples ([#159](https://github.com/grafana/plugin-ci-workflows/issues/159)) ([da3d700](https://github.com/grafana/plugin-ci-workflows/commit/da3d700ad152a55b04526cfefed731e858fcaa56))

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


### 💄 Styles

* add empty lines for tesing releases ([#157](https://github.com/grafana/plugin-ci-workflows/issues/157)) ([faed828](https://github.com/grafana/plugin-ci-workflows/commit/faed828ef944db9b8280f0b04edf5ec14b6c87a8))


### 🤖 Continuous Integrations

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


### 🔧 Chores

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
* Setup release please for actions versioning ([#151](https://github.com/grafana/plugin-ci-workflows/issues/151)) ([5962b96](https://github.com/grafana/plugin-ci-workflows/commit/5962b96ab9016ee5e893c0ef9ef51977388c04ee))
* some nitpicks ([#8](https://github.com/grafana/plugin-ci-workflows/issues/8)) ([4cf907a](https://github.com/grafana/plugin-ci-workflows/commit/4cf907a5633af8a47eb4e549135b18b1604a001e))
* Use get-vault-secrets without exporting env variables ([#130](https://github.com/grafana/plugin-ci-workflows/issues/130)) ([0ff10ef](https://github.com/grafana/plugin-ci-workflows/commit/0ff10ef11ee73912d45684a3820acdce88dd20ee))
