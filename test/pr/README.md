# RichProgress 外部库测试

测试目标：<https://github.com/xiguayiqiu/RichProgress>

先将 GitHub checkout 放入 YScript 全局项目库，再从本目录登记依赖并运行：

```sh
mkdir -p "$HOME/.ysc/pkg"
git clone https://github.com/xiguayiqiu/RichProgress.git "$HOME/.ysc/pkg/RichProgress"
cd yscript/test/pr
ysc mod get RichProgress/progress
ysc main.ys
```

`main.ys` 只显示一条 40 格等号进度条：每秒增长 1%，约 100 秒后显示 `[========================================] 100%`。