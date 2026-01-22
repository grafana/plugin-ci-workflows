# Changelog

## [6.0.0](https://github.com/grafana/plugin-ci-workflows/compare/ci-cd-workflows/v5.1.0...ci-cd-workflows/v6.0.0) (2026-01-19)


### ⚠ BREAKING CHANGES

* **playwright:** remove version and image arguments ([#507](https://github.com/grafana/plugin-ci-workflows/issues/507))
* **cd:** allow releasing to prod from non-main branches ([#378](https://github.com/grafana/plugin-ci-workflows/issues/378))

### 🎉 Features

* **cd:** allow releasing to prod from non-main branches ([#378](https://github.com/grafana/plugin-ci-workflows/issues/378)) ([e1cf5f6](https://github.com/grafana/plugin-ci-workflows/commit/e1cf5f6083704fd837f4ec11761dd397b74f0929))
* **ci:** add warning when ci and cd workflows are not on the same reference ([#492](https://github.com/grafana/plugin-ci-workflows/issues/492)) ([4e9ca04](https://github.com/grafana/plugin-ci-workflows/commit/4e9ca04ed0b13a33baf41d6cd53ce6fedb8cf683))
* **ci:** support react image in e2e matrix ([#496](https://github.com/grafana/plugin-ci-workflows/issues/496)) ([6d8cb8c](https://github.com/grafana/plugin-ci-workflows/commit/6d8cb8c5240cc11619c221beb201ee07e5ee855c))
* **playwright:** remove version and image arguments ([#507](https://github.com/grafana/plugin-ci-workflows/issues/507)) ([f70e822](https://github.com/grafana/plugin-ci-workflows/commit/f70e82266068fbe7cdfa97a5120d4e616ab1275f))


### ✅ Tests

* **act:** add gcom mock, cd.yml changes for testing ([#501](https://github.com/grafana/plugin-ci-workflows/issues/501)) ([e3f1779](https://github.com/grafana/plugin-ci-workflows/commit/e3f1779be66d8a2c7051adb0a42beb73184c5bea))
* **act:** fix validator tests ([#506](https://github.com/grafana/plugin-ci-workflows/issues/506)) ([3c93dda](https://github.com/grafana/plugin-ci-workflows/commit/3c93dda78df9f6ac587e2c7e116ac7b1e04627a3))


### 🔧 Chores

* **ci:** run golangci-lint on internal go packages ([#485](https://github.com/grafana/plugin-ci-workflows/issues/485)) ([303148f](https://github.com/grafana/plugin-ci-workflows/commit/303148fcc64f3ecbca42d2eda4142dceafa2296b))
* **deps:** update actions/attest-build-provenance action to v3.1.0 ([#484](https://github.com/grafana/plugin-ci-workflows/issues/484)) ([ba9c56e](https://github.com/grafana/plugin-ci-workflows/commit/ba9c56e2992bc28da40bce896349d7122c03e3f1))
* **deps:** update actions/cache action to v5 ([#481](https://github.com/grafana/plugin-ci-workflows/issues/481)) ([3eaee5a](https://github.com/grafana/plugin-ci-workflows/commit/3eaee5a68d81c66037d8ba723d4070cb0e1be850))
* **deps:** update stefanzweifel/git-auto-commit-action action to v7.1.0 ([#483](https://github.com/grafana/plugin-ci-workflows/issues/483)) ([f9a22aa](https://github.com/grafana/plugin-ci-workflows/commit/f9a22aa89487aa54a299a0c3b1e59e0832b7a0b9))
* **deps:** update step-security/harden-runner action to v2.14.0 ([#477](https://github.com/grafana/plugin-ci-workflows/issues/477)) ([cdab179](https://github.com/grafana/plugin-ci-workflows/commit/cdab179fe05abb011a6b236c45426d4c6004f8af))

## [5.1.0](https://github.com/grafana/plugin-ci-workflows/compare/ci-cd-workflows/v5.0.0...ci-cd-workflows/v5.1.0) (2026-01-07)


### 🎉 Features

* **ci:** setup Go version from go.mod if possible ([#465](https://github.com/grafana/plugin-ci-workflows/issues/465)) ([5d5ab0a](https://github.com/grafana/plugin-ci-workflows/commit/5d5ab0a5fddf3b9ab538e5207a996d1fee8f908a))


### 🐛 Bug Fixes

* **tests:** fix act checkout in pull_request events ([#463](https://github.com/grafana/plugin-ci-workflows/issues/463)) ([56388a7](https://github.com/grafana/plugin-ci-workflows/commit/56388a7772a3a4f35fe67f54b77598a38ad30d9c))


### ⚡ Performance Improvements

* **ci:** cache mage, Trufflehog and golangci-lint binaries ([#467](https://github.com/grafana/plugin-ci-workflows/issues/467)) ([b736834](https://github.com/grafana/plugin-ci-workflows/commit/b73683467666c8a40cf2f475bec1ccfa77142d3a))


### ✅ Tests

* **act:** add plugin-validator tests ([#457](https://github.com/grafana/plugin-ci-workflows/issues/457)) ([82db8bf](https://github.com/grafana/plugin-ci-workflows/commit/82db8bf61eeea38e68bad0d74262676f0cf6237d))


### 🔧 Chores

* **deps:** update actions/create-github-app-token action to v2.2.1 ([#475](https://github.com/grafana/plugin-ci-workflows/issues/475)) ([eeaa07a](https://github.com/grafana/plugin-ci-workflows/commit/eeaa07abac7139e34586f25eba499b819d6b09ea))
* **deps:** update actions/setup-node action to v6 ([#362](https://github.com/grafana/plugin-ci-workflows/issues/362)) ([69022ab](https://github.com/grafana/plugin-ci-workflows/commit/69022ab90ecdfe03890fa825bb3337d9625aac4f))
* **deps:** update peter-evans/create-pull-request action to v8 ([#478](https://github.com/grafana/plugin-ci-workflows/issues/478)) ([8f4d6ea](https://github.com/grafana/plugin-ci-workflows/commit/8f4d6ea8d1cf3dbe4c7fffd951765b3ec40a0f53))
* **main:** release plugins-frontend-e2e-against-stack 1.0.1 ([#432](https://github.com/grafana/plugin-ci-workflows/issues/432)) ([d3ca365](https://github.com/grafana/plugin-ci-workflows/commit/d3ca3655e308cd68c085f42910760aa01e726b44))

## [5.0.0](https://github.com/grafana/plugin-ci-workflows/compare/ci-cd-workflows/v4.3.0...ci-cd-workflows/v5.0.0) (2025-12-17)


### ⚠ BREAKING CHANGES

* **ci:** `universal-zip-url-*` outputs are now URLs rather than paths
* **deps:** Update default Node version to 24 LTS
* **deps:** Update default golangci-lint version to 2.7.2
* **deps:** Update default Go version to 1.25
* **ci:** use pm-detect to understand nodejs package manager ([#204](https://github.com/grafana/plugin-ci-workflows/issues/204))

### 🎉 Features

* **ci:** use pm-detect to understand nodejs package manager ([#204](https://github.com/grafana/plugin-ci-workflows/issues/204)) ([e739fbe](https://github.com/grafana/plugin-ci-workflows/commit/e739fbe50419be210cc2577abf54bfe6bf31903d))


### 🐛 Bug Fixes

* **ci:** `universal-zip-url-*` outputs are now URLs rather than paths ([b5e3ec2](https://github.com/grafana/plugin-ci-workflows/commit/b5e3ec296c7478c55c2c5a826f7c23b2b805463b))
* **ci:** do not copy backend into the "dist" folder ([b5e3ec2](https://github.com/grafana/plugin-ci-workflows/commit/b5e3ec296c7478c55c2c5a826f7c23b2b805463b))
* use local exec for playwright test ([#451](https://github.com/grafana/plugin-ci-workflows/issues/451)) ([b27d62d](https://github.com/grafana/plugin-ci-workflows/commit/b27d62de5de920e3a1755b645f02a515b4c9535d))


### ✅ Tests

* **act:** add pre-generated dist mockdata for test plugins ([#417](https://github.com/grafana/plugin-ci-workflows/issues/417)) ([cb33db6](https://github.com/grafana/plugin-ci-workflows/commit/cb33db657c080f4a5d60c6ac5edbfe67fec25f30))
* add concurrency to pr-checks-test-ci workflow ([630b056](https://github.com/grafana/plugin-ci-workflows/commit/630b0567a932b5c1595d7fee1db70dd7cb1ba746))
* pin actions in pr-checks-test-ci workflow ([630b056](https://github.com/grafana/plugin-ci-workflows/commit/630b0567a932b5c1595d7fee1db70dd7cb1ba746))
* run smoke tests with act ([#389](https://github.com/grafana/plugin-ci-workflows/issues/389)) ([0f2f7c6](https://github.com/grafana/plugin-ci-workflows/commit/0f2f7c6fc427418a4367065129b398fae3e16831))
* use local repository for release please manifest tag when running act ([630b056](https://github.com/grafana/plugin-ci-workflows/commit/630b0567a932b5c1595d7fee1db70dd7cb1ba746))


### 🔧 Chores

* build and package plugin in its directory, better support for artifact prefix ([#438](https://github.com/grafana/plugin-ci-workflows/issues/438)) ([dbbcc52](https://github.com/grafana/plugin-ci-workflows/commit/dbbcc5276e91ee2be918cc9de9792f07b72deb50))
* **deps:** update actions/checkout action to v5.0.1 ([#413](https://github.com/grafana/plugin-ci-workflows/issues/413)) ([8c591cd](https://github.com/grafana/plugin-ci-workflows/commit/8c591cdfe41c0ead6a235dcdb51f5e74af15642d))
* **deps:** update actions/checkout action to v6 ([#437](https://github.com/grafana/plugin-ci-workflows/issues/437)) ([949a63c](https://github.com/grafana/plugin-ci-workflows/commit/949a63c075e5e00352872d5447919635da59c2f2))
* **deps:** update actions/checkout action to v6.0.1 ([#455](https://github.com/grafana/plugin-ci-workflows/issues/455)) ([aec5016](https://github.com/grafana/plugin-ci-workflows/commit/aec5016c9a2acddc606cb728b5abfbc53eebe6e3))
* **deps:** update actions/create-github-app-token action to v2.2.0 ([#435](https://github.com/grafana/plugin-ci-workflows/issues/435)) ([5b2407c](https://github.com/grafana/plugin-ci-workflows/commit/5b2407c17df02b33f6c4ff83de1ec363d4f44fe7))
* **deps:** update actions/setup-go action to v6.1.0 ([#436](https://github.com/grafana/plugin-ci-workflows/issues/436)) ([5f5e089](https://github.com/grafana/plugin-ci-workflows/commit/5f5e0894940c7ebcd5f0492c0cc7ead07ce0cd84))
* **deps:** Update default Go version to 1.25 ([b1f1b22](https://github.com/grafana/plugin-ci-workflows/commit/b1f1b2282f0c3d7900adc515aada375ba1ef7c45))
* **deps:** Update default golangci-lint version to 2.7.2 ([b1f1b22](https://github.com/grafana/plugin-ci-workflows/commit/b1f1b2282f0c3d7900adc515aada375ba1ef7c45))
* **deps:** Update default Node version to 24 LTS ([b1f1b22](https://github.com/grafana/plugin-ci-workflows/commit/b1f1b2282f0c3d7900adc515aada375ba1ef7c45))
* **deps:** update grafana/shared-workflows/trigger-argo-workflow action to v1.2.1 ([#433](https://github.com/grafana/plugin-ci-workflows/issues/433)) ([6c260c0](https://github.com/grafana/plugin-ci-workflows/commit/6c260c01bda5a3f3b4ae0aa46e54937d1acaaf5a))
* **deps:** update peter-evans/create-pull-request action to v7.0.9 ([#434](https://github.com/grafana/plugin-ci-workflows/issues/434)) ([55c4108](https://github.com/grafana/plugin-ci-workflows/commit/55c4108fe05536f88f323434025fe5332f2f3652))
* **deps:** update pnpm/action-setup action to v4.2.0 ([#450](https://github.com/grafana/plugin-ci-workflows/issues/450)) ([5fbf8e6](https://github.com/grafana/plugin-ci-workflows/commit/5fbf8e65a1b34e114507ad501fceff69f9818c8f))
* **deps:** update softprops/action-gh-release action to v2.5.0 ([#452](https://github.com/grafana/plugin-ci-workflows/issues/452)) ([48548e8](https://github.com/grafana/plugin-ci-workflows/commit/48548e8674b4543e59af0d4da4025e35b3d858f1))
* **deps:** update step-security/harden-runner action to v2.13.3 ([#456](https://github.com/grafana/plugin-ci-workflows/issues/456)) ([fccf74b](https://github.com/grafana/plugin-ci-workflows/commit/fccf74bf9e1223cc0ab9b0d084662850cd7d4b92))
* **main:** release plugins-frontend-e2e-against-stack 1.0.0 ([#430](https://github.com/grafana/plugin-ci-workflows/issues/430)) ([6443ba2](https://github.com/grafana/plugin-ci-workflows/commit/6443ba2a59ec8d75208cc5c04a0e47eedcda583d))

## [4.3.0](https://github.com/grafana/plugin-ci-workflows/compare/ci-cd-workflows/v4.2.0...ci-cd-workflows/v4.3.0) (2025-11-26)


### 🎉 Features

* add a more user-friendly error when pinning plugin-ci-workflows to a commit hash ([f2c945a](https://github.com/grafana/plugin-ci-workflows/commit/f2c945af2f13d37c5ebfa544b549f09fd01c56c1))
* block workflow if pinned to a commit hash rather than a tag/branch reference ([f2c945a](https://github.com/grafana/plugin-ci-workflows/commit/f2c945af2f13d37c5ebfa544b549f09fd01c56c1))


### 🐛 Bug Fixes

* trust grafana-plugins-platform-bot[bot] ([#404](https://github.com/grafana/plugin-ci-workflows/issues/404)) ([dc5bcf5](https://github.com/grafana/plugin-ci-workflows/commit/dc5bcf59949e3a962289aaa43badf3f39f536de3))


### 🔧 Chores

* add ACTIONLINT_VERSION to simplify actionlint version bumps ([f422460](https://github.com/grafana/plugin-ci-workflows/commit/f42246017310591f720cd9f250037b90052b88bc))
* add concurrency to actionlint job ([f422460](https://github.com/grafana/plugin-ci-workflows/commit/f42246017310591f720cd9f250037b90052b88bc))
* changed warning message for rolling release channel detection ([f2c945a](https://github.com/grafana/plugin-ci-workflows/commit/f2c945af2f13d37c5ebfa544b549f09fd01c56c1))

## [4.2.0](https://github.com/grafana/plugin-ci-workflows/compare/ci-cd-workflows/v4.1.0...ci-cd-workflows/v4.2.0) (2025-11-24)


### 🎉 Features

* run plugin-validator as part of CI (opt-in) ([#366](https://github.com/grafana/plugin-ci-workflows/issues/366)) ([3456883](https://github.com/grafana/plugin-ci-workflows/commit/34568839361b8e990d500e52beba150640934b95))


### 🐛 Bug Fixes

* allow dependabot and renovate to access Vault ([#381](https://github.com/grafana/plugin-ci-workflows/issues/381)) ([0e62f4d](https://github.com/grafana/plugin-ci-workflows/commit/0e62f4de4c8e4921de802cc9a70b24f616bd4f87))
* fix check-for-rolling-releases running twice when running cd workflow ([#383](https://github.com/grafana/plugin-ci-workflows/issues/383)) ([54da033](https://github.com/grafana/plugin-ci-workflows/commit/54da033a19ff10c1557f99ca6371cfae266b8d4c))
* only trust renovate and dependabot rather than all bots ([#392](https://github.com/grafana/plugin-ci-workflows/issues/392)) ([45b9f89](https://github.com/grafana/plugin-ci-workflows/commit/45b9f892824a2bec7eb2ecf6e90692b68ce09b12))


### 🔧 Chores

* add job for smoke tests matrix check and pr checks refactoring ([#386](https://github.com/grafana/plugin-ci-workflows/issues/386)) ([80ce38a](https://github.com/grafana/plugin-ci-workflows/commit/80ce38ab114e7ecc15486247dc1657efa3069376))
* **deps:** bump default Trufflehog version to v3.91.0 ([#384](https://github.com/grafana/plugin-ci-workflows/issues/384)) ([431b333](https://github.com/grafana/plugin-ci-workflows/commit/431b33329dc0b8fb542f8330145d229bd1dc5048))
* **deps:** update github artifact actions (major) ([#375](https://github.com/grafana/plugin-ci-workflows/issues/375)) ([0662937](https://github.com/grafana/plugin-ci-workflows/commit/0662937faaa02b98a54172eaee702533b502514f))
* **deps:** update googleapis/release-please-action action to v4.4.0 ([#374](https://github.com/grafana/plugin-ci-workflows/issues/374)) ([5f3d206](https://github.com/grafana/plugin-ci-workflows/commit/5f3d206183ce3d29f184219d30bd990a36138b3f))
* **deps:** update softprops/action-gh-release action to v2.4.2 ([#393](https://github.com/grafana/plugin-ci-workflows/issues/393)) ([3a6ced2](https://github.com/grafana/plugin-ci-workflows/commit/3a6ced2f38434c7d293258d7dbefa4cd7e9ef03a))
* **deps:** update step-security/harden-runner action to v2.13.2 ([#388](https://github.com/grafana/plugin-ci-workflows/issues/388)) ([3cc38b2](https://github.com/grafana/plugin-ci-workflows/commit/3cc38b22fd5bf56864555315a623c6040d7c493b))
* **main:** release plugins-release-please 1.0.1 ([#376](https://github.com/grafana/plugin-ci-workflows/issues/376)) ([80407d0](https://github.com/grafana/plugin-ci-workflows/commit/80407d0289c509581f86bd8f55b82dbf7237fc1b))
* switch GitHub Actions runners to arm64 ([#382](https://github.com/grafana/plugin-ci-workflows/issues/382)) ([38c232b](https://github.com/grafana/plugin-ci-workflows/commit/38c232ba18e13df42cba513fff2e90098a7f9156))

## [4.1.0](https://github.com/grafana/plugin-ci-workflows/compare/ci-cd-workflows/v4.0.0...ci-cd-workflows/v4.1.0) (2025-11-04)


### 🎉 Features

* add warning message when using rolling releases channel (`[@main](https://github.com/main)`) ([#365](https://github.com/grafana/plugin-ci-workflows/issues/365)) ([e3a460d](https://github.com/grafana/plugin-ci-workflows/commit/e3a460d2760d04069db3046d557f7174fa5151bd))


### 🐛 Bug Fixes

* fail when plugin version exceeds 190 chars ([#368](https://github.com/grafana/plugin-ci-workflows/issues/368)) ([a75ba23](https://github.com/grafana/plugin-ci-workflows/commit/a75ba23f00686932691dd82b64ae6eca15f9b69e))


### 🔧 Chores

* **main:** release plugins-publish-publish 2.0.1 ([#371](https://github.com/grafana/plugin-ci-workflows/issues/371)) ([82366d7](https://github.com/grafana/plugin-ci-workflows/commit/82366d73b461d28a73a14fd360798a39425fb6ef))

## [4.0.0](https://github.com/grafana/plugin-ci-workflows/compare/ci-cd-workflows/v3.2.0...ci-cd-workflows/v4.0.0) (2025-11-03)


### ⚠ BREAKING CHANGES

* move unexported actions to `internal` folder ([#338](https://github.com/grafana/plugin-ci-workflows/issues/338))

### 🎉 Features

* adding a non-required input called signature-type ([#356](https://github.com/grafana/plugin-ci-workflows/issues/356)) ([4c800de](https://github.com/grafana/plugin-ci-workflows/commit/4c800dee4b72ceff1816693f74c22fd557be27c4))


### ♻️ Code Refactoring

* move unexported actions to `internal` folder ([#338](https://github.com/grafana/plugin-ci-workflows/issues/338)) ([46a32e8](https://github.com/grafana/plugin-ci-workflows/commit/46a32e8d3c688e39345a945ca835d636e0e5c9a4))


### 🔧 Chores

* **deps:** update softprops/action-gh-release action to v2.4.1 ([#359](https://github.com/grafana/plugin-ci-workflows/issues/359)) ([f48c11a](https://github.com/grafana/plugin-ci-workflows/commit/f48c11a872a3d36624007f2bfad11d8278caeb0a))
* **deps:** update stefanzweifel/git-auto-commit-action action to v7 ([#361](https://github.com/grafana/plugin-ci-workflows/issues/361)) ([50034ea](https://github.com/grafana/plugin-ci-workflows/commit/50034eabfee9faa714b55ef3aacbd8f1c448cd1f))
* **main:** release plugins-release-please 1.0.0 ([#357](https://github.com/grafana/plugin-ci-workflows/issues/357)) ([cdbd54b](https://github.com/grafana/plugin-ci-workflows/commit/cdbd54b803f39ed605a3b2d33c2590a045aee2e0))

## [3.2.0](https://github.com/grafana/plugin-ci-workflows/compare/ci-cd-workflows/v3.1.0...ci-cd-workflows/v3.2.0) (2025-10-14)


### 🎉 Features

* add cd summary step ([#341](https://github.com/grafana/plugin-ci-workflows/issues/341)) ([cd77b4a](https://github.com/grafana/plugin-ci-workflows/commit/cd77b4afd8da3c583bea1cd2fd7532b75539ce6a))
* add smoke tests for the shared ci workflow ([#339](https://github.com/grafana/plugin-ci-workflows/issues/339)) ([f9eb48d](https://github.com/grafana/plugin-ci-workflows/commit/f9eb48d55720c5fb85db6fc6550633df266914c5))
* **cd:** add "staging" as an alias for "ops" for plugins publishing and Argo workflow trigger ([#314](https://github.com/grafana/plugin-ci-workflows/issues/314)) ([bdb0e12](https://github.com/grafana/plugin-ci-workflows/commit/bdb0e125d3a794aad6dd5fd1bdca16aff6a0ff18))


### 🐛 Bug Fixes

* **ci:** allow-unsigned package creation on 'none' environment ([#350](https://github.com/grafana/plugin-ci-workflows/issues/350)) ([c6d400a](https://github.com/grafana/plugin-ci-workflows/commit/c6d400a3580f951415a2fdf5f3f416749901b038))
* **release-please:** update only relevant examples when releasing new versions ([#333](https://github.com/grafana/plugin-ci-workflows/issues/333)) ([38cad01](https://github.com/grafana/plugin-ci-workflows/commit/38cad01cf7feacb644800727cde71b4e6344d82b))


### 🔧 Chores

* **deps:** update actions/cache action to v4.3.0 ([#349](https://github.com/grafana/plugin-ci-workflows/issues/349)) ([56fb09d](https://github.com/grafana/plugin-ci-workflows/commit/56fb09d71994c7b81bcec7f3af32dafc49ddd5e6))
* **main:** release plugins-change-plugin-scope 1.0.0 ([#345](https://github.com/grafana/plugin-ci-workflows/issues/345)) ([538ce0a](https://github.com/grafana/plugin-ci-workflows/commit/538ce0ab29faff8d933d1d24d91bc434fa1f5747))
* **main:** release plugins-publish-publish 1.2.0 ([#324](https://github.com/grafana/plugin-ci-workflows/issues/324)) ([b8016e8](https://github.com/grafana/plugin-ci-workflows/commit/b8016e870b0bcc0bd894757b05d391be92d06fcc))
* **main:** release plugins-publish-publish 2.0.0 ([#347](https://github.com/grafana/plugin-ci-workflows/issues/347)) ([3e071f5](https://github.com/grafana/plugin-ci-workflows/commit/3e071f5c9c3076f7a4ea70623a7175778f278b0a))
* **main:** release plugins-version-bump-changelog 1.0.0 ([#329](https://github.com/grafana/plugin-ci-workflows/issues/329)) ([31b57ff](https://github.com/grafana/plugin-ci-workflows/commit/31b57ff0db96742dadbf22cbffe3c7a1e185560f))
* **main:** release plugins-version-bump-changelog 1.1.0 ([#340](https://github.com/grafana/plugin-ci-workflows/issues/340)) ([d0d69c5](https://github.com/grafana/plugin-ci-workflows/commit/d0d69c5a1890ada83d93217b7e9eed9fce3dbdf9))

## [3.1.0](https://github.com/grafana/plugin-ci-workflows/compare/ci-cd-workflows/v3.0.0...ci-cd-workflows/v3.1.0) (2025-10-03)


### 🎉 Features

* **ci:** add job to check for workflows references in pr-checks.yml ([#315](https://github.com/grafana/plugin-ci-workflows/issues/315)) ([2ae9272](https://github.com/grafana/plugin-ci-workflows/commit/2ae92720c26f0b4645970cf006ab065a1dda9d8a))
* **workflows:** fail ci for failed signs on ops and prod ([#300](https://github.com/grafana/plugin-ci-workflows/issues/300)) ([25aadee](https://github.com/grafana/plugin-ci-workflows/commit/25aadeeb2091eb2ea443a25d478ef296a5ac87c5))


### 🐛 Bug Fixes

* add argo input to decouple catalog+argo ([#311](https://github.com/grafana/plugin-ci-workflows/issues/311)) ([22dfd81](https://github.com/grafana/plugin-ci-workflows/commit/22dfd818077cdfe89d2155b4c70e5310da318638))


### 🔧 Chores

* **cd:** make CD workflow fail if Argo step fails ([#313](https://github.com/grafana/plugin-ci-workflows/issues/313)) ([df4b164](https://github.com/grafana/plugin-ci-workflows/commit/df4b164c37c7b2cdf8ca2b4811019126674466b6))
* **deps:** update actions/create-github-app-token action to v2.1.4 ([#312](https://github.com/grafana/plugin-ci-workflows/issues/312)) ([9e752b9](https://github.com/grafana/plugin-ci-workflows/commit/9e752b9bd01ba0b6fcf2755ba4833aad43ece962))
* **deps:** update step-security/harden-runner action to v2.13.1 ([#295](https://github.com/grafana/plugin-ci-workflows/issues/295)) ([670e07f](https://github.com/grafana/plugin-ci-workflows/commit/670e07f83392185c36009648b10e94037fc200c0))

## [3.0.0](https://github.com/grafana/plugin-ci-workflows/compare/ci-cd-workflows/v2.0.0...ci-cd-workflows/v3.0.0) (2025-09-26)


### ⚠ BREAKING CHANGES

* **deps:** Bump Go version used to build plugin from Go 1.23 to Go 1.24, bump golangci-lint to v1.64.8 ([#292](https://github.com/grafana/plugin-ci-workflows/issues/292))
* **deps:** Bump default Node.js version used to build plugin from 20 to 22 (Active LTS) ([#291](https://github.com/grafana/plugin-ci-workflows/issues/291))

### 🐛 Bug Fixes

* add prod-canary ([#280](https://github.com/grafana/plugin-ci-workflows/issues/280)) ([df43691](https://github.com/grafana/plugin-ci-workflows/commit/df43691c48b0f5a0b3b9bcf6c94c63aa8939c9bf))
* bump e2e-versions to get plugin.json lookup fix  ([#307](https://github.com/grafana/plugin-ci-workflows/issues/307)) ([4fc11cc](https://github.com/grafana/plugin-ci-workflows/commit/4fc11cc331c32e47000444319c4cc70899bcc700))
* **worflows/ci:** make .nvmrc be used over the default node version ([#299](https://github.com/grafana/plugin-ci-workflows/issues/299)) ([2908cea](https://github.com/grafana/plugin-ci-workflows/commit/2908cea8876d7c26e63bdbd3458d28f9681712d1))


### 🤖 Continuous Integrations

* use conventional commits title for PR that restores rolling release ([#289](https://github.com/grafana/plugin-ci-workflows/issues/289)) ([1f78e66](https://github.com/grafana/plugin-ci-workflows/commit/1f78e66915b8f734504e24299b5eb94126a1949e))


### 🔧 Chores

* add zizmor config and ignore "unpinned-uses" for plugin-ci-workflows references ([#285](https://github.com/grafana/plugin-ci-workflows/issues/285)) ([1ad04f9](https://github.com/grafana/plugin-ci-workflows/commit/1ad04f98e1588586591a9f9b38289e1a6073666a))
* **deps:** Bump default Node.js version used to build plugin from 20 to 22 (Active LTS) ([#291](https://github.com/grafana/plugin-ci-workflows/issues/291)) ([080c28b](https://github.com/grafana/plugin-ci-workflows/commit/080c28b04265b4c7333b533e797e951dfb0c3b66))
* **deps:** Bump Go version used to build plugin from Go 1.23 to Go 1.24, bump golangci-lint to v1.64.8 ([#292](https://github.com/grafana/plugin-ci-workflows/issues/292)) ([124ff1c](https://github.com/grafana/plugin-ci-workflows/commit/124ff1cc9e488f2b2b630453a372f8fb0e2e6d79))
* **deps:** Bump Trufflehog to v3.90.6 ([9c039a5](https://github.com/grafana/plugin-ci-workflows/commit/9c039a5e04bb2bbd6f665bf199de0185728f1a6a))
* **deps:** update actions/checkout action to v5 ([#286](https://github.com/grafana/plugin-ci-workflows/issues/286)) ([74a9e53](https://github.com/grafana/plugin-ci-workflows/commit/74a9e531e203fe5abf6dcef4d509857a4c976247))
* manually bumping version of e2e-versions action ([#304](https://github.com/grafana/plugin-ci-workflows/issues/304)) ([2d617a5](https://github.com/grafana/plugin-ci-workflows/commit/2d617a5adc492cae4668d1455d7ac8f2592de0c7))
* pinning e2e-versions and wait-for-grafana ([#301](https://github.com/grafana/plugin-ci-workflows/issues/301)) ([017fa07](https://github.com/grafana/plugin-ci-workflows/commit/017fa07ea57b0bcdd0c0dc33bc3c5da6ae300d82))

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
