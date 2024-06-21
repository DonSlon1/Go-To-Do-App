package main

import (
	"fmt"
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type customButton struct {
	widget.Button
	bgColor color.Color
}

func newCustomButton(label string, tapped func()) *customButton {
	button := &customButton{}
	button.ExtendBaseWidget(button)
	button.Text = label
	button.OnTapped = tapped
	button.bgColor = color.NRGBA{R: 0x80, G: 0, B: 0, A: 0xff}
	return button
}

func (b *customButton) CreateRenderer() fyne.WidgetRenderer {
	b.ExtendBaseWidget(b)
	bg := canvas.NewRectangle(b.bgColor)
	text := canvas.NewText(b.Text, color.White)
	text.Alignment = fyne.TextAlignCenter
	objects := []fyne.CanvasObject{bg, text}

	return &customButtonRenderer{
		button:  b,
		bg:      bg,
		text:    text,
		objects: objects,
	}
}

type customButtonRenderer struct {
	button  *customButton
	bg      *canvas.Rectangle
	text    *canvas.Text
	objects []fyne.CanvasObject
}

func (r *customButtonRenderer) Destroy() {}

func (r *customButtonRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.text.Resize(size)
}

func (r *customButtonRenderer) MinSize() fyne.Size {
	return r.text.MinSize().Add(fyne.NewSize(20, 20))
}

func (r *customButtonRenderer) Refresh() {
	r.bg.FillColor = r.button.bgColor
	r.text.Text = r.button.Text
	r.bg.Refresh()
	r.text.Refresh()
}

func (r *customButtonRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Input and Custom Widget Example")

	// Text input
	textEntry := widget.NewEntry()
	textEntry.SetPlaceHolder("Enter text here")

	// Number input
	numberEntry := widget.NewEntry()
	numberEntry.SetPlaceHolder("Enter a number")

	// Password input
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Enter password")

	// Dropdown
	dropdown := widget.NewSelect([]string{"Option 1", "Option 2", "Option 3"}, func(value string) {
		fmt.Println("Selected:", value)
	})

	// Custom button
	customBtn := newCustomButton("Submit", func() {
		// Get text input
		text := textEntry.Text
		fmt.Println("Text input:", text)

		// Get and validate number input
		numText := numberEntry.Text
		if num, err := strconv.Atoi(numText); err == nil {
			fmt.Println("Number input:", num)
		} else {
			fmt.Println("Invalid number input")
		}

		// Get password input
		password := passwordEntry.Text
		fmt.Println("Password input:", password)

		// Get dropdown selection
		fmt.Println("Dropdown selection:", dropdown.Selected)
	})

	content := container.NewVBox(
		widget.NewLabel("Text Input:"),
		textEntry,
		widget.NewLabel("Number Input:"),
		numberEntry,
		widget.NewLabel("Password Input:"),
		passwordEntry,
		widget.NewLabel("Dropdown:"),
		dropdown,
		customBtn,
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(300, 400))
	myWindow.ShowAndRun()
}
