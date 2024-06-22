package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type TaskStatus int

const (
	NotStarted TaskStatus = iota + 1
	InProgress
	Done
)

func (s TaskStatus) String() string {
	return [...]string{"Not Started", "In Progress", "Done"}[s-1]
}

type DraggableCard struct {
	widget.Card
	isDragging   bool
	dragStartPos fyne.Position
	dragEndPos   fyne.Position
	onDragEnd    func(*DraggableCard)
	parent       fyne.CanvasObject
	contentLabel *widget.Label
	editButton   *widget.Button
	deleteButton *widget.Button
}

func showDeleteCardModal(card *DraggableCard) {
	window := fyne.CurrentApp().Driver().AllWindows()[0]
	deleteDialog := dialog.NewConfirm("Delete Card", "Are you sure you want to delete this card?", func(confirmed bool) {
		if !confirmed {
			return
		}
		for _, c := range columns {
			c.Remove(card)
		}
		content.Refresh()
	}, window)
	deleteDialog.Show()
}

func NewDraggableCard(title, subtitle, content string, onDragEnd func(*DraggableCard), parent fyne.CanvasObject) *DraggableCard {
	card := &DraggableCard{onDragEnd: onDragEnd, parent: parent}
	card.ExtendBaseWidget(card)
	card.SetTitle(title)
	card.SetSubTitle(subtitle)

	card.contentLabel = widget.NewLabel(content)
	card.contentLabel.Wrapping = fyne.TextWrapWord

	card.editButton = widget.NewButtonWithIcon("Edit", resourceEditPenSvgrepoComSvg, func() {
		showEditCardModal(card)
	})
	card.deleteButton = widget.NewButtonWithIcon("Delete", resourceTrashIconSvg, func() {
		showDeleteCardModal(card)
	})
	card.editButton.Importance = widget.HighImportance
	card.deleteButton.Importance = widget.DangerImportance

	// Use a grid layout with 2 columns to make buttons span full width
	buttons := container.New(layout.NewGridLayout(2),
		card.editButton,
		card.deleteButton,
	)

	cardContent := container.NewVBox(
		card.contentLabel,
		buttons,
	)
	card.SetContent(cardContent)

	return card
}

func showEditCardModal(card *DraggableCard) {
	window := fyne.CurrentApp().Driver().AllWindows()[0]

	titleEntry := widget.NewEntry()
	titleEntry.SetText(card.Title)
	titleEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("title cannot be empty")
		}
		return nil
	}

	subtitleEntry := widget.NewEntry()
	subtitleEntry.SetText(card.Subtitle)
	subtitleEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("subtitle cannot be empty")
		}
		return nil
	}

	contentEntry := widget.NewMultiLineEntry()
	contentEntry.SetText(card.contentLabel.Text)
	contentEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("content cannot be empty")
		}
		return nil
	}

	statusEntry := widget.NewSelect([]string{NotStarted.String(), InProgress.String(), Done.String()}, nil)
	foundCard := false
	for i, col := range columns {
		for _, obj := range col.Objects {
			if obj == card {
				statusEntry.SetSelectedIndex(i)
				foundCard = true
				break
			}
		}
		if foundCard {
			break
		}
	}
	if !foundCard {
		statusEntry.SetSelectedIndex(0)
	}

	items := []*widget.FormItem{
		{Text: "Title", Widget: titleEntry},
		{Text: "Subtitle", Widget: subtitleEntry},
		{Text: "Content", Widget: contentEntry},
		{Text: "Status", Widget: statusEntry},
	}
	submit := func(edit bool) {
		if !edit {
			return
		}
		card.SetTitle(titleEntry.Text)
		card.SetSubTitle(subtitleEntry.Text)
		card.contentLabel.SetText(contentEntry.Text)
		cardColumnIndex := -1
		for i, col := range columns {
			for _, obj := range col.Objects {
				if obj == card {
					cardColumnIndex = i
					break
				}
			}
			if cardColumnIndex != -1 {
				break
			}
		}

		card.Refresh()
		if cardColumnIndex != statusEntry.SelectedIndex() {
			for _, c := range columns {
				c.Remove(card)
			}

			card.Refresh()
			columns[statusEntry.SelectedIndex()].Add(card)
		}

		column := columns[statusEntry.SelectedIndex()]
		window.Canvas().Refresh(column)
	}

	form := dialog.NewForm("New Card", "Create", "Cancel", items, submit, window)
	form.Resize(fyne.NewSize(400, 400))
	form.Show()
}

func (d *DraggableCard) Dragged(ev *fyne.DragEvent) {
	if d.isDragging {
		deltaX := ev.Position.X - d.dragStartPos.X
		deltaY := ev.Position.Y - d.dragStartPos.Y
		d.Move(fyne.NewPos(d.Position().X+deltaX, d.Position().Y+deltaY))
		d.dragStartPos = ev.Position
	}
}

func (d *DraggableCard) DragEnd() {
	if d.isDragging {
		d.isDragging = false
		if d.onDragEnd != nil {
			d.onDragEnd(d)
		}
	}
}

func (d *DraggableCard) MouseDown(ev *desktop.MouseEvent) {
	d.isDragging = true
	d.dragStartPos = ev.Position
	if d.parent != nil {
		d.parent.Refresh()
	}
}

func (d *DraggableCard) MouseUp(*desktop.MouseEvent) {
	d.DragEnd()
}

func createColumnWithBorderAndHeader(content *fyne.Container, title string) *fyne.Container {
	border := canvas.NewRectangle(theme.BackgroundColor())
	border.StrokeWidth = 2
	border.StrokeColor = theme.ShadowColor()

	header := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	headerBg := canvas.NewRectangle(theme.PrimaryColor())
	headerContainer := container.NewStack(headerBg, header)

	scrollContainer := container.NewVScroll(content)
	scrollContainer.SetMinSize(fyne.NewSize(200, 0))

	return container.New(layout.NewBorderLayout(headerContainer, nil, nil, nil),
		border,
		headerContainer,
		scrollContainer,
	)
}

// Global variables
var columns []*fyne.Container
var borderedColumns []*fyne.Container
var content *fyne.Container

// Global onDragEnd function
func onDragEnd(card *DraggableCard) {
	cardCenter := card.Position().Add(fyne.NewPos(card.Size().Width/2, 0))
	for i, col := range borderedColumns {
		colPos := col.Position()
		if cardCenter.X >= colPos.X && cardCenter.X < colPos.X+col.Size().Width {
			// Remove card from its current column
			for _, c := range columns {
				c.Remove(card)
			}
			// Add card to the new column
			columns[i].Add(card)
			card.Move(fyne.NewPos(0, 0)) // Reset position within new column
			break
		}
	}
	content.Refresh()
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Movable Sticky Notes")

	columnTitles := []string{NotStarted.String(), InProgress.String(), Done.String()}
	columns = make([]*fyne.Container, 3)
	borderedColumns = make([]*fyne.Container, 3)

	content = container.New(layout.NewGridLayout(3))

	for i := 0; i < 3; i++ {
		columns[i] = container.NewVBox()
		borderedColumns[i] = createColumnWithBorderAndHeader(columns[i], columnTitles[i])
		borderedColumns[i].Resize(fyne.NewSize(200, 0)) // Set a fixed width for each column
		content.Add(borderedColumns[i])
	}

	for i := 1; i <= 5; i++ {
		card := NewDraggableCard(fmt.Sprintf("Sticky note %d", i), "This is a sticky note", "Content", onDragEnd, content)
		columns[i%3].Add(card)
	}

	// Create the plus button
	plusButton := widget.NewButton("+", func() {
		showNewCardModal(myWindow, columns) // Add new cards to the first column by default
	})

	// Create a container with the content and the plus button
	mainContainer := container.NewBorder(nil, container.NewCenter(plusButton), nil, nil, content)

	myWindow.SetContent(mainContainer)
	myWindow.Resize(fyne.NewSize(2000, 1200))
	myWindow.ShowAndRun()
}

func showNewCardModal(window fyne.Window, columns []*fyne.Container) {
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Enter card title")
	titleEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("title cannot be empty")
		}
		return nil
	}

	contentEntry := widget.NewMultiLineEntry()
	contentEntry.SetPlaceHolder("Enter card content")
	contentEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("content cannot be empty")
		}
		return nil
	}

	subtitleEntry := widget.NewEntry()
	subtitleEntry.SetPlaceHolder("Enter card subtitle")
	subtitleEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("subtitle cannot be empty")
		}
		return nil
	}

	statusEntry := widget.NewSelect([]string{NotStarted.String(), InProgress.String(), Done.String()}, nil)
	statusEntry.SetSelected(NotStarted.String())

	items := []*widget.FormItem{
		{Text: "Title", Widget: titleEntry},
		{Text: "Subtitle", Widget: subtitleEntry},
		{Text: "Content", Widget: contentEntry},
		{Text: "Status", Widget: statusEntry},
	}
	submit := func(create bool) {
		if !create {
			return
		}
		newCard := NewDraggableCard(titleEntry.Text, subtitleEntry.Text, contentEntry.Text, onDragEnd, content)
		column := columns[statusEntry.SelectedIndex()]
		column.Add(newCard)
		window.Canvas().Refresh(column)
	}

	form := dialog.NewForm("New Card", "Create", "Cancel", items, submit, window)
	form.Resize(fyne.NewSize(400, 400))
	form.Show()
}
