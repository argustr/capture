package capture

import (
	"errors"
	"fmt"
	"image"
	"unsafe"

	"github.com/sparkle4/logtrace"
	"github.com/sparkle4/win32-extend"
)

type (
	//SystemMetric screeen size etc
	SystemMetric struct {
		XPos, YPos, ScreenX, ScreenY, ScreenH, ScreenW int32
	}
)

func NewCaptureImage() (img *image.RGBA, err error) {
	virtualSystemMetric := getSystemMetrics()
	width := virtualSystemMetric.XPos
	height := virtualSystemMetric.YPos

	hdc := win32_extend.GetDC(0)
	if hdc == 0 {
		return nil, errors.New(fmt.Sprintf("%v:%s", errors.New("win.GetDC failed"), logtrace.WhereAmI()))
	}
	defer win32_extend.ReleaseDC(0, hdc)

	memoryDevice := win32_extend.CreateCompatibleDC(hdc)
	if memoryDevice == 0 {
		return nil, errors.New(fmt.Sprintf("%v:%s", errors.New("win.CreateCompatibleDC failed"), logtrace.WhereAmI()))
	}
	defer win32_extend.DeleteDC(memoryDevice)

	bitmap := win32_extend.CreateCompatibleBitmap(hdc, width, height)
	if bitmap == 0 {
		return nil, errors.New(fmt.Sprintf("%v:%s", errors.New("win.CreateCompatibleBitmap"), logtrace.WhereAmI()))
	}
	defer win32_extend.DeleteObject(win32_extend.HGDIOBJ(bitmap))

	header := win32_extend.BITMAPINFOHEADER{
		BiSize:        uint32(unsafe.Sizeof(win32_extend.BITMAPINFOHEADER{})),
		BiPlanes:      1,
		BiBitCount:    32,
		BiWidth:       width,
		BiHeight:      -height,
		BiCompression: win32_extend.BI_RGB, //Todo: uzak parametre     ,
		BiSizeImage:   0,
	}

	bitmapDataSize := uintptr(((int64(width)*int64(header.BiBitCount) + 31) / 32) * 4 * int64(height))
	hmem := win32_extend.GlobalAlloc(win32_extend.GMEM_MOVEABLE, bitmapDataSize)
	defer win32_extend.GlobalFree(hmem)
	memptr := win32_extend.GlobalLock(hmem)
	defer win32_extend.GlobalUnlock(hmem)

	old := win32_extend.SelectObject(memoryDevice, win32_extend.HGDIOBJ(bitmap))
	if old == 0 {
		return nil, errors.New(fmt.Sprintf("%v:%s", errors.New("win.SelectObject"), logtrace.WhereAmI()))
	}
	defer win32_extend.SelectObject(memoryDevice, old)

	win32_extend.BitBlt(
		memoryDevice,
		0,
		0,
		int32(virtualSystemMetric.ScreenW),
		virtualSystemMetric.ScreenH,
		hdc,
		virtualSystemMetric.ScreenX,
		virtualSystemMetric.ScreenY,
		win32_extend.SRCCOPY)

	if win32_extend.GetDIBits(hdc,
		bitmap,
		0,
		uint32(height),
		(*uint8)(memptr),
		(*win32_extend.BITMAPINFO)(unsafe.Pointer(&header)),
		win32_extend.DIB_RGB_COLORS) == 0 {
		return nil, errors.New(fmt.Sprintf("%v:%s", errors.New("win.GetDIBits"), logtrace.WhereAmI()))
	}

	img = image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	i := 0
	src := uintptr(memptr)
	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {
			v0 := *(*uint8)(unsafe.Pointer(src))
			v1 := *(*uint8)(unsafe.Pointer(src + 1))
			v2 := *(*uint8)(unsafe.Pointer(src + 2))

			img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = v2, v1, v0, 255
			i += 4
			src += 4
		}
	}
	return
}

func getSystemMetrics() SystemMetric {
	var systemMetric SystemMetric
	systemMetric.XPos = win32_extend.GetSystemMetrics(win32_extend.SM_CXVIRTUALSCREEN)
	systemMetric.YPos = win32_extend.GetSystemMetrics(win32_extend.SM_CYVIRTUALSCREEN)
	systemMetric.ScreenX = win32_extend.GetSystemMetrics(win32_extend.SM_XVIRTUALSCREEN)
	systemMetric.ScreenY = win32_extend.GetSystemMetrics(win32_extend.SM_YVIRTUALSCREEN)
	systemMetric.ScreenW = win32_extend.GetSystemMetrics(win32_extend.SM_CXVIRTUALSCREEN)
	systemMetric.ScreenH = win32_extend.GetSystemMetrics(win32_extend.SM_CYVIRTUALSCREEN)
	return systemMetric
}
