# QianWen Skill categories

Try to classify each discovery request into one primary category before generating keywords. Open only the selected category's detail file. If no category clearly matches, leave the request unclassified and generate keywords directly from the user's task. The detail files contain search vocabulary, not official secondary categories.

| Category key | 中文 | Scope | Keyword file |
| --- | --- | --- | --- |
| `intelligence` | 智能 | AI、模型、智能应用与智能化能力 | [category-intelligence.md](category-intelligence.md) |
| `compute` | 计算 | 计算资源、实例、容器与算力 | [category-compute.md](category-compute.md) |
| `network` | 网络 | 网络、连接、域名、流量与网络诊断 | [category-network.md](category-network.md) |
| `storage` | 存储 | 文件、对象、块存储、备份与归档 | [category-storage.md](category-storage.md) |
| `data` | 数据 | 数据库、数据处理、集成与治理 | [category-data.md](category-data.md) |
| `analysis` | 分析 | DataWorks、MaxCompute、元数据与数据开发 | [category-analysis.md](category-analysis.md) |
| `security` | 安全 | 安全检测、身份权限、合规与风险 | [category-security.md](category-security.md) |
| `operations` | 运维 | 跨产品资源、云治理、CLI 与平台认证 | [category-operations.md](category-operations.md) |
| `account` | 账户 | 账号、认证、配额、账单与用量 | [category-account.md](category-account.md) |
| `meta-skill` | 元技能 | Skill、MCP、插件与 Agent 扩展管理 | [category-meta-skill.md](category-meta-skill.md) |

## Classification rules

- Use the desired outcome, not incidental nouns, to choose the primary category.
- Select a secondary category only when the request contains two independent outcomes; search each category separately.
- Use only exact keys from the table.
- When no category clearly matches, leave the request unclassified and generate short keywords directly from the user's task.
- Treat this local table as the complete category registry for this Skill; do not query online category statistics.
- Prefer the primary category over applying several broad filters.
- Distinguish `data` from `analysis` using the platform's current assignments: databases and database diagnosis/access are `data`; DataWorks, MaxCompute/MaxFrame, metadata, and data-development workflows are `analysis`.
- Distinguish `account` from `security`: billing, quota, login, and subscriptions are `account`; authorization policy, secrets, audit, and threats are `security`.
- Use `storage` for OSS, PDS, backup, and SLS query capabilities according to this local category mapping.
- Use `intelligence` for AI capabilities and `meta-skill` for discovering or managing Agent extensions.
