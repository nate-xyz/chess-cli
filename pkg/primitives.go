package pkg

import (
	"strings"

	cv "codeberg.org/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
)

func titlePrimitive(text string, toFocus cv.Primitive) *cv.TextView {
	tv := cv.NewTextView()
	tv.SetTextAlign(cv.AlignCenter)
	tv.SetVerticalAlign(cv.AlignMiddle)
	tv.SetDynamicColors(true)
	tv.SetWordWrap(true) // a narrow column wraps the title instead of cutting it
	tv.SetText(text)
	forwardKeys(tv, toFocus)
	return tv
}

func ribbonPrimitive(text string, toFocus cv.Primitive) *cv.TextView {
	tv := cv.NewTextView()
	tv.SetTextAlign(cv.AlignLeft)
	tv.SetVerticalAlign(cv.AlignTop)
	tv.SetBackgroundColor(tc.ColorKhaki)
	tv.SetTextColor(tc.ColorFireBrick)
	tv.SetText(keysText(text))
	if strings.Contains(text, "arrows/hjkl") {
		keyRibbons[tv] = text
	}
	forwardKeys(tv, toFocus)
	return tv
}

func quoutePrimitive(text string, toFocus cv.Primitive) *cv.TextView {
	tv := cv.NewTextView()
	tv.SetTextAlign(cv.AlignCenter)
	tv.SetVerticalAlign(cv.AlignTop)
	tv.SetTextColor(tc.ColorLightSlateGray)
	tv.SetWordWrap(true)
	tv.SetPadding(1, 1, 10, 10)
	tv.SetText(text)
	forwardKeys(tv, toFocus)
	return tv
}

func NewOptionWindow(msg, op1, op2 string, handler1, handler2 func()) *cv.Flex {
	subFlex := cv.NewFlex()
	button1 := cv.NewButton(op1)
	button1.SetSelectedFunc(handler1)
	button2 := cv.NewButton(op2)
	button2.SetSelectedFunc(handler2)
	spacer := cv.NewBox()
	spacer.SetBackgroundColor(tc.ColorTeal)
	subFlex.SetBackgroundColor(tc.ColorTeal)
	subFlex.AddItem(spacer, 0, 1, false)
	subFlex.AddItem(button1, 0, 2, true)
	subFlex.AddItem(spacer, 0, 1, false)
	subFlex.AddItem(button2, 0, 2, false)
	subFlex.AddItem(spacer, 0, 1, false)
	flex := cv.NewFlex()
	flex.SetDirection(cv.FlexRow)
	tv := cv.NewTextView()
	tv.SetBackgroundColor(tc.ColorTeal)
	tv.SetTextAlign(cv.AlignCenter)
	tv.SetVerticalAlign(cv.AlignMiddle)
	tv.SetText(msg)
	flex.SetBackgroundColor(tc.ColorTeal)
	flex.AddItem(tv, 0, 3, false)
	flex.AddItem(subFlex, 0, 1, true)
	flex.AddItem(spacer, 0, 1, true)
	flex.SetBorder(true)
	return flex
}

func Center(width, height int, p cv.Primitive, toFocus cv.Primitive) cv.Primitive {
	spacer := cv.NewBox()
	forwardKeys(spacer, toFocus)
	subFlex := cv.NewFlex()
	subFlex.SetDirection(cv.FlexRow)
	subFlex.AddItem(spacer, 0, 1, false)
	subFlex.AddItem(p, height, 1, true)
	subFlex.AddItem(spacer, 0, 1, false)
	flex := cv.NewFlex()
	flex.AddItem(spacer, 0, 1, false)
	flex.AddItem(subFlex, width, 1, true)
	flex.AddItem(spacer, 0, 1, false)
	return flex
}
