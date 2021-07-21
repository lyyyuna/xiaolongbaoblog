title: Python 脚本实现读取图片属性信息
date: 2016-01-06 22:05:15
categories: 杂
tags: 
- Python script
- Exif
---

我之前的这篇文章《[暴力修改图片的 GPS 信息](http://www.lyyyuna.com/2015/11/14/exif-hack/)》中，讲述了如何从图片的二进制文件中读取 GPS 信息。

这里改用 Python 脚本读取图片信息。有几个说明

* 没有实现错误处理
* 没有读取所有信息，大概只有 GPS 信息、图片分辨率、图片像素、设备商、拍摄设备等
* 简单修改后应该能实现[暴力修改图片的 GPS 信息](http://www.lyyyuna.com/2015/11/14/exif-hack/)
* 但对于本身没有 GPS 信息的图片，实现则非常复杂，需要仔细计算每个描述符的偏移量

脚本运行后，读取结果如下

![脚本读取的信息](https://raw.githubusercontent.com/lyyyuna/blog_img/master/blog/201601/exif_info3.jpg)

这里和 Windows 属性查看器读到的内容完全一致

![图片信息1](https://raw.githubusercontent.com/lyyyuna/blog_img/master/blog/201601/exif_info1.jpg)

![图片信息2](https://raw.githubusercontent.com/lyyyuna/blog_img/master/blog/201601/exif_info2.jpg)

源码如下

    # -*- coding:utf-8 -*-
    import binascii

    class ParseMethod(object):
        @staticmethod
        def parse_default(f, count, offset):
            pass

        @staticmethod
        def parse_latitude(f, count, offset):
            old_pos = f.tell()
            f.seek(12 + offset)

            latitude = [0,0,0]
            for i in xrange(count):
                byte = f.read(4)
                numerator = byte.encode('hex')

                byte = f.read(4)
                denominator = byte.encode('hex')

                latitude[i] =  float(int(numerator, 16)) / int(denominator, 16)


            print 'Latitude:\t%.2f %.2f\' %.2f\"' % (latitude[0], latitude[1], latitude[2])
            f.seek(old_pos)    


        @staticmethod
        def parse_longtitude(f, count, offset):
            old_pos = f.tell()
            f.seek(12 + offset)

            longtitude = [0,0,0]
            for i in xrange(count):
                byte = f.read(4)
                numerator = byte.encode('hex')

                byte = f.read(4)
                denominator = byte.encode('hex')

                longtitude[i] =  float(int(numerator, 16)) / int(denominator, 16)


            print 'Longtitude:\t%.2f %.2f\' %.2f\"' % (longtitude[0], longtitude[1], longtitude[2])
            f.seek(old_pos) 

        @staticmethod
        def parse_make(f, count, offset):
            old_pos = f.tell()
            f.seek(12 + offset)
            byte = f.read(count)
            a = byte.encode('hex')
            print 'Make:\t\t' + binascii.a2b_hex(a)
            f.seek(old_pos) 

        @staticmethod
        def parse_model(f, count, offset):
            old_pos = f.tell()
            f.seek(12 + offset)
            byte = f.read(count)
            a = byte.encode('hex')
            print 'Model:\t\t' + binascii.a2b_hex(a)
            f.seek(old_pos)         

        @staticmethod
        def parse_datetime(f, count, offset):
            old_pos = f.tell()
            f.seek(12 + offset)
            byte = f.read(count)
            a = byte.encode('hex')
            print 'DateTime:\t' + binascii.a2b_hex(a)
            f.seek(old_pos)

        # rational data type, 05
        @staticmethod
        def parse_xresolution(f, count, offset):
            old_pos = f.tell()
            f.seek(12 + offset)

            byte = f.read(4)
            numerator = byte.encode('hex')
            byte = f.read(4)
            denominator = byte.encode('hex')
            xre = int(numerator, 16) / int(denominator, 16)

            print 'XResolution:\t' + str(xre) + ' dpi'
            f.seek(old_pos)

        @staticmethod
        def parse_yresolution(f, count, offset):
            old_pos = f.tell()
            f.seek(12 + offset)

            byte = f.read(4)
            numerator = byte.encode('hex')
            byte = f.read(4)
            denominator = byte.encode('hex')
            xre = int(numerator, 16) / int(denominator, 16)

            print 'YResolution:\t' + str(xre) + ' dpi'
            f.seek(old_pos)

        @staticmethod
        def parse_exif_ifd(f, count, offset):
            old_pos = f.tell()
            f.seek(12 + offset)

            byte = f.read(2)
            a = byte.encode('hex')        
            exif_ifd_number = int(a, 16)

            for i in xrange(exif_ifd_number):
                byte = f.read(2)
                tag_id = byte.encode('hex')
                #print tag_id,

                byte = f.read(2)
                type_n = byte.encode('hex')
                #print type_n,

                byte = f.read(4)
                count = byte.encode('hex')
                #print count,

                byte = f.read(4)
                value_offset = byte.encode('hex')
                #print value_offset

                value_offset = int(value_offset, 16)
                EXIF_IFD_DICT.get(tag_id, ParseMethod.parse_default)(f, count, value_offset)

            f.seek(old_pos)    

        @staticmethod
        def parse_x_pixel(f, count, value):
            print 'X Pixels:\t' + str(value)

        @staticmethod
        def parse_y_pixel(f, count, value):
            print 'y Pixels:\t' + str(value)

        @staticmethod
        def parse_gps_ifd(f, count, offset):
            old_pos = f.tell()        
            f.seek(12 + offset)
            byte = f.read(2)
            a = byte.encode('hex')   
            gps_ifd_number = int(a, 16)

            for i in xrange(gps_ifd_number):
                byte = f.read(2)
                tag_id = byte.encode('hex')
                #print tag_id,

                byte = f.read(2)
                type_n = byte.encode('hex')
                #print type_n,

                byte = f.read(4)
                count = byte.encode('hex')
                #print count,

                byte = f.read(4)
                value_offset = byte.encode('hex')
                #print value_offset

                count = int(count, 16)
                value_offset = int(value_offset, 16)
                GPS_IFD_DICT.get(tag_id, ParseMethod.parse_default)(f, count, value_offset)

            f.seek(old_pos)  

    IFD_dict = {
        '010f' : ParseMethod.parse_make ,
        '0110' : ParseMethod.parse_model ,
        '0132' : ParseMethod.parse_datetime ,
        '011a' : ParseMethod.parse_xresolution ,
        '011b' : ParseMethod.parse_yresolution ,
        '8769' : ParseMethod.parse_exif_ifd ,
        '8825' : ParseMethod.parse_gps_ifd
    }

    EXIF_IFD_DICT = {
        'a002' : ParseMethod.parse_x_pixel ,
        'a003' : ParseMethod.parse_y_pixel
    }

    GPS_IFD_DICT = {
        '0002' : ParseMethod.parse_latitude ,
        '0004' : ParseMethod.parse_longtitude
    }


    with open('image.jpg', 'rb') as f:
        byte = f.read(2)
        a = byte.encode('hex')
        print 'SOI Marker:\t' + a

        byte = f.read(2)
        a = byte.encode('hex')
        print 'APP1 Marker:\t' + a

        byte = f.read(2)
        a = byte.encode('hex')
        print 'APP1 Length:\t' + str(int(a, 16)) + ' .Dec'

        byte = f.read(4)
        a = byte.encode('hex')
        print 'Identifier:\t' + binascii.a2b_hex(a)

        byte = f.read(2)
        a = byte.encode('hex')
        print 'Pad:\t\t' + a 

        print 

        print 'Begin to print Header.... '
        print 'APP1 Body: '

        byte = f.read(2)
        a = byte.encode('hex')
        print 'Byte Order:\t' + a    

        byte = f.read(2)
        a = byte.encode('hex')
        print '42:\t\t' + a  

        byte = f.read(4)
        a = byte.encode('hex')
        print '0th IFD Offset:\t' + a  

        print 'Finish print Header'

        print 'Begin to print 0th IFD....'
        print
        #print 'Total: ',
        byte = f.read(2)
        a = byte.encode('hex')
        interoperability_number = int(a, 16)
        #print interoperability_number
        

        for i in xrange(interoperability_number):
            byte = f.read(2)
            tag_id = byte.encode('hex')
            #print tag_id,

            byte = f.read(2)
            type_n = byte.encode('hex')
            #print type_n,

            byte = f.read(4)
            count = byte.encode('hex')
            #print count,

            byte = f.read(4)
            value_offset = byte.encode('hex')
            #print value_offset

            count = int(count, 16)
            value_offset = int(value_offset, 16)
            
            # simulate switch
            IFD_dict.get(tag_id, ParseMethod.parse_default)(f, count, value_offset)


        print
        print 'Finish print 0th IFD....'


