package pkg

import (
	"fmt"

	cv "codeberg.org/tslocum/cview"
)

func initConstruct() *cv.Grid {
	grid := cv.NewGrid()

	grid.SetBorders(false)

	tv := cv.NewTextView()
	tv.SetTextAlign(cv.AlignCenter)
	tv.SetVerticalAlign(cv.AlignMiddle)
	tv.SetDynamicColors(true)

	tree := cv.NewTreeView()
	treeRoot := cv.NewTreeNode("New Challenge")
	tree.SetRoot(treeRoot)

	path := []string{}

	list := cv.NewList()
	list.SetWrapAround(true)
	list.SetHover(true)

	Ribbon := ribbonPrimitive(challengeRibbonstr, list)

	challengeTypeOption := []string{"Random", "Friend", "AI"}
	challengeTypeOptionExplain := []string{"Seek a random player.", "Challenge a friend.", "Play against lichess bot."}
	variantOption := []string{"standard", "chess960", "crazyhouse", "antichess", "atomic", "horde", "kingOfTheHill", "racingKings", "threeCheck"}
	timeOptions := []string{"real time", "correspondence", "unlimited"}
	ratedOption := []string{"casual", "rated"}
	colorOptions := []string{"random", "white", "black"}

	var firstOption func()
	var variantSecondOption func()
	var timeThirdOption func()
	var fourthIntervalOption func()
	var fifthRatedOption func()
	var sixthColorOption func()
	var BotPowerLevelOption func()
	var seventhfriendsOptions func()
	var eightSubmitOption func()
	var updateTree func([]string)
	// 10+0 by default: rapid, which Lichess allows for every kind of game
	// started through the Board API (bullet isn't allowed at all)
	var Minutes float64 = 10
	var Seconds int
	var Days int = 1
	var BotPowerLevel int = 1

	// The time and bot-level steps put a form of sliders where the menu is;
	// showList puts the menu back.
	showList := func() {
		grid.Clear()
		grid.AddItem(tv, 0, 0, 1, 2, 0, 0, false)
		grid.AddItem(list, 1, 1, 1, 1, 0, 0, true)
		grid.AddItem(Ribbon, 2, 0, 1, 2, 0, 0, false)
		grid.AddItem(tree, 1, 0, 1, 1, 0, 0, false)
		Root.App.SetFocus(list)
	}
	showForm := func(form *cv.Form) {
		grid.Clear()
		grid.AddItem(tv, 0, 0, 1, 2, 0, 0, false)
		grid.AddItem(form, 1, 1, 1, 1, 0, 0, true)
		grid.AddItem(Ribbon, 2, 0, 1, 2, 0, 0, false)
		grid.AddItem(tree, 1, 0, 1, 1, 0, 0, false)
		Root.App.SetFocus(form) // the menu it replaces had the keyboard
	}

	goHome := func() {
		showList()
		path = []string{}
		firstOption()
		gotoLichessAfterLogin()
	}

	//choose the type of challenge
	firstOption = func() {
		list.Clear()
		updateTree(challengeTypeOption)
		tv.SetText("Select challenge type.")
		for i := 0; i < len(challengeTypeOption); i++ {
			item := cv.NewListItem(challengeTypeOption[i])
			item.SetSecondaryText(challengeTypeOptionExplain[i])
			item.SetShortcut(rune('a' + i))

			item.SetSelectedFunc(func() {
				newChallenge.Type = list.GetCurrentItemIndex() //store the choice
				path = append(path, challengeTypeOption[list.GetCurrentItemIndex()])
				updateTree(variantOption)
				variantSecondOption() //add the new list
			})
			list.AddItem(item)
		}
		//add back
		item := cv.NewListItem("Back")
		item.SetSecondaryText("Back to Lichess Welcome")
		item.SetShortcut(rune('z'))
		item.SetSelectedFunc(goHome)
		list.AddItem(item)

		//add quit
		quitItem := cv.NewListItem("Quit")
		quitItem.SetSecondaryText("Press to exit")
		quitItem.SetShortcut('q')
		quitItem.SetSelectedFunc(func() {
			Root.App.Stop()
		})
		list.AddItem(quitItem)

	}

	variantSecondOption = func() {
		list.Clear()

		tv.SetText("Select challenge variant.")
		for i := 0; i < len(variantOption); i++ {
			item := cv.NewListItem(variantOption[i])
			item.SetShortcut(rune('a' + i))
			item.SetSelectedFunc(func() {

				newChallenge.Variant = variantOption[list.GetCurrentItemIndex()] //store the choice

				path = append(path, variantOption[list.GetCurrentItemIndex()])
				updateTree(timeOptions)

				timeThirdOption() //add the new list
			})
			list.AddItem(item)
		}
		//add back
		item := cv.NewListItem("Back")
		item.SetSecondaryText("Back to selecting the CHALLENGE type.")
		item.SetShortcut(rune('y'))
		item.SetSelectedFunc(func() {
			if len(path) > 0 {
				path = path[:len(path)-1]
			}
			updateTree(challengeTypeOption)
			firstOption()
		})
		list.AddItem(item)

		//add home
		item = cv.NewListItem("Home")
		item.SetSecondaryText("Back to Lichess Welcome")
		item.SetShortcut(rune('z'))
		item.SetSelectedFunc(goHome)
		list.AddItem(item)

		//add quit
		quitItem := cv.NewListItem("Quit")
		quitItem.SetSecondaryText("Press to exit")
		quitItem.SetShortcut('q')
		quitItem.SetSelectedFunc(func() {
			Root.App.Stop()
		})
		list.AddItem(quitItem)

	}

	timeThirdOption = func() {
		list.Clear()

		tv.SetText("Select time variant.")

		for i := 0; i < len(timeOptions); i++ {
			item := cv.NewListItem(timeOptions[i])
			item.SetShortcut(rune('a' + i))
			item.SetSelectedFunc(func() {
				newChallenge.TimeOption = list.GetCurrentItemIndex() //store the choice

				path = append(path, timeOptions[list.GetCurrentItemIndex()])

				//clear the list
				if newChallenge.TimeOption < 2 { //add the new list
					updateTree([]string{"Select interval."})
					fourthIntervalOption()
				} else {
					updateTree(ratedOption)
					fifthRatedOption()
				}
			})

			list.AddItem(item)
		}
		//add back
		item := cv.NewListItem("Back")
		item.SetSecondaryText("Back to selecting the VARIANT type.")
		item.SetShortcut(rune('y'))
		item.SetSelectedFunc(func() {
			if len(path) > 0 {
				path = path[:len(path)-1]
			}
			updateTree(variantOption)
			variantSecondOption()
		})
		list.AddItem(item)

		//add home
		item = cv.NewListItem("Home")
		item.SetSecondaryText("Back to Lichess Welcome")
		item.SetShortcut(rune('z'))
		item.SetSelectedFunc(goHome)
		list.AddItem(item)

		//add quit
		quitItem := cv.NewListItem("Quit")
		quitItem.SetSecondaryText("Press to exit")
		quitItem.SetShortcut('q')
		quitItem.SetSelectedFunc(func() {
			Root.App.Stop()
		})
		list.AddItem(quitItem)
	}

	fourthIntervalOption = func() {
		list.Clear()

		tv.SetText("Select time interval.")
		form := cv.NewForm()

		switch newChallenge.TimeOption {
		case 0: //realtime
			//minute array
			m := []float64{0.25, 0.5, 0.75, 1, 1.5, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 25, 30, 35, 40, 45, 60, 75, 90, 105, 120, 135, 150, 165, 180}

			//second array
			s := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 25, 30, 35, 40, 45, 60, 75, 90, 105, 120, 135, 150, 165, 180}

			//slider 1: Minutes per side
			slider1 := cv.NewSlider()
			slider1.SetChangedFunc(func(value int) {
				Minutes = m[value]
				slider1.SetLabel(fmt.Sprintf("Minutes per side: %3v", m[value]))
			})
			slider1.SetMax(len(m) - 1)
			slider1.SetIncrement(1)
			// open at the current value, so what's shown is what's sent
			slider1.SetProgress(indexOf(m, Minutes))
			slider1.SetLabel(fmt.Sprintf("Minutes per side: %3v", Minutes))

			//slider 2:
			slider2 := cv.NewSlider()

			slider2.SetChangedFunc(func(value int) {
				Seconds = s[value]
				slider2.SetLabel(fmt.Sprintf("Increment in seconds: %3d", s[value]))
			})
			slider2.SetMax(len(s) - 1)
			slider2.SetIncrement(1)
			slider2.SetProgress(indexOf(s, Seconds))
			slider2.SetLabel(fmt.Sprintf("Increment in seconds: %3d", Seconds))

			form.AddFormItem(slider1)
			form.AddFormItem(slider2)
		case 1: //correspondence
			d := []int{1, 2, 3, 5, 7, 10, 14}
			//slider 1: Days
			slider1 := cv.NewSlider()
			slider1.SetChangedFunc(func(value int) {
				Days = d[value]
				slider1.SetLabel(fmt.Sprintf("Days per turn:   %3v", d[value]))
			})
			slider1.SetMax(len(d) - 1)
			slider1.SetIncrement(1)
			slider1.SetProgress(indexOf(d, Days))
			slider1.SetLabel(fmt.Sprintf("Days per turn:   %3v", Days))

			form.AddFormItem(slider1)

		}
		form.AddButton("Submit", func() {

			switch newChallenge.TimeOption {
			case 0:
				newChallenge.MinTurn = Minutes
				newChallenge.ClockLimit = fmt.Sprintf("%v", int(Minutes*60)) //minutes
				newChallenge.ClockIncrement = fmt.Sprintf("%v", Seconds)     //seconds
				path = append(path, fmt.Sprintf("%v minutes, %v seconds", Minutes, Seconds))

			case 1:
				newChallenge.Days = fmt.Sprintf("%v", Days) //days
				path = append(path, fmt.Sprintf("%v days per turn", Days))

			}

			updateTree(ratedOption)

			showList()
			fifthRatedOption()
		})
		form.AddButton("Back", func() {
			showList()
			if len(path) > 0 {
				path = path[:len(path)-1]
			}
			updateTree(timeOptions)
			timeThirdOption()
		})
		form.AddButton("Home", goHome)
		form.AddButton("Quit", Root.App.Stop)
		//form.SetBorder(true)
		form.SetTitle("Choose time interval option:")
		form.SetTitleAlign(cv.AlignCenter)

		showForm(form)
	}

	fifthRatedOption = func() {
		list.Clear()

		tv.SetText("Rated game?")
		for i := 0; i < len(ratedOption); i++ {
			item := cv.NewListItem(ratedOption[i])
			item.SetShortcut(rune('a' + i))
			item.SetSelectedFunc(func() {

				//store the choice
				if list.GetCurrentItemIndex() == 0 {
					newChallenge.Rated = "false"
					newChallenge.RatedBool = false

				} else {
					newChallenge.Rated = "true"
					newChallenge.RatedBool = true

				}

				path = append(path, ratedOption[list.GetCurrentItemIndex()])
				updateTree(colorOptions)

				//clear the list
				//add the new list
				sixthColorOption()
			})

			list.AddItem(item)
		}
		//add back
		item := cv.NewListItem("Back")
		item.SetSecondaryText("Back to previous selction")
		item.SetShortcut(rune('y'))
		item.SetSelectedFunc(func() {
			if len(path) > 0 {
				path = path[:len(path)-1]
			}
			if newChallenge.TimeOption < 2 {
				updateTree([]string{"Select interval."})
				fourthIntervalOption()
			} else {
				updateTree(timeOptions)
				timeThirdOption()
			}
		})
		list.AddItem(item)

		//add home
		item = cv.NewListItem("Home")
		item.SetSecondaryText("Back to Lichess Welcome")
		item.SetShortcut(rune('z'))
		item.SetSelectedFunc(goHome)
		list.AddItem(item)

		//add quit
		quitItem := cv.NewListItem("Quit")
		quitItem.SetSecondaryText("Press to exit")
		quitItem.SetShortcut('q')
		quitItem.SetSelectedFunc(func() {
			Root.App.Stop()
		})
		list.AddItem(quitItem)
	}

	sixthColorOption = func() {
		list.Clear()

		tv.SetText("Select your color.")
		for i := 0; i < len(colorOptions); i++ {
			item := cv.NewListItem(colorOptions[i])
			item.SetShortcut(rune('a' + i))
			item.SetSecondaryText(fmt.Sprintf("You will play %v.", colorOptions[i]))
			item.SetSelectedFunc(func() {
				//store the choice
				newChallenge.Color = colorOptions[list.GetCurrentItemIndex()]
				newChallenge.ColorIndex = list.GetCurrentItemIndex()

				path = append(path, colorOptions[list.GetCurrentItemIndex()])

				//clear the list
				//add the new list
				switch newChallenge.Type {
				case 0:
					eightSubmitOption()
				case 1:
					updateTree(Root.User.Friends)
					seventhfriendsOptions()
				case 2:
					updateTree([]string{"Select AI power level."})
					BotPowerLevelOption()
				}

			})

			list.AddItem(item)
		}
		//add back
		item := cv.NewListItem("Back")
		item.SetSecondaryText("Back to selecting the RATED type.")
		item.SetShortcut(rune('y'))
		item.SetSelectedFunc(func() {
			if len(path) > 0 {
				path = path[:len(path)-1]
			}
			updateTree(ratedOption)
			fifthRatedOption()
		})
		list.AddItem(item)

		//add home
		item = cv.NewListItem("Home")
		item.SetSecondaryText("Back to Lichess Welcome")
		item.SetShortcut(rune('z'))
		item.SetSelectedFunc(goHome)
		list.AddItem(item)

		//add quit
		quitItem := cv.NewListItem("Quit")
		quitItem.SetSecondaryText("Press to exit")
		quitItem.SetShortcut('q')
		quitItem.SetSelectedFunc(func() {
			Root.App.Stop()
		})
		list.AddItem(quitItem)
	}

	seventhfriendsOptions = func() {
		list.Clear()

		tv.SetText("Select friend to challenge.")
		for i := 0; i < len(Root.User.Friends); i++ {
			item := cv.NewListItem(Root.User.Friends[i])
			item.SetShortcut(rune('a' + i))
			item.SetSelectedFunc(func() {

				newChallenge.DestUser = Root.User.Friends[list.GetCurrentItemIndex()] //store the choice
				newChallenge.OpenEnded = false                                        //TODO: open ended

				path = append(path, Root.User.Friends[list.GetCurrentItemIndex()])
				updateTree([]string{})

				eightSubmitOption() //add the new list
			})

			list.AddItem(item)
		}
		//add back
		item := cv.NewListItem("Back")
		item.SetSecondaryText("Back to selecting the COLOR type.")
		item.SetShortcut(rune('y'))
		item.SetSelectedFunc(func() {
			if len(path) > 0 {
				path = path[:len(path)-1]
			}
			updateTree(colorOptions)
			sixthColorOption()
		})
		list.AddItem(item)

		//add home
		item = cv.NewListItem("Home")
		item.SetSecondaryText("Back to Lichess Welcome")
		item.SetShortcut(rune('z'))
		item.SetSelectedFunc(goHome)
		list.AddItem(item)

		//add quit
		quitItem := cv.NewListItem("Quit")
		quitItem.SetSecondaryText("Press to exit")
		quitItem.SetShortcut('q')
		quitItem.SetSelectedFunc(func() {
			Root.App.Stop()
		})
		list.AddItem(quitItem)
	}

	BotPowerLevelOption = func() {
		list.Clear()

		form := cv.NewForm()
		tv.SetText("Select bot power level.")
		p := []int{1, 2, 3, 4, 5, 6, 7, 8}
		slider1 := cv.NewSlider()
		slider1.SetChangedFunc(func(value int) {
			BotPowerLevel = p[value]
			slider1.SetLabel(fmt.Sprintf("AI Strength:   %3v", p[value]))
		})
		slider1.SetMax(len(p) - 1)
		slider1.SetIncrement(1)
		slider1.SetProgress(indexOf(p, BotPowerLevel))
		slider1.SetLabel(fmt.Sprintf("AI Strength:   %3v", BotPowerLevel))

		form.AddFormItem(slider1)

		form.AddButton("Submit", func() {
			newChallenge.Level = fmt.Sprintf("%v", BotPowerLevel)

			showList()

			path = append(path, fmt.Sprintf("AI Strength: %v", BotPowerLevel))
			updateTree([]string{})

			eightSubmitOption()
		})
		form.AddButton("Back", func() {
			showList()

			if len(path) > 0 {
				path = path[:len(path)-1]
			}
			updateTree(colorOptions)
			sixthColorOption()
		})
		form.AddButton("Home", goHome)
		form.AddButton("Quit", Root.App.Stop)
		//form.SetBorder(true)
		form.SetTitle("Choose bot power level:")
		form.SetTitleAlign(cv.AlignCenter)

		showForm(form)
	}

	eightSubmitOption = func() {
		list.Clear()
		tv.SetText("Review and submit your challenge.")
		//add
		submit := cv.NewListItem("challenge ok? submit")
		submit.SetSecondaryText("Submit your constructed challenge.")
		submit.SetShortcut(rune('s'))
		submit.SetSelectedFunc(gotoLoaderFromChallenge)
		list.AddItem(submit)

		//add back
		item := cv.NewListItem("Back")
		item.SetSecondaryText("Back to previous selection to send the challenge to")
		item.SetShortcut(rune('y'))
		item.SetSelectedFunc(func() {
			if len(path) > 0 {
				path = path[:len(path)-1]
			}
			switch newChallenge.Type {
			case 0:
				updateTree(colorOptions)
				sixthColorOption()
			case 1:
				updateTree(Root.User.Friends)
				seventhfriendsOptions()
			case 2:
				updateTree([]string{"Select AI power level."})
				BotPowerLevelOption()
			}
		})
		list.AddItem(item)

		//add home
		item = cv.NewListItem("Home")
		item.SetSecondaryText("Back to Lichess Welcome")
		item.SetShortcut(rune('z'))
		item.SetSelectedFunc(goHome)
		list.AddItem(item)

		//add quit
		quitItem := cv.NewListItem("Quit")
		quitItem.SetSecondaryText("Press to exit")
		quitItem.SetShortcut('q')
		quitItem.SetSelectedFunc(func() {
			Root.App.Stop()
		})
		list.AddItem(quitItem)
	}

	updateTree = func(new []string) {
		treeRoot.AddChild(cv.NewTreeNode(""))
		treeRoot.ClearChildren()
		var c *cv.TreeNode = treeRoot
		for _, n := range path {
			temp := cv.NewTreeNode(n)
			c.AddChild(temp)
			c = temp
		}
		for _, n := range new {
			c.AddChild(cv.NewTreeNode(n))
		}

	}

	firstOption()

	grid.SetColumns(-1, -1)
	fitMenuColumn(grid, list, []int{-1, -1}, 1)
	grid.SetRows(-1, -2, 1)

	grid.AddItem(list, 1, 1, 1, 1, 0, 0, true)
	grid.AddItem(tv, 0, 0, 1, 2, 0, 0, false)
	grid.AddItem(tree, 1, 0, 1, 1, 0, 0, false)
	grid.AddItem(Ribbon, 2, 0, 1, 2, 0, 0, false)
	registerMenus("challenge", false, list)

	return grid
}

// indexOf is the position of v in values (0 if it isn't there).
func indexOf[T comparable](values []T, v T) int {
	for i, x := range values {
		if x == v {
			return i
		}
	}
	return 0
}
