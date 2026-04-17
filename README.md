# GemeForge / 锻核

**GemeForge — GPU Performance Forge**

让 GPU 始终处于锻造状态，拒绝节能降频。

## 简介

GemeForge（锻核）是一款针对 Windows 平台的 GPU 性能保持工具。它通过维持一个隐藏的 OpenGL 3D 负载（Julia Set 分形着色器），欺骗 NVIDIA/AMD 驱动判定为高负载场景，从而强制 GPU 维持 Boost 频率，消除游戏（如 LOL）中因降频导致的卡顿。

## 功能特点

- 🔥 **隐藏 OpenGL 上下文**：后台维持合法 3D 负载，不影响前台操作
- 🧮 **Julia Set 分形着色器**：产生密集浮点运算，确保 GPU 持续高负载
- 📊 **实时硬件遥测**：监控 GPU/CPU 频率、负载、温度等信息
- 🎮 **游戏优化**：强制 GPU 维持 Boost 频率，消除降频卡顿
- 🖥️ **现代化 UI**：基于 Fyne 框架的跨平台图形界面

## 技术原理

1. 隐藏 OpenGL 上下文维持合法 3D 负载
2. Julia Set 分形着色器产生密集浮点运算
3. 欺骗 NVIDIA/AMD 驱动判定为高负载场景
4. 强制 GPU 维持 Boost 频率，消除 LOL 降频卡顿

## 依赖

- Go 1.26+
- OpenGL 3.3+
- GLFW 3.3

## 构建

```bash
go mod tidy
go build -o GemeForge.exe
```

## 运行

直接双击 `GemeForge.exe` 或在终端运行：

```bash
go run main.go
```

## 注意事项

- 本工具仅适用于 Windows 平台
- 需要安装对应的 GPU 驱动
- 如需精确 GPU 温度，建议安装 NVIDIA NVML

## 许可证

MIT License
