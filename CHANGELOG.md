# Changelog

## [0.3.0](https://github.com/timrabl/terraform-provider-snipeit/compare/v0.2.0...v0.3.0) (2026-08-23)


### Features

* detect server version and gate version-specific behavior ([#31](https://github.com/timrabl/terraform-provider-snipeit/issues/31)) ([c611ef5](https://github.com/timrabl/terraform-provider-snipeit/commit/c611ef52f672e0f1abf05eb7d783e50aac505427)), closes [#25](https://github.com/timrabl/terraform-provider-snipeit/issues/25)


### Bug Fixes

* tolerate Snipe-IT response shape changes across versions ([#30](https://github.com/timrabl/terraform-provider-snipeit/issues/30)) ([a527454](https://github.com/timrabl/terraform-provider-snipeit/commit/a5274541502dce4d8400444bd07f487a007b540f)), closes [#24](https://github.com/timrabl/terraform-provider-snipeit/issues/24) [#27](https://github.com/timrabl/terraform-provider-snipeit/issues/27) [#29](https://github.com/timrabl/terraform-provider-snipeit/issues/29)

## [0.2.0](https://github.com/timrabl/terraform-provider-snipeit/compare/v0.1.1...v0.2.0) (2026-08-22)


### Features

* support purchase cost attributes ([#21](https://github.com/timrabl/terraform-provider-snipeit/issues/21)) ([b2a1849](https://github.com/timrabl/terraform-provider-snipeit/commit/b2a184906b1d887328429a31a08195008eb51f06)), closes [#14](https://github.com/timrabl/terraform-provider-snipeit/issues/14)


### Bug Fixes

* **client:** flatten non-envelope error bodies ([#19](https://github.com/timrabl/terraform-provider-snipeit/issues/19)) ([23ea56a](https://github.com/timrabl/terraform-provider-snipeit/commit/23ea56af646c889739360c6cb33e94eacc2557cc)), closes [#18](https://github.com/timrabl/terraform-provider-snipeit/issues/18)
* preserve explicitly configured zero values ([#22](https://github.com/timrabl/terraform-provider-snipeit/issues/22)) ([1fa49ba](https://github.com/timrabl/terraform-provider-snipeit/commit/1fa49bad3ef913337927817cbaf489ded64754e6))

## [0.1.1](https://github.com/timrabl/terraform-provider-snipeit/compare/v0.1.0...v0.1.1) (2026-08-22)


### Bug Fixes

* **assets:** detect soft-deleted assets on read ([e957538](https://github.com/timrabl/terraform-provider-snipeit/commit/e957538676fd07dffe2b2cd39b62dcd4c619a874))
* **client:** treat API redirects as not found ([25e8335](https://github.com/timrabl/terraform-provider-snipeit/commit/25e83350e53200e828459b411cc55875ed8b08a1))

## 0.1.0 (2026-08-22)


### Features

* add provider scaffolding and shared helpers ([3aa8ab9](https://github.com/timrabl/terraform-provider-snipeit/commit/3aa8ab96a2b61646883d4196d97b17a3784f2287))
* **assets:** manufacturers, categories, status labels, models, hardware ([1fd1527](https://github.com/timrabl/terraform-provider-snipeit/commit/1fd1527bc6aee60df01d775f8a61a380736d71ca))
* **client:** add Snipe-IT API transport ([37885cf](https://github.com/timrabl/terraform-provider-snipeit/commit/37885cfc43efc6ca05b92fdf3294388f9e80c810))
* **customfields:** fieldsets, fields and associations ([2edbfb3](https://github.com/timrabl/terraform-provider-snipeit/commit/2edbfb380fa2e99ac58641fb5e54624c96056fd7))
* **inventory:** accessories, consumables, components and checkouts ([af6c041](https://github.com/timrabl/terraform-provider-snipeit/commit/af6c041f0f9684a803c3a6302aae8761d693dcb9))
* **licensing:** licenses and seat assignments ([017c4a6](https://github.com/timrabl/terraform-provider-snipeit/commit/017c4a61c05f673cc59d1ec1d7a37c4c798cd421))
* **operations:** maintenances, activity reports and audit lists ([352ca6e](https://github.com/timrabl/terraform-provider-snipeit/commit/352ca6e0f1b0a0624e975467d83e80f847c51199))
* **organization:** companies, departments, locations, suppliers ([56f3d28](https://github.com/timrabl/terraform-provider-snipeit/commit/56f3d280abb165bbfc004b46b86eb38695b6e0ae))
* **people:** users and groups ([cbbe9b9](https://github.com/timrabl/terraform-provider-snipeit/commit/cbbe9b96d54c801c7405db0b2b5f497d72d5b881))


### Bug Fixes

* **deps:** update kin-openapi past authentication-bypass advisory ([7597285](https://github.com/timrabl/terraform-provider-snipeit/commit/7597285ac301004f10b6f0de34966b227b64a9ed))
* **deps:** update modules with known vulnerabilities ([7e22eea](https://github.com/timrabl/terraform-provider-snipeit/commit/7e22eeac22ed391029570620ea96e5f4f65f57f3))
