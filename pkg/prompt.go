package pkg

import (
	"strings"

	cv "codeberg.org/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
)

// A small boxed prompt for one line of text, shown over a blank screen.
// Enter submits (non-empty text), Esc cancels back to the home screen.

var prompt struct {
	label *cv.TextView
	input *cv.InputField
	done  func(string)
}

func promptInit() cv.Primitive {
	prompt.label = textBox(cv.AlignLeft)
	prompt.input = cv.NewInputField()
	prompt.input.SetDoneFunc(func(key tc.Key) {
		text := strings.TrimSpace(prompt.input.GetText())
		prompt.input.SetText("")
		if key != tc.KeyEnter || text == "" || prompt.done == nil {
			gotoWelcome()
			return
		}
		prompt.done(text)
	})
	box := cv.NewFlex()
	box.SetDirection(cv.FlexRow)
	box.SetBorder(true)
	box.SetPadding(1, 0, 2, 2)
	box.AddItem(prompt.label, 3, 0, false)
	box.AddItem(prompt.input, 1, 0, true)
	return Center(72, 8, box, prompt.input)
}

// askText shows the prompt; done gets the text entered.
func askText(question, label string, done func(string)) {
	prompt.label.SetText(question + "\n[#a0a0a0](Enter to confirm, Esc to cancel)[-]")
	prompt.input.SetLabel(label)
	prompt.done = done
	Root.Switch("prompt")
	Root.App.SetFocus(prompt.input)
}

// promptError shows a problem with what was entered, keeping the prompt open.
func promptError(question, msg string) {
	prompt.label.SetText(question + "\n[red]" + cv.Escape(msg) + "[-]")
	Root.App.SetFocus(prompt.input)
}
