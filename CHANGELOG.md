# Changelog

## [0.2.7](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.2.6...v0.2.7) (2023-04-25)


### Bug Fixes

* remove pointer to loop variable when searching the latest event to analyze ([#328](https://github.com/k8sgpt-ai/k8sgpt/issues/328)) ([2616220](https://github.com/k8sgpt-ai/k8sgpt/commit/2616220935d450030c8a9f2f2741c3607aa4b663))

## [0.2.6](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.2.5...v0.2.6) (2023-04-25)


### Bug Fixes

* explicitly pass in filter to async analysis go routine ([#326](https://github.com/k8sgpt-ai/k8sgpt/issues/326)) ([692cd06](https://github.com/k8sgpt-ai/k8sgpt/commit/692cd06c385c1c6f458994f6e975a9fce2bc1c57))

## [0.2.5](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.2.4...v0.2.5) (2023-04-25)


### Features

* add configuration interface to support customer providers ([84a3cc0](https://github.com/k8sgpt-ai/k8sgpt/commit/84a3cc05fb6e21b732ef777351b42db8045e1093))
* add k8sgpt grafana dashboard ([#316](https://github.com/k8sgpt-ai/k8sgpt/issues/316)) ([ff79982](https://github.com/k8sgpt-ai/k8sgpt/commit/ff799825cfe5856bb97c8f38d939ec36b19fa30a))
* add serve & integration to README ([a65ee7f](https://github.com/k8sgpt-ai/k8sgpt/commit/a65ee7fc0957c7ba9369bdbe12e648818ca3f841))
* add subproject group to CODEOWNERS ([#322](https://github.com/k8sgpt-ai/k8sgpt/issues/322)) ([2391603](https://github.com/k8sgpt-ai/k8sgpt/commit/2391603075e73b91d9988d40eecddfc3e0593405))
* allow to set a baseurl ([#310](https://github.com/k8sgpt-ai/k8sgpt/issues/310)) ([cf797a6](https://github.com/k8sgpt-ai/k8sgpt/commit/cf797a6eb67efba957704077b4b04ed3ee166c24))
* async calls ([#311](https://github.com/k8sgpt-ai/k8sgpt/issues/311)) ([c3cc413](https://github.com/k8sgpt-ai/k8sgpt/commit/c3cc413e7fc3b06b310779dfa3cb4863ea9f3ed2))
* modify error handling to return a list of errors to display to the user at the end of analysis without blocking it if an error is detected (e.g., version of an object is not available on user's cluster) ([fa087ff](https://github.com/k8sgpt-ai/k8sgpt/commit/fa087ff5593871d2a07d68f203dd91e66c57e40b))
* the overall optimization and architecture design of the makefile are made ([#317](https://github.com/k8sgpt-ai/k8sgpt/issues/317)) ([754bf91](https://github.com/k8sgpt-ai/k8sgpt/commit/754bf917e1ac524699d38fb2dc59bc5d858f6d80))
* update readme ([#314](https://github.com/k8sgpt-ai/k8sgpt/issues/314)) ([ddd830c](https://github.com/k8sgpt-ai/k8sgpt/commit/ddd830cc569278c157480c44a671c9be20c95b24))
* use OS conform path for storing cached results ([7eddb8f](https://github.com/k8sgpt-ai/k8sgpt/commit/7eddb8f4a6dc61d5f66fc1bf56c0e8cbf9370229)), closes [#323](https://github.com/k8sgpt-ai/k8sgpt/issues/323)


### Bug Fixes

* **deps:** update module github.com/aquasecurity/trivy-operator to v0.13.1 ([#321](https://github.com/k8sgpt-ai/k8sgpt/issues/321)) ([e7f74db](https://github.com/k8sgpt-ai/k8sgpt/commit/e7f74db6e556146b898437bb777c2b803d1bec4f))
* **deps:** update module github.com/prometheus/client_golang to v1.15.0 ([#303](https://github.com/k8sgpt-ai/k8sgpt/issues/303)) ([df2ed41](https://github.com/k8sgpt-ai/k8sgpt/commit/df2ed4185b5a33a18e6b144c85bec3902c14d209))
* **deps:** update module github.com/sashabaranov/go-openai to v1.9.0 ([#298](https://github.com/k8sgpt-ai/k8sgpt/issues/298)) ([0472c36](https://github.com/k8sgpt-ai/k8sgpt/commit/0472c363a4d8a90556bc744fbf513ad63281e38b))


### Other

* add serviceMonitor in sample yaml ([#304](https://github.com/k8sgpt-ai/k8sgpt/issues/304)) ([0a4ed0d](https://github.com/k8sgpt-ai/k8sgpt/commit/0a4ed0d907c22a924dd79e8945eb9d6d10cd9ce7))
* analyze Pod ReadinessProbe faliure ([3c7e0bb](https://github.com/k8sgpt-ai/k8sgpt/commit/3c7e0bba1d4cc8247d248756dcfef884bc406992))
* change license to Apache-2 ([#313](https://github.com/k8sgpt-ai/k8sgpt/issues/313)) ([d0f7a11](https://github.com/k8sgpt-ai/k8sgpt/commit/d0f7a1105fe7ed317785782d3af45c83766b7d80))

## [0.2.4](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.2.3...v0.2.4) (2023-04-18)


### Features

* improve HPA analyzer to check ScaleTargetRef resources  ([#283](https://github.com/k8sgpt-ai/k8sgpt/issues/283)) ([7173203](https://github.com/k8sgpt-ai/k8sgpt/commit/71732037fa40071cef0c2bc143736019d75eac86))
* init logging middleware on server mode ([6742410](https://github.com/k8sgpt-ai/k8sgpt/commit/6742410025d5e99c60045bb314730799f0e1e5ce))


### Bug Fixes

* deployment/cronjob namespace filtering ([#290](https://github.com/k8sgpt-ai/k8sgpt/issues/290)) ([3d684a2](https://github.com/k8sgpt-ai/k8sgpt/commit/3d684a2af7a9e1821bdb8b1bd6e85867b800d3ee))
* ensure parent directories are created in EnsureDirExists function ([#293](https://github.com/k8sgpt-ai/k8sgpt/issues/293)) ([af8b350](https://github.com/k8sgpt-ai/k8sgpt/commit/af8b350520d1a187a199482dd338db0086118db8))
* resolve language toggle bug (issue [#294](https://github.com/k8sgpt-ai/k8sgpt/issues/294)) ([0313627](https://github.com/k8sgpt-ai/k8sgpt/commit/03136278486ba12e3352580b317b9e63fa3a80f0))

## [0.2.3](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.2.2...v0.2.3) (2023-04-16)


### Features

* add node analyzer ([#272](https://github.com/k8sgpt-ai/k8sgpt/issues/272)) ([6247a1c](https://github.com/k8sgpt-ai/k8sgpt/commit/6247a1c0f3c2ead6a59661afed06973c29e57eca))
* add output query param on serve mode & refactor output logic ([9642202](https://github.com/k8sgpt-ai/k8sgpt/commit/9642202ed1b09c06a687651b7818c2a4df8a0c06))
* add server metrics ([#273](https://github.com/k8sgpt-ai/k8sgpt/issues/273)) ([a3becc9](https://github.com/k8sgpt-ai/k8sgpt/commit/a3becc9906515d0567808fee9a4e322451d6dc3f))
* envs to initialise server ([0071e25](https://github.com/k8sgpt-ai/k8sgpt/commit/0071e25992fc86c3882c2066873a2b04b43fe476))
* rename server/main.go to server/server.go ([9121a98](https://github.com/k8sgpt-ai/k8sgpt/commit/9121a983e52fa15c07bcc3bb361df97b8085c24c))
* running in cluster ([842f08c](https://github.com/k8sgpt-ai/k8sgpt/commit/842f08c655fde66b6b628192490e50be2ac3dcef))
* running in cluster ([3988eb2](https://github.com/k8sgpt-ai/k8sgpt/commit/3988eb2fd0a7d29ffa7b7bbc59960ca91e50466e))
* switch config file to XDG conform location ([dee4355](https://github.com/k8sgpt-ai/k8sgpt/commit/dee435514d7f717e4eb63b15a9d9fdb0722330ac))
* wip blocked until we have envs ([fe2c08c](https://github.com/k8sgpt-ai/k8sgpt/commit/fe2c08cf72a6ca271d1b431be66653f1396f304d))


### Bug Fixes

* add new line after version cmd output ([92e7b3d](https://github.com/k8sgpt-ai/k8sgpt/commit/92e7b3d3fb00c33ac48230caac34f45729e2f6b2))
* **deps:** update module github.com/sashabaranov/go-openai to v1.8.0 ([#277](https://github.com/k8sgpt-ai/k8sgpt/issues/277)) ([51b1b35](https://github.com/k8sgpt-ai/k8sgpt/commit/51b1b352acd24ebdc4cf9d9121f25c90e8f76ba7))
* resolve issue with duplicated integration filters. ([960ba56](https://github.com/k8sgpt-ai/k8sgpt/commit/960ba568d0dcc2ace722dc5c9b7c846366a98070))
* use the aiProvider object when launching the server instead of the deprecated configuration keys ([e7076ed](https://github.com/k8sgpt-ai/k8sgpt/commit/e7076ed6093aa9609d8c884b7a03e295057aaa8e))


### Other

* updated ([f0a0c9a](https://github.com/k8sgpt-ai/k8sgpt/commit/f0a0c9aebf627d65b0192ba3d0786cefd81e1fef))

## [0.2.2](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.2.1...v0.2.2) (2023-04-14)


### Features

* add simple health endpoint ([26c0cb2](https://github.com/k8sgpt-ai/k8sgpt/commit/26c0cb2eed75695220007e6d6f7b492c2641a149))
* anoymization based on pr feedback ([19e1b94](https://github.com/k8sgpt-ai/k8sgpt/commit/19e1b94e7c9ce4092f1dabd659023a193b2c4a92))
* anoymization based on pr feedback ([fe52951](https://github.com/k8sgpt-ai/k8sgpt/commit/fe529510b68ac5fbd39c147c7719abe2e7d20894))
* check for auth only in case of --explain ([57790e5](https://github.com/k8sgpt-ai/k8sgpt/commit/57790e5bc7037f57a4f73248fe05cac192511470))
* first version of serve ([b2e8add](https://github.com/k8sgpt-ai/k8sgpt/commit/b2e8adda333fbd508f0f01f2afcabc57bf9948c2))
* unified cmd and api ([9157d4d](https://github.com/k8sgpt-ai/k8sgpt/commit/9157d4dd1312bf75b336beb0e097422b303d22f1))
* updated api ([adae2ef](https://github.com/k8sgpt-ai/k8sgpt/commit/adae2ef71d81431711c552159362336e496b21ee))


### Bug Fixes

* bool conversion ([336ec2a](https://github.com/k8sgpt-ai/k8sgpt/commit/336ec2a42693d0df325b95cbebd9545b19e27725))
* **deps:** update module helm.sh/helm/v3 to v3.11.3 ([4dd91ed](https://github.com/k8sgpt-ai/k8sgpt/commit/4dd91ed8263292476054bc70d3d6a3149f88f1b3))
* naming ([159b385](https://github.com/k8sgpt-ai/k8sgpt/commit/159b3851ec54e93a447b0f13aa4ceb7b8b8f62db))
* start message ([6b63027](https://github.com/k8sgpt-ai/k8sgpt/commit/6b630275eb64b799c50e3074cb22a3b41bb893de))


### Docs

* fix Slack link ([1dccaea](https://github.com/k8sgpt-ai/k8sgpt/commit/1dccaea3f4f96b2da52999eed5031f02a89c0b6e))


### Other

* added oidc ([bffad41](https://github.com/k8sgpt-ai/k8sgpt/commit/bffad41134d231b16f136a619174ff3bee61765a))
* additional analyzers ([23071fd](https://github.com/k8sgpt-ai/k8sgpt/commit/23071fd2e6b421f0f5fcd6e7e4985c6900e5405c))
* **deps:** bump github.com/docker/docker ([#268](https://github.com/k8sgpt-ai/k8sgpt/issues/268)) ([7d1e2ac](https://github.com/k8sgpt-ai/k8sgpt/commit/7d1e2acaf3eaf00929ff43b9373df6a4be100795))
* **deps:** update actions/checkout digest to 83b7061 ([cbe6f27](https://github.com/k8sgpt-ai/k8sgpt/commit/cbe6f27c05e82f55f41b648b01972ba2c43f1534))
* **deps:** update actions/checkout digest to 8e5e7e5 ([#266](https://github.com/k8sgpt-ai/k8sgpt/issues/266)) ([0af34a1](https://github.com/k8sgpt-ai/k8sgpt/commit/0af34a1a95502dc26d7e08bac896f691e4969090))
* **deps:** update module oras.land/oras-go to v1.2.3 ([#249](https://github.com/k8sgpt-ai/k8sgpt/issues/249)) ([13c9231](https://github.com/k8sgpt-ai/k8sgpt/commit/13c9231aafef3a259fd678a80063ad2e968d6e95))
* fixing up tests ([f9b25d9](https://github.com/k8sgpt-ai/k8sgpt/commit/f9b25d9e85a8faaf1aae59d7bedc4c0f3538181e))
* fixing up tests ([498d454](https://github.com/k8sgpt-ai/k8sgpt/commit/498d454c174c7d39da1ca63b2a201e797d7e5e1c))
* Merge branch 'main' into feat/additional-analyzers ([4d36248](https://github.com/k8sgpt-ai/k8sgpt/commit/4d3624830ff840f9ccf11d7da20953bdf4c7c7fc))
* removing field ([ddb51c7](https://github.com/k8sgpt-ai/k8sgpt/commit/ddb51c7af470044a8514ed013b44cc135e4c0f10))

## [0.2.1](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.2.0...v0.2.1) (2023-04-12)


### Features

* add anonymization example to README ([8a60b57](https://github.com/k8sgpt-ai/k8sgpt/commit/8a60b579409c67f092156ba1adf1be22cce37b8c))
* add anonymization flag ([d2a84ea](https://github.com/k8sgpt-ai/k8sgpt/commit/d2a84ea2b5c800dd900aac3a48b1914bd9ddb917))
* add more details on anonymize flag ([b687473](https://github.com/k8sgpt-ai/k8sgpt/commit/b687473e6169406002b0ee8be6ebb9ce43b46495))
* add storage class names' check. ([c8ba7d6](https://github.com/k8sgpt-ai/k8sgpt/commit/c8ba7d62d2f1d262263d1dff8f980e91cdcd50e8))
* improve documentation ([6f08654](https://github.com/k8sgpt-ai/k8sgpt/commit/6f0865413fc2854450d217225199cec199972490))
* improve documentation & update hpa message ([11326c1](https://github.com/k8sgpt-ai/k8sgpt/commit/11326c1c5f307c718e8d1e56099537314ffedadd))
* improve security of the MaskString function ([08f2a89](https://github.com/k8sgpt-ai/k8sgpt/commit/08f2a89e54a65544322814286977b2c05acce89d))
* initial impl of integration ([b0e5170](https://github.com/k8sgpt-ai/k8sgpt/commit/b0e517006e65ac2b4e2d4e2696531d4bbf62c34b))
* initial impl of integration ([61d6e52](https://github.com/k8sgpt-ai/k8sgpt/commit/61d6e524657272cf3a967c724f212677fcfe7d2b))
* integration ready for first review ([3682f5c](https://github.com/k8sgpt-ai/k8sgpt/commit/3682f5c7ebb9590e92162eed214a8127f71bcd81))
* introduce StatefulSet analyser. ([c041ce2](https://github.com/k8sgpt-ai/k8sgpt/commit/c041ce2bbb4ecbc6f5637207c9f3071eee022744))
* refactor integration to use Failure object ([c0afc0f](https://github.com/k8sgpt-ai/k8sgpt/commit/c0afc0f5c91cfa50b1f7af901800ff0a2b492d18))
* return errors if filter specified by flag does not exist. ([dd5824f](https://github.com/k8sgpt-ai/k8sgpt/commit/dd5824f4365b01e3c501d8b5cda914dff138e03d))


### Bug Fixes

* **deps:** update kubernetes packages to v0.27.0 ([7a97034](https://github.com/k8sgpt-ai/k8sgpt/commit/7a97034cf41cb265111c752ee3d54fd90524ef59))
* **deps:** update module github.com/sashabaranov/go-openai to v1.7.0 ([#227](https://github.com/k8sgpt-ai/k8sgpt/issues/227)) ([5f3a5a5](https://github.com/k8sgpt-ai/k8sgpt/commit/5f3a5a54a02967acce40f8b4e9dd3a154c83f58c))
* exit progressbar on error ([#99](https://github.com/k8sgpt-ai/k8sgpt/issues/99)) ([fe261b3](https://github.com/k8sgpt-ai/k8sgpt/commit/fe261b375f4d7990906620f53ac26e792a34731b))
* exit progressbar on error ([#99](https://github.com/k8sgpt-ai/k8sgpt/issues/99)) ([ab55f15](https://github.com/k8sgpt-ai/k8sgpt/commit/ab55f157ef026502d29eadf5ad83e917fe085a6c))
* improve ReplaceIfMatch regex ([fd936ce](https://github.com/k8sgpt-ai/k8sgpt/commit/fd936ceaf725d1c1ed1f53eaa2204455dcd1e2af))
* pdb test ([705d2a0](https://github.com/k8sgpt-ai/k8sgpt/commit/705d2a0dcebb63783782e06b6b775393daf1efb7))
* use hpa namespace instead analyzer namespace ([#230](https://github.com/k8sgpt-ai/k8sgpt/issues/230)) ([a582d44](https://github.com/k8sgpt-ai/k8sgpt/commit/a582d444c5c53f25d7172947c690b35cad2cc176))


### Docs

* add statefulSet analyzer in the docs. ([#233](https://github.com/k8sgpt-ai/k8sgpt/issues/233)) ([b45ff1a](https://github.com/k8sgpt-ai/k8sgpt/commit/b45ff1aa8ef447df2b74bb8c6225e2f3d7c5bd63))
* add statefulSet analyzer in the docs. ([#233](https://github.com/k8sgpt-ai/k8sgpt/issues/233)) ([ba01bd4](https://github.com/k8sgpt-ai/k8sgpt/commit/ba01bd4b6ecd64fbe249be54f20471afc6339208))


### Other

* add fakeai provider ([#218](https://github.com/k8sgpt-ai/k8sgpt/issues/218)) ([e449cb6](https://github.com/k8sgpt-ai/k8sgpt/commit/e449cb60230d440d5b8e00062db63de5d6d413bf))
* adding k8sgpt-approvers ([#238](https://github.com/k8sgpt-ai/k8sgpt/issues/238)) ([db1388f](https://github.com/k8sgpt-ai/k8sgpt/commit/db1388fd20dcf21069adcecd2796f2e1231162c8))
* adding k8sgpt-approvers ([#238](https://github.com/k8sgpt-ai/k8sgpt/issues/238)) ([992b107](https://github.com/k8sgpt-ai/k8sgpt/commit/992b107c2d906663bb22998004a0859bccd45c77))
* compiling successfully ([80ac51c](https://github.com/k8sgpt-ai/k8sgpt/commit/80ac51c804351226e1764e3e649ac56e22de3749))
* **deps:** update anchore/sbom-action action to v0.14.1 ([#228](https://github.com/k8sgpt-ai/k8sgpt/issues/228)) ([9423b53](https://github.com/k8sgpt-ai/k8sgpt/commit/9423b53c1dbae3d0762420a0bacbdace9a2c18c9))
* **deps:** update google-github-actions/release-please-action digest to c078ea3 ([a1d8012](https://github.com/k8sgpt-ai/k8sgpt/commit/a1d8012a5c748aee3f16621d6da9a0f0c8cba293))
* **deps:** update google-github-actions/release-please-action digest to f7edb9e ([#241](https://github.com/k8sgpt-ai/k8sgpt/issues/241)) ([55dda43](https://github.com/k8sgpt-ai/k8sgpt/commit/55dda432ab89c4917bd28fceabcbe5569c0bf530))
* **deps:** update google-github-actions/release-please-action digest to f7edb9e ([#241](https://github.com/k8sgpt-ai/k8sgpt/issues/241)) ([21dc61c](https://github.com/k8sgpt-ai/k8sgpt/commit/21dc61c04f4d772b5147b38a4d28e5dbddf5cdd8))
* fix mistake introduced by ab55f157 ([#240](https://github.com/k8sgpt-ai/k8sgpt/issues/240)) ([428c348](https://github.com/k8sgpt-ai/k8sgpt/commit/428c3485868a7be95ea6776694e30b36badf4b5c))
* fix mistake introduced by ab55f157 ([#240](https://github.com/k8sgpt-ai/k8sgpt/issues/240)) ([3845d47](https://github.com/k8sgpt-ai/k8sgpt/commit/3845d4747f4e0fc823d1bcf631d6ecdd5e4ccd03))
* Fixing broken tests ([c809af3](https://github.com/k8sgpt-ai/k8sgpt/commit/c809af3f47388599fda3a88a4638feae1dc90492))
* fixing filters ([258c69a](https://github.com/k8sgpt-ai/k8sgpt/commit/258c69a17c977867dfd0a7ad02727270b7c172e7))
* fixing filters ([4d20f70](https://github.com/k8sgpt-ai/k8sgpt/commit/4d20f70fb40ff326ceb279f699068ec4956a2f10))
* merged ([096321b](https://github.com/k8sgpt-ai/k8sgpt/commit/096321b31a6cf0d53b1861a3e4ad1efe84f697cc))
* updated analysis_test.go ([825e9a4](https://github.com/k8sgpt-ai/k8sgpt/commit/825e9a43bd3ab7aa3ea52f315993cd778ea039e3))
* updated link output ([1b7f4ce](https://github.com/k8sgpt-ai/k8sgpt/commit/1b7f4ce44a499e5389aec42fdee00bfa81ef0888))
* updating based on feedback ([5e5d4b6](https://github.com/k8sgpt-ai/k8sgpt/commit/5e5d4b6de160dc7533067e1c0d8403c3faac1a9f))
* weird new line after filter removed ([fabe01a](https://github.com/k8sgpt-ai/k8sgpt/commit/fabe01aa019f1db45ed2ff780f0d6d63297b230b))

## [0.2.0](https://github.com/k8sgpt-ai/k8sgpt/compare/v0.1.8...v0.2.0) (2023-04-05)


### ⚠ BREAKING CHANGES

* The format of the configuration file has changed. Users must update their configuration files to use the new format.

### Features

* add support for new configuration format ([9b243cd](https://github.com/k8sgpt-ai/k8sgpt/commit/9b243cdcaab1742fca2516bc1ae5710505e0eb65))
* add tests for hpa analyzer ([5a59abb](https://github.com/k8sgpt-ai/k8sgpt/commit/5a59abb55d9e76f3095b903ea973138b1afdccf2))
* add tests for ingress analyzer && Use t.Fatalf to report a fatal error if RunAnalysis returns an unexpected error ([e27e940](https://github.com/k8sgpt-ai/k8sgpt/commit/e27e9409dcaad78eefd79659e91364617407ae59))


### Bug Fixes

* **deps:** update module github.com/sashabaranov/go-openai to v1.6.1 ([#207](https://github.com/k8sgpt-ai/k8sgpt/issues/207)) ([eeac731](https://github.com/k8sgpt-ai/k8sgpt/commit/eeac731858999f6f462a7b6ccf210af603674b30))
* **deps:** update module github.com/spf13/cobra to v1.7.0 ([5d5e082](https://github.com/k8sgpt-ai/k8sgpt/commit/5d5e082f417905954be33b3a620efef674f2588d))
* **deps:** update module github.com/stretchr/testify to v1.8.2 ([f5e3ca0](https://github.com/k8sgpt-ai/k8sgpt/commit/f5e3ca0bcab9325145a2e1d8624f585ffee8e29f))
* **deps:** update module golang.org/x/term to v0.7.0 ([8ab3573](https://github.com/k8sgpt-ai/k8sgpt/commit/8ab3573e13a3e3ab98e6f0aa76e429117c888f7f))
* details in json output ([2f21002](https://github.com/k8sgpt-ai/k8sgpt/commit/2f2100289953af7820bbb01f2c980cf5492de079))
* fixed hpa tests after rebase ([a24d1f1](https://github.com/k8sgpt-ai/k8sgpt/commit/a24d1f1b304e9448e63c1b7fc283b4cc8bc639aa))
* regression on dynamic filters ([93bcc62](https://github.com/k8sgpt-ai/k8sgpt/commit/93bcc627ba64a9139e65290a8512e0a9b4bf1a69))
* Spelling ([ba4d701](https://github.com/k8sgpt-ai/k8sgpt/commit/ba4d7016814ce97353e98658d5bbcd692007e4a9))


### Docs

* add curl command and release-please annoations ([1849209](https://github.com/k8sgpt-ai/k8sgpt/commit/184920988f7da928cca7fae4a676e4ee5f13cad1))
* add guide to details block ([ddc120e](https://github.com/k8sgpt-ai/k8sgpt/commit/ddc120e7c2657385737d3490def28dbabdd2242d))
* add installation guide via packages ([8e4ce6a](https://github.com/k8sgpt-ai/k8sgpt/commit/8e4ce6a974813258fb9cbeabbcaa3b8f6966a748))
* minor change ([53c1330](https://github.com/k8sgpt-ai/k8sgpt/commit/53c13305383eb454fe45fefa3483cef4821d5d34))
* modify README ([fc47c58](https://github.com/k8sgpt-ai/k8sgpt/commit/fc47c58ae1c2b5511ebbe0ed35714e4ecbb4bb7a))
* modify README ([0f46ceb](https://github.com/k8sgpt-ai/k8sgpt/commit/0f46ceb4456a90e7e05aeff23d25d5775bbf9c2b))


### Other

* added initial tests for json output ([22e3166](https://github.com/k8sgpt-ai/k8sgpt/commit/22e31661bff27b28339898826a34ffdcfcff3583))
* analyzer and ai interfacing ([#200](https://github.com/k8sgpt-ai/k8sgpt/issues/200)) ([0195bfa](https://github.com/k8sgpt-ai/k8sgpt/commit/0195bfab30ab748b3bb7f1b8c8f0e988b99ee54d))
* **deps:** pin anchore/sbom-action action to 448520c ([#203](https://github.com/k8sgpt-ai/k8sgpt/issues/203)) ([9ff3fbc](https://github.com/k8sgpt-ai/k8sgpt/commit/9ff3fbc382ed78b55a7a1966ecdae186c03b2848))
* **deps:** update golang docker tag to v1.20.3 ([e9994b8](https://github.com/k8sgpt-ai/k8sgpt/commit/e9994b8d167d4f1d9c0d0dabf8385ff22cfd16a4))
* made json output prettier and improved output ([db40734](https://github.com/k8sgpt-ai/k8sgpt/commit/db40734a0db89850a2a685c9a7f5f5559875b7b3))

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
