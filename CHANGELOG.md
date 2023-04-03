# Changelog

## [0.1.8](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.1.7...v0.1.8) (2023-04-03)


### Features

* add password flag for backend authentication ([#199](https://github.com/k8sgpt-ai/k8sgpt/issues/199)) ([075a940](https://github.com/k8sgpt-ai/k8sgpt/commit/075a940d2c9bdd8aa9162940ed46abad47d46998))
* adding shields to readme ([213ecd8](https://github.com/k8sgpt-ai/k8sgpt/commit/213ecd8e83933fabaa5d3d674c67958599dd72ce))
* adding unit testing and example ([35b838b](https://github.com/k8sgpt-ai/k8sgpt/commit/35b838bfafa248dbf3932c7a3ee708b1a1539f18))
* alias filter to filters ([dde4e83](https://github.com/k8sgpt-ai/k8sgpt/commit/dde4e833b0e87553dea4e5c1e17a14e303956bc1))
* analyzer ifacing ([426f562](https://github.com/k8sgpt-ai/k8sgpt/commit/426f562be83ed0e708a07b9e1900ac06fa017c27))
* service test ([44cc8f7](https://github.com/k8sgpt-ai/k8sgpt/commit/44cc8f7ad68d152ec577e57cab7d8d9ab9613378))
* test workflow ([5f30a4d](https://github.com/k8sgpt-ai/k8sgpt/commit/5f30a4ddf44ebff949bb0573f261667539a2dcfb))


### Bug Fixes

* **deps:** update module github.com/sashabaranov/go-openai to v1.5.8 ([91fb065](https://github.com/k8sgpt-ai/k8sgpt/commit/91fb06530a21259da6e72c28342e743d2b481294))


### Other

* create linux packages ([#201](https://github.com/k8sgpt-ai/k8sgpt/issues/201)) ([67753be](https://github.com/k8sgpt-ai/k8sgpt/commit/67753be6f317c462ebe1d9a316f2b0c9684ca4e5))
* **deps:** pin dependencies ([#198](https://github.com/k8sgpt-ai/k8sgpt/issues/198)) ([f8291aa](https://github.com/k8sgpt-ai/k8sgpt/commit/f8291aab085209f9fee13a6c92c96076163e2e90))

## [0.1.7](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.1.6...v0.1.7) (2023-04-02)


### Features

* add hpa analyzer and init additionalAnalyzers ([3603872](https://github.com/k8sgpt-ai/k8sgpt/commit/360387249feb9a999286aaa874a13007986219a5))
* add pda analyzer ([532a5ce](https://github.com/k8sgpt-ai/k8sgpt/commit/532a5ce0332a8466df42bc944800e6668e349801))
* check if ScaleTargetRef is possible option ([5dad75f](https://github.com/k8sgpt-ai/k8sgpt/commit/5dad75fbe9fd15cfa7bfa69c046b851ea905876f))


### Bug Fixes

* hpaAnalyzer analysis result is using wrong parent ([1190fe6](https://github.com/k8sgpt-ai/k8sgpt/commit/1190fe60fdd6e66ce435874628039df7047a52b9))
* spelling of PodDisruptionBudget ([ceff008](https://github.com/k8sgpt-ai/k8sgpt/commit/ceff0084df1b6de16f1ed503ee8a4b3c1a9f8648))
* update client API call to use StatefulSet instead of Deployment ([4916fef](https://github.com/k8sgpt-ai/k8sgpt/commit/4916fef9d6b75c54bcfbc5d136550018e96e3632))


### Refactoring

* merged main into branch ([3e836d8](https://github.com/k8sgpt-ai/k8sgpt/commit/3e836d81b7c33ce5c0c133c2e1ca3b0c8d3eeeb0)), closes [#101](https://github.com/k8sgpt-ai/k8sgpt/issues/101)


### Other

* **deps:** update anchore/sbom-action action to v0.14.1 ([80f29da](https://github.com/k8sgpt-ai/k8sgpt/commit/80f29dae4fd6f6348967192ce2f51f0e0fb5dea0))
* merge branch 'chetanguptaa-some-fixes' ([071ee56](https://github.com/k8sgpt-ai/k8sgpt/commit/071ee560f36b64b4c65274181e2d13bb14d5b914))
* refine renovate config ([#172](https://github.com/k8sgpt-ai/k8sgpt/issues/172)) ([d23da9a](https://github.com/k8sgpt-ai/k8sgpt/commit/d23da9ae836a07f0fd59c20a1c3c71d6b7f75277))
* removes bar on normal analyze events ([e1d8992](https://github.com/k8sgpt-ai/k8sgpt/commit/e1d89920b097db4417c55b020fb23dd8cbaf19ed))
* removes bar on normal analyze events ([96d0d75](https://github.com/k8sgpt-ai/k8sgpt/commit/96d0d754eab67c0742d3a36a1eefb9c28df59e96))
* update dependencies ([#174](https://github.com/k8sgpt-ai/k8sgpt/issues/174)) ([9d9c262](https://github.com/k8sgpt-ai/k8sgpt/commit/9d9c26214fbb4c4faba7ef85f2204bc961396de8))


### Docs

* add pdbAnalyzer as optional analyzer ([f6974d0](https://github.com/k8sgpt-ai/k8sgpt/commit/f6974d07581384e260059f121242854320dfc58b))

## [0.1.6](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.1.5...v0.1.6) (2023-03-31)


### Bug Fixes

* analysis detail not displayed when --explain ([869ba90](https://github.com/k8sgpt-ai/k8sgpt/commit/869ba909075a5543413fb6ae7fc79aa067c08da4))

## [0.1.5](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.1.4...v0.1.5) (2023-03-31)


### Features

* add & remove default filter(s) to analyze. ([32ddf66](https://github.com/k8sgpt-ai/k8sgpt/commit/32ddf6691ce083fd4283a1d5ac4b9f02e90df867))
* add filter command add "list" subcommand ([#159](https://github.com/k8sgpt-ai/k8sgpt/issues/159)) ([6e17c9e](https://github.com/k8sgpt-ai/k8sgpt/commit/6e17c9e285e3871bb8f694b734a8cd6fd02e60f0))
* check if filters does not empty on add & remove ([975813d](https://github.com/k8sgpt-ai/k8sgpt/commit/975813d3284719c877630ad20f90c6fe163283da))
* remove filter prefix on subcommand ([30faf84](https://github.com/k8sgpt-ai/k8sgpt/commit/30faf842541c0be6b6483f71f6cf04d5cafecef5))
* rework filters ([3ed545f](https://github.com/k8sgpt-ai/k8sgpt/commit/3ed545f33fb3ecb3827c03e8c89027c61386c44f))
* update filters add & remove to be more consistent ([9aa0e89](https://github.com/k8sgpt-ai/k8sgpt/commit/9aa0e8960ee340208b4749954c99867842ba58b9))


### Bug Fixes

* kubecontext flag has no effect ([a8bf451](https://github.com/k8sgpt-ai/k8sgpt/commit/a8bf45134ff3a72dc3e531d720f119790faff9d4))
* spelling on dupplicateFilters ([0a12448](https://github.com/k8sgpt-ai/k8sgpt/commit/0a124484a23789376258413e73628c7b1d7abded))


### Other

* renamed filter list file ([25f8dc3](https://github.com/k8sgpt-ai/k8sgpt/commit/25f8dc390cccd66965993f464351e671af11f8ac))

## [0.1.4](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.1.3...v0.1.4) (2023-03-30)


### Features

* add Ingress class validation ([#154](https://github.com/k8sgpt-ai/k8sgpt/issues/154)) ([b061566](https://github.com/k8sgpt-ai/k8sgpt/commit/b061566404ef80288ca29add2d401574109d44c0))
* output selected backend ([#153](https://github.com/k8sgpt-ai/k8sgpt/issues/153)) ([be061da](https://github.com/k8sgpt-ai/k8sgpt/commit/be061da5b65045938acd70ad2eb2d21b87d2d6bf))


### Bug Fixes

* now supports different kubeconfig and kubectx ([c8f3c94](https://github.com/k8sgpt-ai/k8sgpt/commit/c8f3c946b00c00cd185961a4fa777806da94014e))


### Refactoring

* removed sample flag ([0afd528](https://github.com/k8sgpt-ai/k8sgpt/commit/0afd52844b96579391f77698bf0555145b6d2be8))

## [0.1.3](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.1.2...v0.1.3) (2023-03-30)


### Features

* add secret validation to ingress analyzer ([#141](https://github.com/k8sgpt-ai/k8sgpt/issues/141)) ([86c7e81](https://github.com/k8sgpt-ai/k8sgpt/commit/86c7e81e18db02ebcbfe35d470682c982871375f))
* bugfix for output ([2eab0c5](https://github.com/k8sgpt-ai/k8sgpt/commit/2eab0c544fbb6026f6aea79b08d8f29c061acf2e))
* CODE_OF_CONDUCT.md ([#129](https://github.com/k8sgpt-ai/k8sgpt/issues/129)) ([fe73633](https://github.com/k8sgpt-ai/k8sgpt/commit/fe73633273c5c1f4188bca48471283535967d5aa))
* create-security.md ([27b8916](https://github.com/k8sgpt-ai/k8sgpt/commit/27b8916f297570907437686c6d958636fb249d50))
* improvement to analysis speed ([548039e](https://github.com/k8sgpt-ai/k8sgpt/commit/548039ebe62bb609c1aa288e5e49845850fd2dd8))
* init ingress analyzer ([#138](https://github.com/k8sgpt-ai/k8sgpt/issues/138)) ([fe683b7](https://github.com/k8sgpt-ai/k8sgpt/commit/fe683b71b84fe82459b0ffe366b4dcfa1c978cfe))


### Bug Fixes

* add Ingress in GetParent switch case ([14ba8d5](https://github.com/k8sgpt-ai/k8sgpt/commit/14ba8d555005f31fc2201cb8b61653093c19b8a7))
* bugfix for output ([#148](https://github.com/k8sgpt-ai/k8sgpt/issues/148)) ([172c2df](https://github.com/k8sgpt-ai/k8sgpt/commit/172c2df6c55f5fddbfec7f8526be5f2323d1b900))
* Change ObjectMeta value in Ingress analyser. ([bf49a51](https://github.com/k8sgpt-ai/k8sgpt/commit/bf49a51c62af450cff51a590547ef30989bd2e93))
* typo in description of the filter flag in analyze command ([#147](https://github.com/k8sgpt-ai/k8sgpt/issues/147)) ([f4765be](https://github.com/k8sgpt-ai/k8sgpt/commit/f4765bed1b1ad121a81b35878fdb866354b5e34a))


### Other

* **deps:** update google-github-actions/release-please-action digest to ee9822e ([#132](https://github.com/k8sgpt-ai/k8sgpt/issues/132)) ([01b2826](https://github.com/k8sgpt-ai/k8sgpt/commit/01b282647512a4eaebd42ab5847b5534de148d14))


### Docs

* add new slack link ([#134](https://github.com/k8sgpt-ai/k8sgpt/issues/134)) ([#135](https://github.com/k8sgpt-ai/k8sgpt/issues/135)) ([cad2b38](https://github.com/k8sgpt-ai/k8sgpt/commit/cad2b38d037658495024ec0166ebd3e936f65c2e))

## [0.1.2](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.1.1...v0.1.2) (2023-03-28)


### Features

* added namespace filter ([#127](https://github.com/k8sgpt-ai/k8sgpt/issues/127)) ([b78ab3d](https://github.com/k8sgpt-ai/k8sgpt/commit/b78ab3d9b503a256bf6ccf18276e20140ae17d1c))
* prefix templates ([#125](https://github.com/k8sgpt-ai/k8sgpt/issues/125)) ([65a568e](https://github.com/k8sgpt-ai/k8sgpt/commit/65a568e937a8fdacc179f5e8b1a021a0178c04f0))


### Bug Fixes

* readme code blocks ([#126](https://github.com/k8sgpt-ai/k8sgpt/issues/126)) ([c8b92aa](https://github.com/k8sgpt-ai/k8sgpt/commit/c8b92aaa0e2795aa8d65f84277c8adfe0f1d14e3))
* update README.md ([#119](https://github.com/k8sgpt-ai/k8sgpt/issues/119)) ([05abe97](https://github.com/k8sgpt-ai/k8sgpt/commit/05abe975dd859cd85096a1a7182f17b0437ad20f))


### Other

* added default issue template ([#96](https://github.com/k8sgpt-ai/k8sgpt/issues/96)) ([#121](https://github.com/k8sgpt-ai/k8sgpt/issues/121)) ([11c227b](https://github.com/k8sgpt-ai/k8sgpt/commit/11c227b82e16dac8b46cbd03bb04d9cc1c2b5ac3))


### Docs

* add new issue templates ([dbd305f](https://github.com/k8sgpt-ai/k8sgpt/commit/dbd305f901cca09b7148254c3aa7a7435504d6cc))
* add WSL gcc instructions ([4d5566b](https://github.com/k8sgpt-ai/k8sgpt/commit/4d5566b4df7aedf43edbeeb03130f0ba77dbed1a))
* added Windows and Linux instalation steps in README ([#116](https://github.com/k8sgpt-ai/k8sgpt/issues/116)) ([3bfb278](https://github.com/k8sgpt-ai/k8sgpt/commit/3bfb278f81a9c550ee37a88c0cb0377331802542))
* fix indentations ([a46416d](https://github.com/k8sgpt-ai/k8sgpt/commit/a46416dce0f5cee2d42b27525023b04af1a8e3c0))
* rename ISSUE_TEMPLATE ([#124](https://github.com/k8sgpt-ai/k8sgpt/issues/124)) ([cb4932c](https://github.com/k8sgpt-ai/k8sgpt/commit/cb4932c39df4903a4b48ae5f0428860027f76fd2))

## [0.1.1](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.1.0...v0.1.1) (2023-03-28)


### Features

* this stops service exiting the program ([6f90386](https://github.com/k8sgpt-ai/k8sgpt/commit/6f90386fc93b2e39e59832468922e8ba7210b8e5))
* updated readme ([e0141d1](https://github.com/k8sgpt-ai/k8sgpt/commit/e0141d1cf54b5b37b25a5caeb9d5c940b9410ea7))


### Bug Fixes

* short term solution for exhaustion ([5890e3a](https://github.com/k8sgpt-ai/k8sgpt/commit/5890e3a79c80a2973af2feb7d50e7f9c57c563c2))


### Other

* update README.md ([93b947f](https://github.com/k8sgpt-ai/k8sgpt/commit/93b947f261e401c10dde6dc1854e6e22187437d6))
* update root.go path ([2cb1c9c](https://github.com/k8sgpt-ai/k8sgpt/commit/2cb1c9c150d052bb3942d9f62ded9d54b0e1873e))

## [0.1.0](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.0.9...v0.1.0) (2023-03-28)


### Features

* added british alias ([39c0444](https://github.com/k8sgpt-ai/k8sgpt/commit/39c0444fac9b46d0faa347b45df779b97019e5b6)), closes [#93](https://github.com/k8sgpt-ai/k8sgpt/issues/93)
* enables overwriting of cache ([#95](https://github.com/k8sgpt-ai/k8sgpt/issues/95)) ([a270f7c](https://github.com/k8sgpt-ai/k8sgpt/commit/a270f7c89fb8bec35984715c5e4d160a2307e678))


### Other

* add CODEOWNERS ([c5c6162](https://github.com/k8sgpt-ai/k8sgpt/commit/c5c6162df1f3701659e47bce6e9fc6e3c569e539))
* add codeowners file ([#102](https://github.com/k8sgpt-ai/k8sgpt/issues/102)) ([829ff56](https://github.com/k8sgpt-ai/k8sgpt/commit/829ff566c0a964250d3d8d45306d410e1b9d9d35))
* release 0.1.0 ([f9c7daf](https://github.com/k8sgpt-ai/k8sgpt/commit/f9c7daf3dcd06dcd9cea5603108b8a42ee273348))

## [0.0.9](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.0.8...v0.0.9) (2023-03-28)


### Other

* small update ([202e8e2](https://github.com/k8sgpt-ai/k8sgpt/commit/202e8e2977422b2b4506a80dc9b76a392c5457eb))

## [0.0.8](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.0.7...v0.0.8) (2023-03-27)


### Features

* add generation of api-keys to cli ([#87](https://github.com/k8sgpt-ai/k8sgpt/issues/87)) ([1c653ec](https://github.com/k8sgpt-ai/k8sgpt/commit/1c653ecc51b74a2f51ce7240ffaee0fe75f2e8dd))
* add generation of api-keys to cli ([#87](https://github.com/k8sgpt-ai/k8sgpt/issues/87)) ([bb2db5c](https://github.com/k8sgpt-ai/k8sgpt/commit/bb2db5ca7923e2049308d1674bb59ae8154e415c))
* addition of simple language support ([c3008c5](https://github.com/k8sgpt-ai/k8sgpt/commit/c3008c5e75acbb35d864135199ca9c034f59e35f))
* version ([0c231d6](https://github.com/k8sgpt-ai/k8sgpt/commit/0c231d635e7ad71609bb80abac5e0ade15ffb860))
* version ([931f072](https://github.com/k8sgpt-ai/k8sgpt/commit/931f072e0ab0cfd77f261b0b719cf0819f85b951))

## [0.0.7](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.0.6...v0.0.7) (2023-03-27)


### Features

* wip fixing missing details ([0852c65](https://github.com/k8sgpt-ai/k8sgpt/commit/0852c658ded33b91e1d323bd8cba6ac6935cb525))


### Other

* moved code ([a194d4a](https://github.com/k8sgpt-ai/k8sgpt/commit/a194d4a509329cbc5a00724b0a19c75726c2a0d3))
* return success on no issues ([009f47c](https://github.com/k8sgpt-ai/k8sgpt/commit/009f47c8e8ee6d3ce9b36110c36edae97690c949))
* updated readme ([06fb807](https://github.com/k8sgpt-ai/k8sgpt/commit/06fb8073dc5b0b5bd9f8d115d9ec206ab238d68f))

## [0.0.6](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.0.6...v0.0.6) (2023-03-26)


### Features

* add service analysis ([961fb6c](https://github.com/k8sgpt-ai/k8sgpt/commit/961fb6c555f59f1276531f462739b76b1508830e))
* added analysis for pvcs ([88d49ae](https://github.com/k8sgpt-ai/k8sgpt/commit/88d49ae21c7d889d59361de157360f80503683be))
* also fixes bug if the events feed is empty ([#73](https://github.com/k8sgpt-ai/k8sgpt/issues/73)) ([a1093dc](https://github.com/k8sgpt-ai/k8sgpt/commit/a1093dcfe468a7671c9e543372f73780fb38418e))
* build container ([260640f](https://github.com/k8sgpt-ai/k8sgpt/commit/260640f865baefba8ac256f800d4992f25ca15fd))
* find parent objects ([b29c6e4](https://github.com/k8sgpt-ai/k8sgpt/commit/b29c6e45825807d07dd6fdb954457772f40b1b0e))
* find parent objects and add information about them ([#72](https://github.com/k8sgpt-ai/k8sgpt/issues/72)) ([14e85b0](https://github.com/k8sgpt-ai/k8sgpt/commit/14e85b08ff7d9a571796905260db7f1056b6e838))
* find replicaset errors ([8ac56e0](https://github.com/k8sgpt-ai/k8sgpt/commit/8ac56e062baef2a0cf7c7ce2b4c97753f079f157))
* initial json implementation ([#68](https://github.com/k8sgpt-ai/k8sgpt/issues/68)) ([979f13f](https://github.com/k8sgpt-ai/k8sgpt/commit/979f13f043f54a5bc74d0a49fee0db2faaf0a4f8))
* interfaced out ai clients ([90b3c08](https://github.com/k8sgpt-ai/k8sgpt/commit/90b3c0898c8ab1299ce8b60effe981f5fc9ed63b))
* support for multi-auth ([51aa59a](https://github.com/k8sgpt-ai/k8sgpt/commit/51aa59aea8c0fd5533d2300c7a79c0b9008ef887))
* updated readme ([7336924](https://github.com/k8sgpt-ai/k8sgpt/commit/73369240b4fc8c91dae0ae272e671f7b413e3bdc))


### Bug Fixes

* add permissions to read repository ([d6cc4cf](https://github.com/k8sgpt-ai/k8sgpt/commit/d6cc4cfcbffbf84f27c7e4e4159da1e42dd5d689))
* build ([1fbed3e](https://github.com/k8sgpt-ai/k8sgpt/commit/1fbed3e44ff790fccfef502ddafae92e34629c21))
* container naming ([115276e](https://github.com/k8sgpt-ai/k8sgpt/commit/115276e01a38fc1692d6b66ab56a33f1e1793974))
* **deps:** update module github.com/sashabaranov/go-openai to v1.5.5 ([105fe44](https://github.com/k8sgpt-ai/k8sgpt/commit/105fe44680e5a987d4a65ff9c58b5b2211808c5e))
* **deps:** update module github.com/sashabaranov/go-openai to v1.5.6 ([37a1d3f](https://github.com/k8sgpt-ai/k8sgpt/commit/37a1d3f47e07caddb168f228627973870a9d867e))
* **deps:** update module github.com/sashabaranov/go-openai to v1.5.7 ([7f7726d](https://github.com/k8sgpt-ai/k8sgpt/commit/7f7726d59a63baeaf8ff110e00b30a20ec7f1df5))
* minor adaptions ([ef17b84](https://github.com/k8sgpt-ai/k8sgpt/commit/ef17b845ba3c65c16ed5dcc417e3e3d3d40dd04e))
* missing parent when explain is used ([9c7d559](https://github.com/k8sgpt-ai/k8sgpt/commit/9c7d55955b777ad201307cb946e0fc81cf9c4b99))
* release please config ([c402c7b](https://github.com/k8sgpt-ai/k8sgpt/commit/c402c7bab7baababbbc7c82965d8337de7d50d35))
* remove sboms from goreleaser ([addc01f](https://github.com/k8sgpt-ai/k8sgpt/commit/addc01f700dd2ea31ec24dcf4995bb7ed4a4785e))
* semantic commit token permission ([#69](https://github.com/k8sgpt-ai/k8sgpt/issues/69)) ([0181c0a](https://github.com/k8sgpt-ai/k8sgpt/commit/0181c0aeb56ad82fd232ce1c7788c43b7bd03bf2))


### Docs

* add some important information to contributing ([9ab7f58](https://github.com/k8sgpt-ai/k8sgpt/commit/9ab7f587620d69e4e8fc98faabce6417c35f7497))
* update CONTRIBUTING ([05a787d](https://github.com/k8sgpt-ai/k8sgpt/commit/05a787d53dfe5e625c6449ac1e21ec36e66ddd28))
* update CONTRIBUTING ([26449e1](https://github.com/k8sgpt-ai/k8sgpt/commit/26449e10efd8926cccd4a2eaa4e9dc3afa8bd01a))


### Other

* add bot secret to goreleaser ([171e58b](https://github.com/k8sgpt-ai/k8sgpt/commit/171e58b51107f75717694e35c4e249ee41f0409a))
* add brew tap generation on release ([2992c4e](https://github.com/k8sgpt-ai/k8sgpt/commit/2992c4e5c8abad50c90ed85523c732f19ab1f31c))
* add initial renovate config ([e37dbc7](https://github.com/k8sgpt-ai/k8sgpt/commit/e37dbc7909f1c520c4c6660c25b45de5847ea581))
* add pull request template ([a6d5132](https://github.com/k8sgpt-ai/k8sgpt/commit/a6d5132b8c2ff077680e2edfd8361a93008197fd))
* add release-please ([da7b409](https://github.com/k8sgpt-ai/k8sgpt/commit/da7b40978d55a6afed4c3a1ca83a756238feaca8))
* add semantic pr validation ([#66](https://github.com/k8sgpt-ai/k8sgpt/issues/66)) ([ad594c7](https://github.com/k8sgpt-ai/k8sgpt/commit/ad594c7cb2105e0eff72d1767b2ddcc4dc0e3d38))
* change module repo ([a307c13](https://github.com/k8sgpt-ai/k8sgpt/commit/a307c132b3464ff2e949c8a5588e01d344de91a0))
* **deps:** pin amannn/action-semantic-pull-request action to c3cd5d1 ([3621766](https://github.com/k8sgpt-ai/k8sgpt/commit/36217667ceb87d9b97b44dc91e0ff6e7a1b86e14))
* **deps:** pin dependencies ([f6072f5](https://github.com/k8sgpt-ai/k8sgpt/commit/f6072f56cbe2c073b7b7ebef6c12fa98120e54e2))
* **deps:** pin dependencies ([5b360de](https://github.com/k8sgpt-ai/k8sgpt/commit/5b360de2ae6094cf850a4ae973a22855c21a9040))
* **deps:** pin dependencies ([7fea7d1](https://github.com/k8sgpt-ai/k8sgpt/commit/7fea7d14a572fe0fd05f5f241b98e93655fb1965))
* **deps:** update actions/checkout digest to 8f4b7f8 ([9955d75](https://github.com/k8sgpt-ai/k8sgpt/commit/9955d754505b60f28d17397132a1d02e95ffe303))
* **main:** release 0.0.3 ([53c9947](https://github.com/k8sgpt-ai/k8sgpt/commit/53c994725ea2c2c54898ffe5307d9df40e9c1fe5))
* **main:** release 0.0.3 ([f5d8609](https://github.com/k8sgpt-ai/k8sgpt/commit/f5d86092f49faef8d71cb950986d76c3f92daf46))
* **main:** release 0.0.3 ([22873a6](https://github.com/k8sgpt-ai/k8sgpt/commit/22873a67163e58484d2a0ad343b4ba3c83e51d8f))
* **main:** release 0.0.4 ([13b7d58](https://github.com/k8sgpt-ai/k8sgpt/commit/13b7d58e590078f086a0af2f9d1800e0e65a28bb))
* **main:** release 0.0.4 ([aef7256](https://github.com/k8sgpt-ai/k8sgpt/commit/aef7256dc3a85817573744f8b4a54f834368bac7))
* **main:** release 0.0.4 ([6dbcde9](https://github.com/k8sgpt-ai/k8sgpt/commit/6dbcde94e961a6e5a1fc0559d2a1da5567a659de))
* **main:** release 0.0.5 ([9fecc1e](https://github.com/k8sgpt-ai/k8sgpt/commit/9fecc1ea6df4104412fc1230372de6f26aa1ade2))
* **main:** release 0.0.6 ([d554bba](https://github.com/k8sgpt-ai/k8sgpt/commit/d554bba38494745f83b5a8931f665429af35a31a))
* release 0.0.3 ([4840aa0](https://github.com/k8sgpt-ai/k8sgpt/commit/4840aa081e3aa4a7a01fd3fd5f837fa6f0c3c02c))
* release 0.0.3 ([de02795](https://github.com/k8sgpt-ai/k8sgpt/commit/de027955ea18a751c5f991e7ff0f60b90ae704b0))
* release 0.0.3 ([a927c32](https://github.com/k8sgpt-ai/k8sgpt/commit/a927c32def806bb8b99e1cfcd4ee3dcdeca6ae5d))
* release 0.0.4 ([08f2c31](https://github.com/k8sgpt-ai/k8sgpt/commit/08f2c3112e2cc16b49b9cf8fdbd97368acecc754))
* release 0.0.5 ([8da8945](https://github.com/k8sgpt-ai/k8sgpt/commit/8da8945d1b8d898440be235f88bdb2c08b0f9f84))
* release 0.0.6 ([dc2bfa9](https://github.com/k8sgpt-ai/k8sgpt/commit/dc2bfa918c080a6c1b2e5ef66d699d9e08e28e10))

## [0.0.6](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.0.5...v0.0.6) (2023-03-26)


### Features

* add service analysis ([961fb6c](https://github.com/k8sgpt-ai/k8sgpt/commit/961fb6c555f59f1276531f462739b76b1508830e))
* added analysis for pvcs ([88d49ae](https://github.com/k8sgpt-ai/k8sgpt/commit/88d49ae21c7d889d59361de157360f80503683be))
* also fixes bug if the events feed is empty ([#73](https://github.com/k8sgpt-ai/k8sgpt/issues/73)) ([a1093dc](https://github.com/k8sgpt-ai/k8sgpt/commit/a1093dcfe468a7671c9e543372f73780fb38418e))
* find parent objects ([b29c6e4](https://github.com/k8sgpt-ai/k8sgpt/commit/b29c6e45825807d07dd6fdb954457772f40b1b0e))
* find parent objects and add information about them ([#72](https://github.com/k8sgpt-ai/k8sgpt/issues/72)) ([14e85b0](https://github.com/k8sgpt-ai/k8sgpt/commit/14e85b08ff7d9a571796905260db7f1056b6e838))
* initial json implementation ([#68](https://github.com/k8sgpt-ai/k8sgpt/issues/68)) ([979f13f](https://github.com/k8sgpt-ai/k8sgpt/commit/979f13f043f54a5bc74d0a49fee0db2faaf0a4f8))
* interfaced out ai clients ([90b3c08](https://github.com/k8sgpt-ai/k8sgpt/commit/90b3c0898c8ab1299ce8b60effe981f5fc9ed63b))
* support for multi-auth ([51aa59a](https://github.com/k8sgpt-ai/k8sgpt/commit/51aa59aea8c0fd5533d2300c7a79c0b9008ef887))


### Bug Fixes

* missing parent when explain is used ([9c7d559](https://github.com/k8sgpt-ai/k8sgpt/commit/9c7d55955b777ad201307cb946e0fc81cf9c4b99))
* semantic commit token permission ([#69](https://github.com/k8sgpt-ai/k8sgpt/issues/69)) ([0181c0a](https://github.com/k8sgpt-ai/k8sgpt/commit/0181c0aeb56ad82fd232ce1c7788c43b7bd03bf2))


### Other

* add semantic pr validation ([#66](https://github.com/k8sgpt-ai/k8sgpt/issues/66)) ([ad594c7](https://github.com/k8sgpt-ai/k8sgpt/commit/ad594c7cb2105e0eff72d1767b2ddcc4dc0e3d38))
* **deps:** pin amannn/action-semantic-pull-request action to c3cd5d1 ([3621766](https://github.com/k8sgpt-ai/k8sgpt/commit/36217667ceb87d9b97b44dc91e0ff6e7a1b86e14))

## [0.0.5](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.0.4...v0.0.5) (2023-03-24)


### Other

* release 0.0.5 ([8da8945](https://github.com/k8sgpt-ai/k8sgpt/commit/8da8945d1b8d898440be235f88bdb2c08b0f9f84))

## [0.0.4](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.0.4...v0.0.4) (2023-03-24)


### Features

* build container ([260640f](https://github.com/k8sgpt-ai/k8sgpt/commit/260640f865baefba8ac256f800d4992f25ca15fd))
* find replicaset errors ([8ac56e0](https://github.com/k8sgpt-ai/k8sgpt/commit/8ac56e062baef2a0cf7c7ce2b4c97753f079f157))


### Bug Fixes

* add permissions to read repository ([d6cc4cf](https://github.com/k8sgpt-ai/k8sgpt/commit/d6cc4cfcbffbf84f27c7e4e4159da1e42dd5d689))
* build ([1fbed3e](https://github.com/k8sgpt-ai/k8sgpt/commit/1fbed3e44ff790fccfef502ddafae92e34629c21))
* container naming ([115276e](https://github.com/k8sgpt-ai/k8sgpt/commit/115276e01a38fc1692d6b66ab56a33f1e1793974))
* **deps:** update module github.com/sashabaranov/go-openai to v1.5.5 ([105fe44](https://github.com/k8sgpt-ai/k8sgpt/commit/105fe44680e5a987d4a65ff9c58b5b2211808c5e))
* **deps:** update module github.com/sashabaranov/go-openai to v1.5.6 ([37a1d3f](https://github.com/k8sgpt-ai/k8sgpt/commit/37a1d3f47e07caddb168f228627973870a9d867e))
* **deps:** update module github.com/sashabaranov/go-openai to v1.5.7 ([7f7726d](https://github.com/k8sgpt-ai/k8sgpt/commit/7f7726d59a63baeaf8ff110e00b30a20ec7f1df5))
* minor adaptions ([ef17b84](https://github.com/k8sgpt-ai/k8sgpt/commit/ef17b845ba3c65c16ed5dcc417e3e3d3d40dd04e))
* release please config ([c402c7b](https://github.com/k8sgpt-ai/k8sgpt/commit/c402c7bab7baababbbc7c82965d8337de7d50d35))
* remove sboms from goreleaser ([addc01f](https://github.com/k8sgpt-ai/k8sgpt/commit/addc01f700dd2ea31ec24dcf4995bb7ed4a4785e))


### Docs

* add some important information to contributing ([9ab7f58](https://github.com/k8sgpt-ai/k8sgpt/commit/9ab7f587620d69e4e8fc98faabce6417c35f7497))
* update CONTRIBUTING ([05a787d](https://github.com/k8sgpt-ai/k8sgpt/commit/05a787d53dfe5e625c6449ac1e21ec36e66ddd28))
* update CONTRIBUTING ([26449e1](https://github.com/k8sgpt-ai/k8sgpt/commit/26449e10efd8926cccd4a2eaa4e9dc3afa8bd01a))


### Other

* add bot secret to goreleaser ([171e58b](https://github.com/k8sgpt-ai/k8sgpt/commit/171e58b51107f75717694e35c4e249ee41f0409a))
* add brew tap generation on release ([2992c4e](https://github.com/k8sgpt-ai/k8sgpt/commit/2992c4e5c8abad50c90ed85523c732f19ab1f31c))
* add initial renovate config ([e37dbc7](https://github.com/k8sgpt-ai/k8sgpt/commit/e37dbc7909f1c520c4c6660c25b45de5847ea581))
* add pull request template ([a6d5132](https://github.com/k8sgpt-ai/k8sgpt/commit/a6d5132b8c2ff077680e2edfd8361a93008197fd))
* add release-please ([da7b409](https://github.com/k8sgpt-ai/k8sgpt/commit/da7b40978d55a6afed4c3a1ca83a756238feaca8))
* change module repo ([a307c13](https://github.com/k8sgpt-ai/k8sgpt/commit/a307c132b3464ff2e949c8a5588e01d344de91a0))
* **deps:** pin dependencies ([f6072f5](https://github.com/k8sgpt-ai/k8sgpt/commit/f6072f56cbe2c073b7b7ebef6c12fa98120e54e2))
* **deps:** pin dependencies ([5b360de](https://github.com/k8sgpt-ai/k8sgpt/commit/5b360de2ae6094cf850a4ae973a22855c21a9040))
* **deps:** pin dependencies ([7fea7d1](https://github.com/k8sgpt-ai/k8sgpt/commit/7fea7d14a572fe0fd05f5f241b98e93655fb1965))
* **deps:** update actions/checkout digest to 8f4b7f8 ([9955d75](https://github.com/k8sgpt-ai/k8sgpt/commit/9955d754505b60f28d17397132a1d02e95ffe303))
* **main:** release 0.0.3 ([53c9947](https://github.com/k8sgpt-ai/k8sgpt/commit/53c994725ea2c2c54898ffe5307d9df40e9c1fe5))
* **main:** release 0.0.3 ([f5d8609](https://github.com/k8sgpt-ai/k8sgpt/commit/f5d86092f49faef8d71cb950986d76c3f92daf46))
* **main:** release 0.0.3 ([22873a6](https://github.com/k8sgpt-ai/k8sgpt/commit/22873a67163e58484d2a0ad343b4ba3c83e51d8f))
* **main:** release 0.0.4 ([aef7256](https://github.com/k8sgpt-ai/k8sgpt/commit/aef7256dc3a85817573744f8b4a54f834368bac7))
* **main:** release 0.0.4 ([6dbcde9](https://github.com/k8sgpt-ai/k8sgpt/commit/6dbcde94e961a6e5a1fc0559d2a1da5567a659de))
* release 0.0.3 ([4840aa0](https://github.com/k8sgpt-ai/k8sgpt/commit/4840aa081e3aa4a7a01fd3fd5f837fa6f0c3c02c))
* release 0.0.3 ([de02795](https://github.com/k8sgpt-ai/k8sgpt/commit/de027955ea18a751c5f991e7ff0f60b90ae704b0))
* release 0.0.3 ([a927c32](https://github.com/k8sgpt-ai/k8sgpt/commit/a927c32def806bb8b99e1cfcd4ee3dcdeca6ae5d))
* release 0.0.4 ([08f2c31](https://github.com/k8sgpt-ai/k8sgpt/commit/08f2c3112e2cc16b49b9cf8fdbd97368acecc754))

## [0.0.4](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.0.4...v0.0.4) (2023-03-24)


### Features

* build container ([260640f](https://github.com/k8sgpt-ai/k8sgpt/commit/260640f865baefba8ac256f800d4992f25ca15fd))
* find replicaset errors ([8ac56e0](https://github.com/k8sgpt-ai/k8sgpt/commit/8ac56e062baef2a0cf7c7ce2b4c97753f079f157))


### Bug Fixes

* add permissions to read repository ([d6cc4cf](https://github.com/k8sgpt-ai/k8sgpt/commit/d6cc4cfcbffbf84f27c7e4e4159da1e42dd5d689))
* build ([1fbed3e](https://github.com/k8sgpt-ai/k8sgpt/commit/1fbed3e44ff790fccfef502ddafae92e34629c21))
* container naming ([115276e](https://github.com/k8sgpt-ai/k8sgpt/commit/115276e01a38fc1692d6b66ab56a33f1e1793974))
* **deps:** update module github.com/sashabaranov/go-openai to v1.5.5 ([105fe44](https://github.com/k8sgpt-ai/k8sgpt/commit/105fe44680e5a987d4a65ff9c58b5b2211808c5e))
* **deps:** update module github.com/sashabaranov/go-openai to v1.5.6 ([37a1d3f](https://github.com/k8sgpt-ai/k8sgpt/commit/37a1d3f47e07caddb168f228627973870a9d867e))
* **deps:** update module github.com/sashabaranov/go-openai to v1.5.7 ([7f7726d](https://github.com/k8sgpt-ai/k8sgpt/commit/7f7726d59a63baeaf8ff110e00b30a20ec7f1df5))
* minor adaptions ([ef17b84](https://github.com/k8sgpt-ai/k8sgpt/commit/ef17b845ba3c65c16ed5dcc417e3e3d3d40dd04e))
* release please config ([c402c7b](https://github.com/k8sgpt-ai/k8sgpt/commit/c402c7bab7baababbbc7c82965d8337de7d50d35))
* remove sboms from goreleaser ([addc01f](https://github.com/k8sgpt-ai/k8sgpt/commit/addc01f700dd2ea31ec24dcf4995bb7ed4a4785e))


### Docs

* add some important information to contributing ([9ab7f58](https://github.com/k8sgpt-ai/k8sgpt/commit/9ab7f587620d69e4e8fc98faabce6417c35f7497))
* update CONTRIBUTING ([05a787d](https://github.com/k8sgpt-ai/k8sgpt/commit/05a787d53dfe5e625c6449ac1e21ec36e66ddd28))
* update CONTRIBUTING ([26449e1](https://github.com/k8sgpt-ai/k8sgpt/commit/26449e10efd8926cccd4a2eaa4e9dc3afa8bd01a))


### Other

* add bot secret to goreleaser ([171e58b](https://github.com/k8sgpt-ai/k8sgpt/commit/171e58b51107f75717694e35c4e249ee41f0409a))
* add brew tap generation on release ([2992c4e](https://github.com/k8sgpt-ai/k8sgpt/commit/2992c4e5c8abad50c90ed85523c732f19ab1f31c))
* add initial renovate config ([e37dbc7](https://github.com/k8sgpt-ai/k8sgpt/commit/e37dbc7909f1c520c4c6660c25b45de5847ea581))
* add pull request template ([a6d5132](https://github.com/k8sgpt-ai/k8sgpt/commit/a6d5132b8c2ff077680e2edfd8361a93008197fd))
* add release-please ([da7b409](https://github.com/k8sgpt-ai/k8sgpt/commit/da7b40978d55a6afed4c3a1ca83a756238feaca8))
* change module repo ([a307c13](https://github.com/k8sgpt-ai/k8sgpt/commit/a307c132b3464ff2e949c8a5588e01d344de91a0))
* **deps:** pin dependencies ([f6072f5](https://github.com/k8sgpt-ai/k8sgpt/commit/f6072f56cbe2c073b7b7ebef6c12fa98120e54e2))
* **deps:** pin dependencies ([5b360de](https://github.com/k8sgpt-ai/k8sgpt/commit/5b360de2ae6094cf850a4ae973a22855c21a9040))
* **deps:** pin dependencies ([7fea7d1](https://github.com/k8sgpt-ai/k8sgpt/commit/7fea7d14a572fe0fd05f5f241b98e93655fb1965))
* **deps:** update actions/checkout digest to 8f4b7f8 ([9955d75](https://github.com/k8sgpt-ai/k8sgpt/commit/9955d754505b60f28d17397132a1d02e95ffe303))
* **main:** release 0.0.3 ([53c9947](https://github.com/k8sgpt-ai/k8sgpt/commit/53c994725ea2c2c54898ffe5307d9df40e9c1fe5))
* **main:** release 0.0.3 ([f5d8609](https://github.com/k8sgpt-ai/k8sgpt/commit/f5d86092f49faef8d71cb950986d76c3f92daf46))
* **main:** release 0.0.3 ([22873a6](https://github.com/k8sgpt-ai/k8sgpt/commit/22873a67163e58484d2a0ad343b4ba3c83e51d8f))
* **main:** release 0.0.4 ([6dbcde9](https://github.com/k8sgpt-ai/k8sgpt/commit/6dbcde94e961a6e5a1fc0559d2a1da5567a659de))
* release 0.0.3 ([4840aa0](https://github.com/k8sgpt-ai/k8sgpt/commit/4840aa081e3aa4a7a01fd3fd5f837fa6f0c3c02c))
* release 0.0.3 ([de02795](https://github.com/k8sgpt-ai/k8sgpt/commit/de027955ea18a751c5f991e7ff0f60b90ae704b0))
* release 0.0.3 ([a927c32](https://github.com/k8sgpt-ai/k8sgpt/commit/a927c32def806bb8b99e1cfcd4ee3dcdeca6ae5d))
* release 0.0.4 ([08f2c31](https://github.com/k8sgpt-ai/k8sgpt/commit/08f2c3112e2cc16b49b9cf8fdbd97368acecc754))

## [0.0.4](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.0.3...v0.0.4) (2023-03-24)


### Bug Fixes

* **deps:** update module github.com/sashabaranov/go-openai to v1.5.7 ([7f7726d](https://github.com/k8sgpt-ai/k8sgpt/commit/7f7726d59a63baeaf8ff110e00b30a20ec7f1df5))


### Docs

* add some important information to contributing ([9ab7f58](https://github.com/k8sgpt-ai/k8sgpt/commit/9ab7f587620d69e4e8fc98faabce6417c35f7497))
* update CONTRIBUTING ([05a787d](https://github.com/k8sgpt-ai/k8sgpt/commit/05a787d53dfe5e625c6449ac1e21ec36e66ddd28))
* update CONTRIBUTING ([26449e1](https://github.com/k8sgpt-ai/k8sgpt/commit/26449e10efd8926cccd4a2eaa4e9dc3afa8bd01a))


### Other

* add bot secret to goreleaser ([171e58b](https://github.com/k8sgpt-ai/k8sgpt/commit/171e58b51107f75717694e35c4e249ee41f0409a))
* add brew tap generation on release ([2992c4e](https://github.com/k8sgpt-ai/k8sgpt/commit/2992c4e5c8abad50c90ed85523c732f19ab1f31c))
* **deps:** update actions/checkout digest to 8f4b7f8 ([9955d75](https://github.com/k8sgpt-ai/k8sgpt/commit/9955d754505b60f28d17397132a1d02e95ffe303))

## [0.0.3](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.0.3...v0.0.3) (2023-03-23)


### Features

* build container ([260640f](https://github.com/k8sgpt-ai/k8sgpt/commit/260640f865baefba8ac256f800d4992f25ca15fd))
* find replicaset errors ([8ac56e0](https://github.com/k8sgpt-ai/k8sgpt/commit/8ac56e062baef2a0cf7c7ce2b4c97753f079f157))


### Bug Fixes

* add permissions to read repository ([d6cc4cf](https://github.com/k8sgpt-ai/k8sgpt/commit/d6cc4cfcbffbf84f27c7e4e4159da1e42dd5d689))
* build ([1fbed3e](https://github.com/k8sgpt-ai/k8sgpt/commit/1fbed3e44ff790fccfef502ddafae92e34629c21))
* container naming ([115276e](https://github.com/k8sgpt-ai/k8sgpt/commit/115276e01a38fc1692d6b66ab56a33f1e1793974))
* **deps:** update module github.com/sashabaranov/go-openai to v1.5.5 ([105fe44](https://github.com/k8sgpt-ai/k8sgpt/commit/105fe44680e5a987d4a65ff9c58b5b2211808c5e))
* **deps:** update module github.com/sashabaranov/go-openai to v1.5.6 ([37a1d3f](https://github.com/k8sgpt-ai/k8sgpt/commit/37a1d3f47e07caddb168f228627973870a9d867e))
* minor adaptions ([ef17b84](https://github.com/k8sgpt-ai/k8sgpt/commit/ef17b845ba3c65c16ed5dcc417e3e3d3d40dd04e))
* release please config ([c402c7b](https://github.com/k8sgpt-ai/k8sgpt/commit/c402c7bab7baababbbc7c82965d8337de7d50d35))


### Other

* add initial renovate config ([e37dbc7](https://github.com/k8sgpt-ai/k8sgpt/commit/e37dbc7909f1c520c4c6660c25b45de5847ea581))
* add pull request template ([a6d5132](https://github.com/k8sgpt-ai/k8sgpt/commit/a6d5132b8c2ff077680e2edfd8361a93008197fd))
* add release-please ([da7b409](https://github.com/k8sgpt-ai/k8sgpt/commit/da7b40978d55a6afed4c3a1ca83a756238feaca8))
* change module repo ([a307c13](https://github.com/k8sgpt-ai/k8sgpt/commit/a307c132b3464ff2e949c8a5588e01d344de91a0))
* **deps:** pin dependencies ([5b360de](https://github.com/k8sgpt-ai/k8sgpt/commit/5b360de2ae6094cf850a4ae973a22855c21a9040))
* **deps:** pin dependencies ([7fea7d1](https://github.com/k8sgpt-ai/k8sgpt/commit/7fea7d14a572fe0fd05f5f241b98e93655fb1965))
* **main:** release 0.0.3 ([f5d8609](https://github.com/k8sgpt-ai/k8sgpt/commit/f5d86092f49faef8d71cb950986d76c3f92daf46))
* **main:** release 0.0.3 ([22873a6](https://github.com/k8sgpt-ai/k8sgpt/commit/22873a67163e58484d2a0ad343b4ba3c83e51d8f))
* release 0.0.3 ([4840aa0](https://github.com/k8sgpt-ai/k8sgpt/commit/4840aa081e3aa4a7a01fd3fd5f837fa6f0c3c02c))
* release 0.0.3 ([de02795](https://github.com/k8sgpt-ai/k8sgpt/commit/de027955ea18a751c5f991e7ff0f60b90ae704b0))
* release 0.0.3 ([a927c32](https://github.com/k8sgpt-ai/k8sgpt/commit/a927c32def806bb8b99e1cfcd4ee3dcdeca6ae5d))

## [0.0.3](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.0.3...v0.0.3) (2023-03-23)


### Other

* release 0.0.3 ([de02795](https://github.com/k8sgpt-ai/k8sgpt/commit/de027955ea18a751c5f991e7ff0f60b90ae704b0))
* release 0.0.3 ([a927c32](https://github.com/k8sgpt-ai/k8sgpt/commit/a927c32def806bb8b99e1cfcd4ee3dcdeca6ae5d))

## [0.0.3](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.0.2...v0.0.3) (2023-03-23)


### Features

* build container ([260640f](https://github.com/k8sgpt-ai/k8sgpt/commit/260640f865baefba8ac256f800d4992f25ca15fd))
* find replicaset errors ([8ac56e0](https://github.com/k8sgpt-ai/k8sgpt/commit/8ac56e062baef2a0cf7c7ce2b4c97753f079f157))


### Bug Fixes

* add permissions to read repository ([d6cc4cf](https://github.com/k8sgpt-ai/k8sgpt/commit/d6cc4cfcbffbf84f27c7e4e4159da1e42dd5d689))
* build ([1fbed3e](https://github.com/k8sgpt-ai/k8sgpt/commit/1fbed3e44ff790fccfef502ddafae92e34629c21))
* container naming ([115276e](https://github.com/k8sgpt-ai/k8sgpt/commit/115276e01a38fc1692d6b66ab56a33f1e1793974))
* **deps:** update module github.com/sashabaranov/go-openai to v1.5.6 ([37a1d3f](https://github.com/k8sgpt-ai/k8sgpt/commit/37a1d3f47e07caddb168f228627973870a9d867e))
* minor adaptions ([ef17b84](https://github.com/k8sgpt-ai/k8sgpt/commit/ef17b845ba3c65c16ed5dcc417e3e3d3d40dd04e))
* release please config ([c402c7b](https://github.com/k8sgpt-ai/k8sgpt/commit/c402c7bab7baababbbc7c82965d8337de7d50d35))


### Other

* add release-please ([da7b409](https://github.com/k8sgpt-ai/k8sgpt/commit/da7b40978d55a6afed4c3a1ca83a756238feaca8))
* **deps:** pin dependencies ([5b360de](https://github.com/k8sgpt-ai/k8sgpt/commit/5b360de2ae6094cf850a4ae973a22855c21a9040))
