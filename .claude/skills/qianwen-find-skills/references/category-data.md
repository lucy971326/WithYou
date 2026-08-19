# data · 数据

Use for database access, database assistants, SQL optimization, diagnosis, and database knowledge-base operations.

| Theme | Core queries | High-signal anchor queries |
| --- | --- | --- |
| Database access | 数据库、访问、审计 | DMS, database, SQL |
| Natural-language query | 问数、自然语言 | NL2SQL, data Q&A |
| SQL optimization | SQL、优化、慢 SQL | RDS, DAS |
| Database diagnosis | 诊断、连接、异常 | RDS, Lindorm, PolarDB, DAS |
| Cache database | 缓存、Redis | Tair |
| Analytical database | 分析型数据库、知识库 | AnalyticDB |
| Sessions and locks | 锁、会话 | DAS |

Use `analysis` for DataWorks, MaxCompute/MaxFrame, metadata, lineage, and data-development workflows.

A generic database request may legitimately match DMS, Lindorm, DAS, or another database Skill; do not force an `RDS` anchor. Use `RDS` when the user names RDS or requests a capability specifically associated with RDS, such as RDS troubleshooting or SQL optimization for RDS.
