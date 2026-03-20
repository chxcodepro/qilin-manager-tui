# 银河麒麟 TUI 管理面板

一个给银河麒麟 `V10` 用的终端管理面板，偏系统运维和桌面维护。

## 功能

- 合并查看系统信息和网络信息
- 分析磁盘文件占用
- 合并展示 CPU 和内存占用
- 编辑并保存网卡 IPv4 配置
- 切换银河麒麟软件源
- 清理包缓存
- 清理 `.log` 文件
- 安装固定软件清单

## 一行运行

安装并启动：

```bash
curl -fsSL https://gh-proxy.org/https://raw.githubusercontent.com/chxcodepro/qilin-manager-tui/main/scripts/install.sh | bash
```

## 页面说明

- `系统`
- `磁盘`
- `资源`
- `软件`
- `维护`

## 按键

- 全局：`Tab/Shift+Tab` 切页，`r` 刷新，`` ` `` 展开控制台，`?` 显示帮助，`q` 退出
- `系统` 页：`↑/↓` 选网卡，`Enter` 或 `e` 编辑，`Ctrl+S` 保存网卡配置
- `磁盘` 页：`↑/↓` 选子项，`Enter` 进入目录，`Backspace` 返回上一级
- `资源` 页：`↑/↓` 选进程，`x` 终止进程
- `软件` 页：`↑/↓` 选软件，`空格` 勾选，`/` 搜索，`i` 安装，`d` 卸载，`Esc` 返回默认列表
- `维护` 页：`↑/↓` 选择操作，`Enter` 执行

## 固定软件清单

- `wps-office`
- `electronic-wechat`
- `linuxqq`
- `netdisk`
- `kylin-software-center`
