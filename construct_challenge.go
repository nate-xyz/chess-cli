package main

import (
	"fmt"

	cv "code.rocketnine.space/tslocum/cview"
)

func initConstruct() *cv.Grid {
	grid := cv.NewGrid()
	grid.SetColumns(-1)
	grid.SetRows(-1)
	grid.SetBorders(false)
	list := cv.NewList()
	list.SetWrapAround(true)
	//selection := []int{}

	challengeTypeOption := []string{"Random", "Friend", "AI"}
	challengeTypeOptionExplain := []string{"Seek a random player.", "Challenge a friend.", "Play against lichess bot."}
	variantOption := []string{"standard", "chess960", "crazyhouse", "antichess", "atomic", "horde", "kingOfTheHill", "racingKings", "threeCheck"}
	timeOptions := []string{"real time", "correspondence", "unlimited"}
	ratedOption := []string{"casual", "rated"}
	colorOptions := []string{"random", "white", "black"}

	//title_array := []string{"options", "variants", "time options", "time interval", "rated/casual", "choose color", "select friend to challenge", "submit challenge"}

	var firstOption func()
	var variantSecondOption func()
	var timeThirdOption func()
	var fourthIntervalOption func()
	var fifthRatedOption func()
	var sixthColorOption func()
	var seventhfriendsOptions func()
	var eightSubmitOption func()
	var Minutes float64 = 0.25
	var Seconds int
	var Days int = 1
	//choose the type of challenge
	firstOption = func() {
		list.Clear()
		for i := 0; i < len(challengeTypeOption); i++ {
			item := cv.NewListItem(challengeTypeOption[i])
			item.SetSecondaryText(challengeTypeOptionExplain[i])
			item.SetShortcut(rune('a' + i))
			item.SetSelectedFunc(func() {
				//store the choice
				newChallenge.Type = list.GetCurrentItemIndex()
				//clear the list
				//add the new list
				variantSecondOption()
			})
			list.AddItem(item)
		}
		//add back
		item := cv.NewListItem("Back")
		item.SetSecondaryText("Back to Lichess Welcome")
		item.SetShortcut(rune('z'))
		item.SetSelectedFunc(gotoLichessAfterLogin)
		list.AddItem(item)

		//add quit
		quitItem := cv.NewListItem("Quit")
		quitItem.SetSecondaryText("Press to exit")
		quitItem.SetShortcut('q')
		quitItem.SetSelectedFunc(func() {
			root.app.Stop()
		})
		list.AddItem(quitItem)

	}

	variantSecondOption = func() {
		list.Clear()
		for i := 0; i < len(variantOption); i++ {
			item := cv.NewListItem(variantOption[i])
			item.SetShortcut(rune('a' + i))
			item.SetSelectedFunc(func() {
				//store the choice
				newChallenge.Variant = variantOption[list.GetCurrentItemIndex()]
				//clear the list
				//add the new list
				timeThirdOption()
			})
			list.AddItem(item)
		}
		//add back
		item := cv.NewListItem("Back")
		item.SetSecondaryText("Back to selecting the CHALLENGE type.")
		item.SetShortcut(rune('y'))
		item.SetSelectedFunc(firstOption)
		list.AddItem(item)

		//add back
		item = cv.NewListItem("Home")
		item.SetSecondaryText("Back to Lichess Welcome")
		item.SetShortcut(rune('z'))
		item.SetSelectedFunc(gotoLichessAfterLogin)
		list.AddItem(item)

		//add quit
		quitItem := cv.NewListItem("Quit")
		quitItem.SetSecondaryText("Press to exit")
		quitItem.SetShortcut('q')
		quitItem.SetSelectedFunc(func() {
			root.app.Stop()
		})
		list.AddItem(quitItem)

	}

	timeThirdOption = func() {
		list.Clear()
		for i := 0; i < len(timeOptions); i++ {
			item := cv.NewListItem(timeOptions[i])
			item.SetShortcut(rune('a' + i))
			item.SetSelectedFunc(func() {
				//store the choice
				newChallenge.TimeOption = list.GetCurrentItemIndex()
				//clear the list
				//add the new list
				if newChallenge.TimeOption < 2 {
					fourthIntervalOption()
				} else {
					fifthRatedOption()
				}

			})
			list.AddItem(item)
		}
		//add back
		item := cv.NewListItem("Back")
		item.SetSecondaryText("Back to selecting the VARIANT type.")
		item.SetShortcut(rune('y'))
		item.SetSelectedFunc(variantSecondOption)
		list.AddItem(item)

		//add back
		item = cv.NewListItem("Home")
		item.SetSecondaryText("Back to Lichess Welcome")
		item.SetShortcut(rune('z'))
		item.SetSelectedFunc(gotoLichessAfterLogin)
		list.AddItem(item)

		//add quit
		quitItem := cv.NewListItem("Quit")
		quitItem.SetSecondaryText("Press to exit")
		quitItem.SetShortcut('q')
		quitItem.SetSelectedFunc(func() {
			root.app.Stop()
		})
		list.AddItem(quitItem)
	}

	fourthIntervalOption = func() {
		list.Clear()
		grid.Clear()
		form := cv.NewForm()

		switch newChallenge.TimeOption {
		case 0: //realtime
			//minute array
			m := []float64{0.25, 0.5, 0.75, 1.0, 1.5}
			lof, hif := 2, 20
			m_ := make([]float64, hif-lof+1)
			for i := range m_ {
				m_[i] = float64(i + lof)
			}
			m = append(m, m_...)
			lof, hif = 25, 45
			m_ = make([]float64, ((hif-lof+1)/5)+1)
			for i := range m_ {
				m_[i] = float64(i*5 + lof)
			}
			m = append(m, m_...)
			lof, hif = 60, 180
			m_ = make([]float64, ((hif-lof+1)/15)+1)
			for i := range m_ {
				m_[i] = float64(i*15 + lof)
			}
			m = append(m, m_...)

			//second array
			s := []int{}
			lof, hif = 0, 20
			s_ := make([]int, hif-lof+1)
			for i := range s_ {
				s_[i] = int(i + lof)
			}
			s = append(s, s_...)
			lof, hif = 25, 45
			s_ = make([]int, ((hif-lof+1)/5)+1)
			for i := range s_ {
				s_[i] = int(i*5 + lof)
			}
			s = append(s, s_...)
			lof, hif = 60, 180
			s_ = make([]int, ((hif-lof+1)/15)+1)
			for i := range m_ {
				s_[i] = int(i*15 + lof)
			}
			s = append(s, s_...)

			//slider 1: Minutes per side
			slider1 := cv.NewSlider()
			slider1.SetLabel("Minutes per side:   0.25")
			slider1.SetChangedFunc(func(value int) {
				Minutes = m[value]
				slider1.SetLabel(fmt.Sprintf("Minutes per side: %3v", m[value]))
			})
			slider1.SetMax(len(m) - 1)
			slider1.SetIncrement(1)

			//slider 2:
			slider2 := cv.NewSlider()
			slider2.SetLabel("Increment in seconds:   0")

			slider2.SetChangedFunc(func(value int) {
				Seconds = s[value]
				slider2.SetLabel(fmt.Sprintf("Increment in seconds: %3d", s[value]))
			})
			slider2.SetMax(len(s) - 1)
			slider2.SetIncrement(1)

			form.AddFormItem(slider1)
			form.AddFormItem(slider2)
		case 1: //correspondence
			d := []int{1, 2, 3, 5, 7, 10, 14}
			//slider 1: Days
			slider1 := cv.NewSlider()
			slider1.SetLabel("Days per turn:   1")
			slider1.SetChangedFunc(func(value int) {
				Days = d[value]
				slider1.SetLabel(fmt.Sprintf("Days per turn:   %3v", d[value]))
			})
			slider1.SetMax(len(d) - 1)
			slider1.SetIncrement(1)

			form.AddFormItem(slider1)

		}
		form.AddButton("Submit", func() {

			switch newChallenge.TimeOption {
			case 0:
				newChallenge.MinTurn = Minutes
				newChallenge.ClockLimit = fmt.Sprintf("%v", int(Minutes*60)) //minutes
				newChallenge.ClockIncrement = fmt.Sprintf("%v", Seconds)     //seconds
			case 1:
				newChallenge.Days = fmt.Sprintf("%v", Days) //days
			}

			grid.Clear()
			grid.AddItem(list, 0, 0, 1, 1, 0, 0, true)
			fifthRatedOption()
		})
		form.AddButton("Back", func() {
			grid.Clear()
			grid.AddItem(list, 0, 0, 1, 1, 0, 0, true)
			timeThirdOption()
		})
		form.AddButton("Home", gotoLichessAfterLogin)
		form.AddButton("Quit", root.app.Stop)
		//form.SetBorder(true)
		form.SetTitle("Choose time interval option:")
		form.SetTitleAlign(cv.AlignCenter)

		grid.AddItem(form, 0, 0, 1, 1, 0, 0, true)
	}

	fifthRatedOption = func() {
		list.Clear()
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
			if newChallenge.TimeOption < 2 {
				fourthIntervalOption()
			} else {
				timeThirdOption()
			}
		})
		list.AddItem(item)

		//add back
		item = cv.NewListItem("Home")
		item.SetSecondaryText("Back to Lichess Welcome")
		item.SetShortcut(rune('z'))
		item.SetSelectedFunc(gotoLichessAfterLogin)
		list.AddItem(item)

		//add quit
		quitItem := cv.NewListItem("Quit")
		quitItem.SetSecondaryText("Press to exit")
		quitItem.SetShortcut('q')
		quitItem.SetSelectedFunc(func() {
			root.app.Stop()
		})
		list.AddItem(quitItem)
	}

	sixthColorOption = func() {
		list.Clear()
		for i := 0; i < len(colorOptions); i++ {
			item := cv.NewListItem(colorOptions[i])
			item.SetShortcut(rune('a' + i))
			item.SetSecondaryText(fmt.Sprintf("You will play %v.", colorOptions[i]))
			item.SetSelectedFunc(func() {
				//store the choice
				newChallenge.Color = colorOptions[list.GetCurrentItemIndex()]
				newChallenge.ColorIndex = list.GetCurrentItemIndex()
				//clear the list
				//add the new list
				seventhfriendsOptions()
			})
			list.AddItem(item)
		}
		//add back
		item := cv.NewListItem("Back")
		item.SetSecondaryText("Back to selecting the RATED type.")
		item.SetShortcut(rune('y'))
		item.SetSelectedFunc(fifthRatedOption)
		list.AddItem(item)

		//add back
		item = cv.NewListItem("Home")
		item.SetSecondaryText("Back to Lichess Welcome")
		item.SetShortcut(rune('z'))
		item.SetSelectedFunc(gotoLichessAfterLogin)
		list.AddItem(item)

		//add quit
		quitItem := cv.NewListItem("Quit")
		quitItem.SetSecondaryText("Press to exit")
		quitItem.SetShortcut('q')
		quitItem.SetSelectedFunc(func() {
			root.app.Stop()
		})
		list.AddItem(quitItem)
	}

	seventhfriendsOptions = func() {
		list.Clear()
		for i := 0; i < len(allFriends); i++ {
			item := cv.NewListItem(allFriends[i])
			item.SetShortcut(rune('a' + i))
			item.SetSelectedFunc(func() {
				//store the choice
				newChallenge.DestUser = allFriends[list.GetCurrentItemIndex()]
				newChallenge.OpenEnded = false //TODO: open ended
				//clear the list
				//add the new list
				eightSubmitOption()
			})
			list.AddItem(item)
		}
		//add back
		item := cv.NewListItem("Back")
		item.SetSecondaryText("Back to selecting the COLOR type.")
		item.SetShortcut(rune('y'))
		item.SetSelectedFunc(sixthColorOption)
		list.AddItem(item)

		//add back
		item = cv.NewListItem("Home")
		item.SetSecondaryText("Back to Lichess Welcome")
		item.SetShortcut(rune('z'))
		item.SetSelectedFunc(gotoLichessAfterLogin)
		list.AddItem(item)

		//add quit
		quitItem := cv.NewListItem("Quit")
		quitItem.SetSecondaryText("Press to exit")
		quitItem.SetShortcut('q')
		quitItem.SetSelectedFunc(func() {
			root.app.Stop()
		})
		list.AddItem(quitItem)
	}

	eightSubmitOption = func() {
		list.Clear()

		//add
		submit := cv.NewListItem("challenge ok? submit")
		submit.SetSecondaryText("Submit your constructed challenge.")
		submit.SetShortcut(rune('s'))
		submit.SetSelectedFunc(gotoLoaderFromChallenge)
		list.AddItem(submit)

		//add back
		item := cv.NewListItem("Back")
		item.SetSecondaryText("Back to selecting the FRIEND to send the challenge to")
		item.SetShortcut(rune('y'))
		item.SetSelectedFunc(seventhfriendsOptions)
		list.AddItem(item)

		//add back
		item = cv.NewListItem("Home")
		item.SetSecondaryText("Back to Lichess Welcome")
		item.SetShortcut(rune('z'))
		item.SetSelectedFunc(gotoLichessAfterLogin)
		list.AddItem(item)

		//add quit
		quitItem := cv.NewListItem("Quit")
		quitItem.SetSecondaryText("Press to exit")
		quitItem.SetShortcut('q')
		quitItem.SetSelectedFunc(func() {
			root.app.Stop()
		})
		list.AddItem(quitItem)
	}

	firstOption()

	grid.AddItem(list, 0, 0, 1, 1, 0, 0, true)
	return grid
}
