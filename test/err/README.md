# 诊断定位样例

本目录用于对照 VS Code 的 Problems 面板与 `ysc -c` 的诊断内容、级别和行列位置。每个文件只覆盖一个主要行为；请直接在 VS Code 中打开文件查看波浪线，再从仓库根目录运行对应检查命令。

| 文件 | 预期结果 |
| --- | --- |
| `00_clean.ys` | 无错误、无警告 |
| `01_syntax_error.ys` | 第 4 行语法错误 |
| `02_unused_variable.ys` | 第 4 行 `unused_value` 未使用警告 |
| `03_top_level_statement.ys` | 第 3 行顶层语句错误 |
| `04_catch_variable_use.ys` | 无错误、无警告；第 6 行的 `psm` 在 catch 表达式调用参数中被使用，不应报未使用 |
| `05_missing_import.ys` | `http` 未导入，准确位置应为第 4 行 `http.Get` |

在 `yscript` 仓库目录执行单个文件：

```sh
./ysc -c test/err/01_syntax_error.ys
```

验证插件使用的 stdin 检查路径：

```sh
./ysc -c - < test/err/02_unused_variable.ys
```

对照错误/警告时，检查 Problems 面板的文件、行、列、级别和消息是否与 `ysc -c` 输出一致。`ysc -c` 对发现诊断的源码仍可能以成功状态结束，因此应核对输出，不要只看进程退出码。

已知定位差异：当前 `ysc -c test/err/05_missing_import.ys` 的诊断消息指出第 4 行未导入 `http`，但外层位置会落在第 3 行第 6 列（`func main`）。此样例刻意保留该情况，用于确认 VS Code 与 CLI 表现一致并追踪编译器定位修复；准确定位应落在第 4 行的 `http` 标识符。