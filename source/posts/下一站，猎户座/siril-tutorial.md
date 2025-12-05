title: Siril 教程：天文摄影中的图像堆栈与处理
date: 2025-12-05 10:56:33

---

## 为什么使用 Siril？

![堆栈后的 M31 仙女座星系](46.png)

Siril 是一款免费的开源天文图像处理软件，提供精确的对齐和校准工具。它的堆栈算法可以降低噪声的同时保留细节，还具有背景提取和色彩校准功能，是天文摄影爱好者的理想选择。

## 安装 Siril

你可以从 [siril.org](https://siril.org) 下载安装最新的 Siril。

## 前置条件

在使用 Siril 之前，你需要：

- 对同一天体在相似条件下拍摄的多张图像
- 图像数量越多，堆栈后的效果越好
- 理想情况下还应该有暗场、偏置场和平场图像用于校准

## 实践天文摄影

### 文件组织

首先，创建一个主目录来存放你的项目文件：

![创建主目录](one.png)

然后在主目录中创建以下子目录：

![创建子目录](two.png)

- **lights** - 存放你的目标天体图像
- **darks** - 存放暗场图像（用于校准热噪声）
- **biases** - 存放偏置场图像
- **flats** - 存放平场图像（用于校准光学系统的不均匀性）

将你的目标天体图像放入 `lights` 文件夹：

![放置光学文件](three.png)

## 图像校准

### 启动 Siril

打开 Siril 软件：

![打开 Siril](four.png)

### 设置工作目录

1. 点击主界面上的"主页"图标：

![主页图标位置](five~2.png)

2. 在文件浏览器中导航到你创建的主目录：

![文件浏览器界面](seven~2.png)

3. 选择并确认目录：

![确认目录选择](eight~2.png)

### 安装脚本

Siril 提供了自动化处理脚本，可以大大简化工作流程。

1. 点击三横线菜单图标：

![三横线菜单图标](nine~2.png)

2. 选择"Get Scripts"选项：

![Get Scripts 选项](ten~2.png)

3. 这会打开 GitLab 链接，访问脚本仓库：

![GitLab 链接](111.png)

4. 选择"Preprocessing"目录：

![Preprocessing 选择](122.png)

5. 下载 \`OSC_Preprocessing_WithoutDBF.ssf\` 脚本（用于不含深空背景扁平化的一步式彩色相机预处理）：

![OSC_Preprocessing_WithoutDBF.ssf 脚本](13.png)

6. 下载脚本文件：

![下载脚本](144.png)

### 堆栈图像

1. 回到 Siril，点击"Scripts"菜单：

![Scripts 菜单](15~2.png)

2. 选择你下载的脚本：

![选择脚本](16~2.png)

3. 点击"Run"按钮开始执行：

![Run 按钮](17~2.png)

4. 控制台窗口会显示处理进度：

![控制台进度窗口](18~2.png)

脚本会自动完成以下步骤：
- 校准图像（使用暗场、偏置场、平场）
- 对齐所有图像
- 堆栈图像以降低噪声

### 查看堆栈结果

1. 处理完成后，点击"Open"菜单：

![Open 菜单选项](19~2.png)

2. 选择生成的 \`result.fit\` 文件：

![选择 result.fit 文件](20~2.png)

3. 你会看到原始堆栈图像：

![原始堆栈图像显示](21~2.png)

4. 切换到"Autostretch"模式以查看更多细节：

![Linear 到 Autostretch 切换](22~2.png)

![Autostretch 视图](23~2.png)

![Autostretch 结果](24~2.png)

## 背景提取

背景提取可以去除光污染和渐变，使天体更加突出。

1. 打开直方图视图：

![直方图视图选项](26~2.png)

2. 使用裁剪工具去除图像边缘的黑边：

![裁剪工具界面](27~3.png)

3. 打开"Image Processing"菜单：

![Image Processing 菜单](28.png)

4. 选择"Background Extraction"：

![Background Extraction 选择](29.png)

5. 点击"Generate"生成采样点：

![生成采样点](33.png)

6. 手动移除靠近亮星和目标天体的采样点，避免影响提取效果：

![在亮星附近移除采样点](30.png)

7. 点击"Compute Background"和"Apply"应用背景提取：

![计算背景并应用](31.png)

## 光度色彩校准

色彩校准可以根据星表数据校正图像的色彩平衡。

1. 打开"Color Calibration"菜单：

![Color Calibration 菜单](50.png)

2. 选择"Photometric Color Calibration"：

![Photometric Color Calibration](51.png)

3. 输入天体的星表编号（例如 M31 输入 "M31" 或 "NGC 224"）：

![输入星表编号窗口](52.png)

软件会自动从在线星表获取数据并进行色彩校准。

## 去除绿噪

数码相机拍摄的天文图像常常会有绿色噪声，需要去除。

1. 选择"Remove green noise"选项：

![Remove green noise 选项](34.png)

2. 调整去噪设置：

![Remove green noise 设置](35.png)

3. 应用后的结果：

![去噪后的结果](36.png)

## 拉伸图像

图像拉伸可以增强暗部细节，使天体更加明显。

1. 打开"Histogram Transformation"：

![Histogram Transformation](40.png)

2. 点击"Autostretch"算法图标：

![Autostretch 算法图标](41.png)

3. 选择"Asinh"变换：

![Asinh 变换选择](37.png)

4. 调整拉伸因子（Stretch Factor）直到获得满意的效果：

![拉伸因子调整](38.png)

## 应用色彩

增强色彩饱和度可以让天体颜色更加鲜艳。

1. 选择"Color Saturation"选项：

![Color Saturation 选项](42.png)

2. 调整饱和度参数：

![色彩饱和度调整](43.png)

## 保存图像

1. 点击保存/下载按钮：

![保存/下载按钮](47.png)

2. 选择文件格式（推荐 TIFF 或 PNG 以保留最多细节）：

![文件格式选择](49.png)

## 总结

本教程介绍了使用 Siril 进行天文摄影图像堆栈和后期处理的基本流程，包括：

- 文件组织和工作目录设置
- 使用脚本自动化校准和堆栈
- 背景提取去除光污染
- 光度色彩校准
- 去除绿噪
- 图像拉伸增强细节
- 色彩饱和度调整

这是一个面向初学者的指南，还有很多高级技术（如去卷积、StarNet 处理等）未涉及。如果你想深入学习，推荐访问 [Deep Space Astro YouTube 频道](https://www.youtube.com/c/DeepSpaceAstro)。

## 致谢

感谢 Siril 开发团队提供如此优秀的免费软件，以及天文摄影社区的无私分享。

---

*本教程翻译自 [https://sathvikacharyaa.github.io/sirilastro/](https://sathvikacharyaa.github.io/sirilastro/)*
