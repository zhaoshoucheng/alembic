---
name: sync-progress
description: 完成一个模块后的收尾同步：更新 docs/progress/status.md 对照表与 Roadmap、同步顶层 CLAUDE.md 的 Architecture 段、校验红线。当用户说"收尾"、"同步进展"、"更新 status"或一个模块刚做完时使用。
---

# /sync-progress —— 模块完成后的收尾同步

依据 `docs/progress/status.md` 的"维护约定"，保证文档与代码一致。

## 步骤

1. **确认本次改动范围**：用 `git status` / `git diff` 看清这轮做了什么，
   对应 status.md 对照表里的哪一行（哪个层、哪个模块）。

2. **更新 `docs/progress/status.md`**：
   - 对照表：更新状态（✅ / 🟡 / ❌）、工程位置、说明；若思想来自某参考项目，
     在说明里记"来自 <代号> 的 <模块>"。
   - "一句话现状"：重写为当前最大的缺口。
   - Roadmap：勾掉/调整已完成项；顶部"最近更新"日期改为今天。

3. **同步顶层 `CLAUDE.md` 的 Architecture 段**：与对照表逐行核对，
   新模块补进对应层的描述，"待做"列表里删去已完成项。

4. **核对 `docs/progress/module-map.md`**：仅当本次工作改变了对某层的理解
   （拆分方式、啃的顺序）才更新，否则不动。

5. **红线自检**（`docs/progress/rules.md`）：对本轮全部改动检查——
   - 无参考工程品牌字眼（清单在 `.claude/local-notes.md`；vendored 第三方代码除外）；
   - 无硬编码密钥，凭据只走环境变量。

6. **验证**：跑 `make build`，若本轮动了带测试的包再跑 `make test`，确认通过。

## 产出

最后向用户汇报：状态表改了哪几行、CLAUDE.md 是否同步、红线检查结果、构建/测试结果。
