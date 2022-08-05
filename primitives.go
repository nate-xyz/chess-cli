package main

import (
	cv "code.rocketnine.space/tslocum/cview"
	tc "github.com/gdamore/tcell/v2"
)

func titlePrimitive(text string) cv.Primitive {
	tv := cv.NewTextView()
	tv.SetTextAlign(cv.AlignCenter)
	tv.SetVerticalAlign(cv.AlignMiddle)
	tv.SetDynamicColors(true)
	tv.SetText(text)
	return tv
}

func ribbonPrimitive(text string) cv.Primitive {
	tv := cv.NewTextView()
	tv.SetTextAlign(cv.AlignLeft)
	tv.SetVerticalAlign(cv.AlignTop)
	tv.SetBackgroundColor(tc.ColorKhaki)
	tv.SetTextColor(tc.ColorFireBrick)
	tv.SetText(text)
	return tv
}

func quoutePrimitive(text string) cv.Primitive {
	tv := cv.NewTextView()
	tv.SetTextAlign(cv.AlignCenter)
	tv.SetVerticalAlign(cv.AlignTop)
	tv.SetTextColor(tc.ColorLightSlateGray)
	tv.SetText(text)
	return tv
}

func boardPrimitive(handler func(row, col int)) *cv.Table {
	table := cv.NewTable()
	table.SetSelectable(true, true)
	//table.SetSelectedFunc(tableHandler)
	table.SetSelectionChangedFunc(handler)
	table.SetSortClicked(false)
	table.SetFixed(11, 11)
	return table
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

func Center(width, height int, p cv.Primitive) cv.Primitive {
	subFlex := cv.NewFlex()
	subFlex.SetDirection(cv.FlexRow)
	subFlex.AddItem(cv.NewBox(), 0, 1, false)
	subFlex.AddItem(p, height, 1, true)
	subFlex.AddItem(cv.NewBox(), 0, 1, false)
	flex := cv.NewFlex()
	flex.AddItem(cv.NewBox(), 0, 1, false)
	flex.AddItem(subFlex, width, 1, true)
	flex.AddItem(cv.NewBox(), 0, 1, false)
	return flex
}
