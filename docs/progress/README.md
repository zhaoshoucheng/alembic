# 项目进展文档（progress）

这是一个**探索性学习项目**：用 Go 一点一点重写参考工程 `agent-server`（一个生产级
agent monorepo）的核心思路，目的是逐个模块地理解和实现一个智能体框架。

每次打开 Claude，先读这套文档就能知道：项目是什么、参考工程怎么拆模块、现在做到哪了、
下一步可能做什么、有哪些不可违反的规则。

## 文档导航

- **[rules.md](./rules.md)** — 两条硬性红线（禁品牌字眼、禁硬编码密钥）。**每次必读。**
- **[module-map.md](./module-map.md)** — 参考工程 `agent-server` 的 6 层模块地图，啃的顺序。
- **[status.md](./status.md)** — 当前进展对照表 + Roadmap（最常更新的文件）。

## 工作方式

1. 模块很大，先看 `module-map.md` 选定要做的层。
2. 要看参考工程某模块的源码细节时，去 `.claude/local-notes.md`（不入库）里的速查表
   找对应位置，再按需深入。
3. 做完一个模块，更新 `status.md`，并同步顶层 `CLAUDE.md` 的 Architecture 段。
