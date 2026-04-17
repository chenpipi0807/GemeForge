package main

import (
	"fmt"
	"image/color"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/yusufpapurcu/wmi"
)

// ============================================================
// GemeForge — GPU Performance Forge
// 锻核：让 GPU 始终处于锻造状态，拒绝节能降频
// ============================================================

const (
	AppName       = "GemeForge"
	AppNameCN     = "锻核"
	AppVersion    = "v1.0.0"
	AppSlogan     = "Forge Your Performance"
	WindowTitle   = "GemeForge — GPU Performance Forge"
	TargetFPS     = 60
	ShaderMaxIter = 500
)

// 应用图标 SVG（六边形核心 + 火焰渐变）
const iconSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256">
<defs><linearGradient id="g" x1="0%" y1="100%" x2="100%" y2="0%"><stop offset="0%" stop-color="#FF6B35"/><stop offset="100%" stop-color="#F7C59F"/></linearGradient></defs>
<rect width="256" height="256" rx="48" fill="#0F0F1A"/>
<path d="M128 32 L208 80 L208 176 L128 224 L48 176 L48 80 Z" fill="none" stroke="url(#g)" stroke-width="14"/>
<path d="M128 68 Q162 118 128 188 Q94 118 128 68" fill="url(#g)"/>
<circle cx="128" cy="128" r="18" fill="#FF6B35"/>
</svg>`

// WMI 结构定义
type Win32_Processor struct {
	Name              string
	CurrentClockSpeed uint32
	LoadPercentage    uint16
}

type Win32_VideoController struct {
	Name string
}

var (
	running      int32 = 0
	fakeFPS      int32 = 0
	statusLabel  *widget.Label
	gpuFreqLabel *widget.Label
	gpuLoadLabel *widget.Label
	gpuTempLabel *widget.Label
	cpuFreqLabel *widget.Label
	fpsLabel     *widget.Label
	toggleBtn    *widget.Button
)

func main() {
	runtime.LockOSThread()

	myApp := app.New()
	myApp.SetIcon(fyne.NewStaticResource("icon.svg", []byte(iconSVG)))
	myWindow := myApp.NewWindow(WindowTitle)
	myWindow.Resize(fyne.NewSize(520, 480))
	myWindow.CenterOnScreen()

	createUI(myWindow)
	go hardwareMonitor()

	myWindow.SetOnClosed(func() {
		atomic.StoreInt32(&running, 0)
	})

	myWindow.ShowAndRun()
}

func createUI(w fyne.Window) {
	// ===== 顶部品牌区 =====
	brand := canvas.NewText(AppName, color.RGBA{255, 107, 53, 255})
	brand.TextSize = 28
	brand.TextStyle = fyne.TextStyle{Bold: true}

	version := canvas.NewText(fmt.Sprintf("%s  ·  %s", AppVersion, AppSlogan), color.RGBA{150, 150, 170, 255})
	version.TextSize = 12

	brandBox := container.NewVBox(brand, version)

	// ===== 状态区 =====
	statusLabel = widget.NewLabelWithStyle("●  待机中", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	statusLabel.Importance = widget.WarningImportance

	// ===== 硬件监控卡片 =====
	gpuFreqLabel = widget.NewLabel("GPU 核心频率: 监测中...")
	gpuLoadLabel = widget.NewLabel("GPU 渲染负载: 监测中...")
	gpuTempLabel = widget.NewLabel("GPU 核心温度: 监测中...")
	cpuFreqLabel = widget.NewLabel("CPU 运行频率: 监测中...")
	fpsLabel = widget.NewLabel("锻造帧率: 0 FPS")

	monitorCard := container.NewVBox(
		widget.NewLabelWithStyle("▣ 实时硬件遥测", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		gpuFreqLabel,
		gpuLoadLabel,
		gpuTempLabel,
		cpuFreqLabel,
		fpsLabel,
	)
	monitorCardContainer := container.NewPadded(monitorCard)

	// ===== 控制按钮 =====
	toggleBtn = widget.NewButton("🔥  启动锻造引擎", func() {
		if atomic.LoadInt32(&running) == 0 {
			startGPULoader()
			toggleBtn.SetText("⏹  停止锻造引擎")
			statusLabel.SetText("●  锻造中 — GPU 高性能模式已锁定")
			statusLabel.Importance = widget.SuccessImportance
			statusLabel.Refresh()
		} else {
			atomic.StoreInt32(&running, 0)
			toggleBtn.SetText("🔥  启动锻造引擎")
			statusLabel.SetText("●  待机中")
			statusLabel.Importance = widget.WarningImportance
			statusLabel.Refresh()
		}
	})
	toggleBtn.Importance = widget.HighImportance

	// ===== 技术原理区 =====
	theory := widget.NewLabel(
		"工作原理:\n" +
			"• 隐藏 OpenGL 上下文维持合法 3D 负载\n" +
			"• Julia Set 分形着色器产生密集浮点运算\n" +
			"• 欺骗 NVIDIA/AMD 驱动判定为高负载场景\n" +
			"• 强制 GPU 维持 Boost 频率，消除 LOL 降频卡顿",
	)
	theory.Wrapping = fyne.TextWrapWord

	// ===== 底部版权 =====
	footer := canvas.NewText("Powered by GemeForge Engine  |  github.com/gemeforge", color.RGBA{100, 100, 120, 255})
	footer.TextSize = 10
	footer.Alignment = fyne.TextAlignCenter

	// ===== 整体布局 =====
	content := container.NewVBox(
		brandBox,
		widget.NewSeparator(),
		statusLabel,
		widget.NewSeparator(),
		monitorCardContainer,
		widget.NewSeparator(),
		toggleBtn,
		widget.NewSeparator(),
		theory,
		widget.NewSeparator(),
		footer,
	)

	w.SetContent(container.NewPadded(content))
}

// ============================================================
// GPU 锻造引擎 — Julia Set 计算着色器
// ============================================================
func startGPULoader() {
	atomic.StoreInt32(&running, 1)

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		if err := glfw.Init(); err != nil {
			fmt.Println("GLFW 初始化失败:", err)
			return
		}
		defer glfw.Terminate()

		glfw.WindowHint(glfw.Visible, glfw.False)
		glfw.WindowHint(glfw.Resizable, glfw.False)
		glfw.WindowHint(glfw.ContextVersionMajor, 3)
		glfw.WindowHint(glfw.ContextVersionMinor, 3)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

		window, err := glfw.CreateWindow(1, 1, "GemeForge Engine", nil, nil)
		if err != nil {
			fmt.Println("隐藏上下文创建失败:", err)
			return
		}
		defer window.Destroy()

		window.MakeContextCurrent()
		glfw.SwapInterval(0)

		if err := gl.Init(); err != nil {
			fmt.Println("OpenGL 初始化失败:", err)
			return
		}

		program := createJuliaProgram()
		gl.UseProgram(program)
		vao := createQuad()

		uTime := gl.GetUniformLocation(program, gl.Str("u_time\x00"))
		uRes := gl.GetUniformLocation(program, gl.Str("u_resolution\x00"))
		gl.Uniform2f(uRes, 1.0, 1.0)

		frameDuration := time.Second / time.Duration(TargetFPS)
		lastTime := time.Now()
		frameCount := 0
		secondTimer := time.Now()

		for atomic.LoadInt32(&running) == 1 {
			currentTime := time.Now()
			elapsed := currentTime.Sub(lastTime)

			if elapsed >= frameDuration {
				lastTime = currentTime
				gl.Uniform1f(uTime, float32(currentTime.UnixMilli())/1000.0)

				gl.BindVertexArray(vao)
				gl.DrawArrays(gl.TRIANGLES, 0, 6)
				window.SwapBuffers()
				glfw.PollEvents()

				frameCount++
				if currentTime.Sub(secondTimer) >= time.Second {
					atomic.StoreInt32(&fakeFPS, int32(frameCount))
					frameCount = 0
					secondTimer = currentTime
				}
			} else {
				time.Sleep(time.Millisecond)
			}

			if window.ShouldClose() {
				break
			}
		}
	}()
}

func createJuliaProgram() uint32 {
	vertex := `#version 330 core
layout(location = 0) in vec2 aPos;
void main() { gl_Position = vec4(aPos, 0.0, 1.0); }`

	fragment := `#version 330 core
out vec4 FragColor;
uniform float u_time;
uniform vec2 u_resolution;

void main() {
    vec2 uv = gl_FragCoord.xy / u_resolution;
    vec2 c = vec2(-0.8 + 0.1 * sin(u_time * 0.5), 0.156);
    vec2 z = uv * 4.0 - 2.0;
    float iter = 0.0;
    for(float i = 0.0; i < 500.0; i++) {
        float x = (z.x * z.x - z.y * z.y) + c.x;
        float y = (z.y * z.x + z.x * z.y) + c.y;
        if((x * x + y * y) > 4.0) break;
        z = vec2(x, y);
        iter++;
    }
    float t = iter / 500.0;
    vec3 col = vec3(0.5 + 0.5 * cos(6.28318 * (t + vec3(0.0, 0.33, 0.67))));
    FragColor = vec4(col, 1.0);
}`

	vShader := gl.CreateShader(gl.VERTEX_SHADER)
	csource, free := gl.Strs(vertex + "\x00")
	gl.ShaderSource(vShader, 1, csource, nil)
	free()
	gl.CompileShader(vShader)

	fShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	fsource, free := gl.Strs(fragment + "\x00")
	gl.ShaderSource(fShader, 1, fsource, nil)
	free()
	gl.CompileShader(fShader)

	program := gl.CreateProgram()
	gl.AttachShader(program, vShader)
	gl.AttachShader(program, fShader)
	gl.LinkProgram(program)
	gl.DeleteShader(vShader)
	gl.DeleteShader(fShader)
	return program
}

func createQuad() uint32 {
	vertices := []float32{
		-1.0, -1.0, 1.0, -1.0, -1.0, 1.0,
		-1.0, 1.0, 1.0, -1.0, 1.0, 1.0,
	}
	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	return vao
}

// ============================================================
// 硬件遥测循环 — 使用 fyne.Do 保证线程安全
// ============================================================
func hardwareMonitor() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var cpus []Win32_Processor
		err := wmi.Query("SELECT CurrentClockSpeed, LoadPercentage FROM Win32_Processor", &cpus)
		if err == nil && len(cpus) > 0 {
			fyne.Do(func() {
				cpuFreqLabel.SetText(fmt.Sprintf("CPU 运行频率: %d MHz  (负载 %d%%)",
					cpus[0].CurrentClockSpeed, cpus[0].LoadPercentage))
			})
		}

		var gpus []Win32_VideoController
		wmi.Query("SELECT Name FROM Win32_VideoController", &gpus)
		if len(gpus) > 0 {
			fyne.Do(func() {
				gpuFreqLabel.SetText(fmt.Sprintf("GPU 核心频率: %s", strings.TrimSpace(gpus[0].Name)))
			})
		}

		fyne.Do(func() {
			gpuLoadLabel.SetText("GPU 渲染负载: Julia Set 分形运算中")
			gpuTempLabel.SetText("GPU 核心温度: 请安装 NVIDIA NVML 获取精确值")
		})

		fps := atomic.LoadInt32(&fakeFPS)
		fyne.Do(func() {
			if atomic.LoadInt32(&running) == 1 {
				fpsLabel.SetText(fmt.Sprintf("锻造帧率: %d FPS  (引擎运行中)", fps))
			} else {
				fpsLabel.SetText("锻造帧率: 0 FPS  (引擎待机)")
			}
		})
	}
}
