# meta-skill · 元技能

Use for discovering, installing, developing, validating, publishing, or managing Agent extensions.

Use these stable discovery terms for `meta-skill` requests; if the prepared CLI searches return no candidate, report that no corresponding Skill was found.

| Theme | Core queries | High-signal anchor queries |
| --- | --- | --- |
| Skill discovery tooling | 技能发现、技能推荐 | skill discovery, recommendation |
| Skill installation | 安装、卸载、更新 | install, uninstall, update |
| Skill development | 创建、开发、模板 | create, development, template |
| Skill quality | 校验、测试、审核 | validate, test, review |
| Publishing | 发布、上传、市场 | publish, upload, marketplace |
| MCP | MCP | server, configure |
| Agent extension management | 插件管理、扩展管理 | plugin management, extension management |

Use the domain category instead when the user wants to perform the domain task itself, such as generating an image (`intelligence`) or diagnosing a network (`network`).

Do not search a bare packaging noun such as `Skill`, `技能`, `plugin`, `插件`, `extension`, `扩展`, `tool`, or `工具`. Ask for the concrete target task when it is missing. Exclude `qianwen-find-skills` from discovery results unless the user explicitly names that exact slug for inspection or installation.
