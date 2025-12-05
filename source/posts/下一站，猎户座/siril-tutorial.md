title: Siril 教程：天文摄影中的图像堆栈与处理
date: 2025-12-05 10:56:33

---

这是一份使用 Siril 进行天文摄影图像堆栈的综合指南，Siril 是一款功能强大且免费的天文图像处理软件。

天文摄影让我们能够捕捉夜空之美，但由于长曝光时间和低光条件，单张图像往往包含噪声且缺乏细节。为了克服这一问题，天文摄影师使用一种称为图像堆栈的技术。通过组合同一场景的多张图像，我们可以增强信噪比并揭示更多天体细节。

在本指南中，我将带你了解使用 Siril 堆栈天文摄影图像的过程，Siril 是一款专为天文图像处理设计的功能强大的免费软件。

## 为什么使用 Siril？

![堆栈后的 M31 仙女座星系](46.png)

Siril 作为专业的天文摄影图像堆栈软件脱颖而出，提供精确的对齐和校准工具，可以校正由大气和光学因素引起的偏移和畸变。它的堆栈算法在降低噪声的同时保留细节，还具有背景提取和色彩校准等功能。该软件同时支持自动化工作流程和手动控制，适合从初学者到高级用户的各类人群。作为免费的开源软件，Siril 持续更新并对爱好者开放。

## 安装 Siril

你可以从 [siril.org](https://siril.org) 下载安装最新的 Siril。

**安装命令：**
- Ubuntu/Debian: `sudo apt install siril`
- Fedora: `sudo dnf install siril`
- Arch Linux: `sudo pacman -S siril`

## 前置条件

需要在相似条件下以一致的设置拍摄同一天体或天区的多张图像，以获得最佳效果。由于噪声减少和细节增强，图像数量越多，最终效果越好。

## 实践天文摄影图像堆栈

练习图像来源：
- [Astropix](https://astropix.com)
- [AstroBackyard](https://astrobackyard.com)

### 在 Siril 中组织图像文件以进行堆栈

创建一个以目标天体命名的主目录（例如，M31 - 仙女座星系）。

![创建主目录](one.png)

在主目录内创建四个子目录：

![创建子目录](two.png)

- **lights** - 存放要堆栈的目标天体原始图像
- **darks** - 存放暗场图像，用于传感器噪声校正
- **biases** - 存放偏置场图像，用于读出噪声校正
- **flats** - 存放平场图像，用于暗角和灰尘校正

**重要提示：** "这些目录至关重要，因为 Siril 的内置脚本依赖它们来执行堆栈过程。如果这些目录没有正确创建并填充相应的图像，脚本将无法运行。"

将你的目标天体图像放入 `lights` 文件夹：

![放置光学文件](three.png)

## 图像校准

### 启动 Siril

打开 Siril 软件：

![打开 Siril](four.png)

### 设置工作区

1. 打开 Siril
2. 点击左上角的主页图标：

![主页图标位置](five~2.png)

3. 导航到你的图像目录：

![文件浏览器界面](seven~2.png)

4. 点击"Open"将其设置为工作目录：

![确认目录选择](eight~2.png)

确认目录显示在左上角。

### 安装脚本

1. 点击右上角的三横线图标：

![三横线菜单图标](nine~2.png)

2. 选择"Get Scripts"：

![Get Scripts 选项](ten~2.png)

3. 导航到 Siril 文档，在"Getting More Scripts"下找到 GitLab 链接：

![GitLab 链接](111.png)

4. 点击"preprocessing"：

![Preprocessing 选择](122.png)

5. 选择"OSC_Preprocessing_WithoutDBF.ssf"：

![OSC_Preprocessing_WithoutDBF.ssf 脚本](13.png)

6. 下载脚本：

![下载脚本](144.png)

将下载的脚本放置在：
`/This PC/Local Disk (C:)/Program Files/Siril/scripts/OSC_Preprocessing_WithoutDBF.ssf`

### 堆栈图像

1. 在顶部菜单中点击"Scripts"：

![Scripts 菜单](15~2.png)

2. 选择 OSC_Preprocessing_WithoutDBF：

![选择脚本](16~2.png)

3. 点击"Run"：

![Run 按钮](17~2.png)

4. 在控制台窗口监控进度：

![控制台进度窗口](18~2.png)

最终堆栈图像以 FITS 格式保存为 **result.fit**。

5. 在顶部菜单中点击"Open"：

![Open 菜单选项](19~2.png)

6. 选择 result.fit 并点击"Open"：

![选择 result.fit 文件](20~2.png)

7. 在线性视图中查看原始堆栈图像：

![原始堆栈图像显示](21~2.png)

8. 从 Linear 切换到 Autostretch 以获得更好的可见性：

![Linear 到 Autostretch 切换](22~2.png)

![Autostretch 视图](23~2.png)

![Autostretch 结果](24~2.png)

9. 访问直方图视图以控制色调范围

**注意：** "Autostretch 对图像数据应用临时拉伸，使微弱的细节更加可见，而不会永久改变原始数据。"

## 背景提取

背景提取是天文摄影图像处理中的关键步骤。它有助于从图像中去除不需要的背景噪声或变化，使天体更加清晰地突出。在本节中，我们将介绍在 Siril 中执行背景提取的步骤。

Siril 使用这些样本来估算由光污染或传感器噪声等因素引起的背景强度和渐变。

要开始后期处理，首先裁剪边缘周围的伪影和噪声，这些在直方图视图中是可见的。

1. 使用左键选择，然后右键点击"Crop"，裁剪掉边缘的伪影和噪声：

![裁剪工具界面](27~3.png)

2. 切换回 Autostretch 视图：

![直方图视图选项](26~2.png)

3. 导航到"Image Processing"：

![Image Processing 菜单](28.png)

4. 选择"Background Extraction"：

![Background Extraction 选择](29.png)

5. 生成采样点：

![生成采样点](33.png)

6. 右键点击移除太靠近亮星的采样点：

![在亮星附近移除采样点](30.png)

7. 点击"Compute Background"并应用：

![计算背景并应用](31.png)

## 光度色彩校准

光度色彩校准在天文摄影图像堆栈中至关重要，用于校正由光污染、传感器特性和不同曝光条件等因素引起的色彩不平衡。

该过程调整堆栈图像的颜色以匹配天体的真实颜色，确保准确且自然的结果。Siril 通过分析图像中恒星和其他天体的颜色，并调整红、绿、蓝通道来去除任何不需要的色偏，从而实现这一目标。

1. 点击"Color Calibration"：

![Color Calibration 菜单](50.png)

2. 选择"Photometric Color Calibration"：

![Photometric Color Calibration](51.png)

3. 输入星表编号（对于 M31，使用 NGC 224）：

![输入星表编号窗口](52.png)

4. 从以下星表中选择：VizieR、SIMBAD 或 CDS
5. 添加焦距和像素尺寸以提高准确度
6. 点击"OK"

## 去除绿噪

绿噪是天文摄影图像中的常见问题，特别是在使用某些相机或传感器时。它表现为图像中的绿色色调或色偏，会降低最终结果的整体质量和准确性。

去除绿噪对于获得更加平衡和自然的图像至关重要。

1. 点击"Remove green noise"：

![Remove green noise 选项](34.png)

2. 保持默认选项不变：

![Remove green noise 设置](35.png)

3. 点击"Apply"：

![去噪后的结果](36.png)

## 拉伸图像

拉伸图像是天文摄影图像处理中的关键步骤。它增强了堆栈图像中微弱细节和颜色的可见性，使其在视觉上更具吸引力。

1. 将视图更改为 Linear 并点击"Histogram Transformation"：

![Histogram Transformation](40.png)

2. 应用 autostretch 算法：

![Autostretch 算法图标](41.png)

3. 选择 Asinh 变换：

![Asinh 变换选择](37.png)

4. 将 Stretch Factor 调整到最大值：

![拉伸因子调整](38.png)

5. 点击"Apply"

## 应用色彩

色彩饱和度调整可以增强图像中天体的颜色，使它们更加生动和视觉上更具吸引力。在本节中，我们将使用 Siril 的色彩饱和度工具来提升图像的整体色彩强度。

1. 点击"Color Saturation"：

![Color Saturation 选项](42.png)

2. 选择 Global
3. 通过数量滑块增加色彩饱和度：

![色彩饱和度调整](43.png)

4. 点击"Apply"

## 保存图像

1. 从右上角保存或下载：

![保存/下载按钮](47.png)

2. 选择所需的文件格式：

![文件格式选择](49.png)

3. 保存以下载最终图像

## 结论

"Siril 是天文摄影师的一款极其强大的工具，提供了一系列功能，使图像堆栈和校准等复杂任务变得更加容易。"本指南面向使用望远镜的初学者，不涵盖高级技术，如双曲线拉伸变换、去卷积或噪声减少。对于高级技术，作者推荐 YouTube 上的 Deep Space Astro 频道。

## 致谢

图像来源于 [Astropix](https://astropix.com) 和 [AstroBackyard](https://astrobackyard.com)，感谢他们通过高质量天文摄影资源做出的"宝贵贡献"。

---

*本教程翻译自 [https://sathvikacharyaa.github.io/sirilastro/](https://sathvikacharyaa.github.io/sirilastro/)*

