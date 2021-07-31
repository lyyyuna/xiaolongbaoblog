title: 修改图片的 GPS 信息
date: 2015-11-14 09:03:39
categories: 杂
tags: Exif
---

## 序

在一些应用中上传图片可以显示拍摄地信息，这是因为使用带 GPS 功能的照相机或者是智能手机拍照后，会在图片中含有拍摄地的 GPS 信息。查看图片的详细信息之后，你会发现如下图所示记录了经度、纬度和海拔。

![属性查看](/img/blog/201511/1.png)

在 Windows 上，详细信息中的时间等可以直接修改，但唯独 GPS 信息是不可修改。本着研究探索的精神，我打算试试能不能直接修改二进制文件的方法来修改 GPS，将任意一幅图片修改为在北京拍摄。通过谷歌，了解到图片信息是由一种 Exchangeable image file format (Exif) 的格式来描述。网上我能查到的最新 specification 版本为 2012 年的 Exif v2.3 版本。[Exif 规格书链接](http://www.cipa.jp/std/documents/e/DC-008-2012_E.pdf)

### Exif 规范的图片

Exif 规范其实包含了图片和音频两部分内容，这里我们只关心图片。整个图片的结构如下：

*   File Header 
*   0th IFD
*   0th IFD Value
*   1st IFD
*   1st IFD Value
*   1st (Thumbnail) Image Data
*   0th (Primary) Image Data

![图片文件结构](/img/blog/201511/2.png)

可以看到，IFD 记录了如图片长宽的信息，EXIF IFD 记录图片的拍摄信息，GPS IFD 则记录了 GPS。

### IFD 结构

一个 IFD 由四部分组成，每一个 IFD 都是固定的 12 个字节，分别是

*   Bytes 0-1   Tag
*   Bytes 2-3   Type
*   Bytes 4-7   Count
*   Bytes 8-11  Value Offset

*Tag* 是标记这个 IFD 的类别。

*Type* 是指数据的类型，有 BYTE(1, 8-bit), ASCII(2, 8-bit), SHORT(3, 2-byte), LONG(4, 4-byte), RATIONAL(5, 2-long) 等。其中 RATIONAL 的第一个 long 是分子，第二个 long 是分母。

*Count* 是指数据的数量，比如纬度就用度、分、秒三个数来描述。

*Value Offset* 是指真实数据所在的偏移地址（相对于 File header），而且需要注意的是，这里记录的值小于 4 个字节，则数据左对齐。

### Exif IFD 结构

EXif IFD Pointer

    Tag     = 34665(8769.H)
    Type    = LONG 
    Count   = 1
    Default = None

GPS IFD Pointer 

    Tag     = 34853(8825.H)
    Type    = LONG 
    Count   = 1 
    Default = None

### GPS 属性信息

我们这边只关心 GPS 的经纬度，与之相关的一些 Tag 信息如下。

GPSLatitudeRef

    Tag     = 1
    Type    = ASCII  
    Count   = 2
    Default = None 
    'N'     = North latitude
    'S'     = South latitude 
    Other   = reserved

GPSLatitude 

    Tag     = 2
    Type    = RATIONAL
    Count   = 3 
    Default = None

GPSLongtitudeRef

    Tag     = 3
    Type    = ASCII  
    Count   = 2
    Default = None 
    'E'     = East longtitude
    'W'     = West longtitude
    Other   = reserved

GPSLongtitude

    Tag     = 4
    Type    = RATIONAL
    Count   = 3 
    Default = None

## 实战

### 寻找头部

用二进制编辑器打开手机拍摄的图片。阅读了 Specification 之后，我们了解到图片文件开始处所包含的信息为 SOI Marker(FFD8.H), APP1 Marker(FFE1.H), APP1 Length(xxxx.H), Identifier('Exif'), Pad(00.H), APP1 Body(File Header, 0th IFD, 0th IFD Value, ...)。

查看一下二进制，

![图片起始](/img/blog/201511/3.png)

可以看到真的有 FFD8.H, FFE1.H, 后面的 45.H, 78.H, 69.H, 66.H，则正好是 'Exif' 的 ASCII 值。

接下来 8 个字节，分别是 Byte Order(4D4D.H), 42(002A.H), 0th IFD Offset(00000008.H) 和 Number of Interoperability(000B.H)。让我们检察一下二进制文件。

![APP1 Body](/img/blog/201511/4.png)

接下来就是各种 IFD 的信息，我们需要先找到我们关心的 GPS IFD Pointer, 其 Tag = 8825.H。搜索一下，找到啦，在地址 0000008e.H 处。

![GPS IFD Pointer](/img/blog/201511/5.png)

### 修改 GPS

按照之前描述的 IFD 结构，我们将接下来的 10 个字节如图分割，可以看到其指出真正的 GPS 信息位于地址 03EC.H 处。由于这里的地址是相对于 File Header (也就是地址 000C.H)，所以真实的地址为 03EC.H + 000C.H = 03F8.H。

![偏移地址](/img/blog/201511/6.png)

#### 第一个 IFD

![GPSLatitudeRef](/img/blog/201511/7.png)

开始两个字节为 GPS IFD Number，忽略。接下来的 4 个 IFD 就是我们需要修改的 GPS 信息。我们可以看出这是一个 GPSLatitudeRef，其中 4E000000.H 是 'N' 的 ASCII 码值，而且之前说过数据必须左对齐。这个我们不需要修改。

#### 第二个 IFD

![N](/img/blog/201511/8.png)

我们可以看出这是一个 GPSLatitude, 其指出真正 GPS 纬度信息存放在地址 049A.H + 000C.H = 04A6.H 处。

![纬度](/img/blog/201511/9.png)

1F.H / 01.H = 31, 1D.H / 01.H = 29, 0C71.H / 64.H = 31。计算结果与 Windows 右键属性查看的完全一致。

![原纬度值](/img/blog/201511/10.png)

谷歌搜索得知北京天安门位于北纬 39(27.H) 度 54(36.H) 分，那我们只需要修改对应位就行。

![修改后纬度值](/img/blog/201511/11.png)

#### 第三/四个 IFD

我们可以找到并修改经度信息。04B2.H + 0C.H = 04BE.H

![原经度值](/img/blog/201511/12.png)

北京天安门位于东经 116 度 23 分。

![现经度值](/img/blog/201511/13.png)

再用 Windows 右键属性查看一下。

![属性查看](/img/blog/201511/14.png)

已经在北京了。

我们可以用 QQ 上传照片的功能验证一下。

![QQ](/img/blog/201511/15.png)

确实在北京天安门附近。

## Conclusion

对于不带 GPS 信息的图片怎么办呢，那就比较麻烦了，因为其本身不含 GPS IFD，要手动插入 IFD，计算各个偏移量，再将后面的 jpeg 数据整体后移。

该研究起初源于我想恶作剧在空间传照片时显示任意地点，结果等我弄完后发现其实 QQ 支持在上传添加地点时候乱填 -_-

最后是我用来实验的图片，摄于江南大学，2015.2.11，iphone 5c.

![flower](/img/blog/201511/IMG_0596.JPG)