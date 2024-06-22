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

var (
	board   *KanbanBoard
	content *fyne.Container
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
	parent       *Column
	contentLabel *widget.Label
	editButton   *widget.Button
	deleteButton *widget.Button
}

type KanbanBoard struct {
	Columns []*Column
}

type Column struct {
	Title string
	Cards []*DraggableCard
}

type KanbanLayout struct {
	board *KanbanBoard
}

func (k *KanbanLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	columnWidth := size.Width / float32(len(k.board.Columns))
	for i, col := range k.board.Columns {
		if i >= len(objects) {
			break // Don't process more columns than we have objects
		}
		colObj := objects[i].(*fyne.Container)
		colObj.Resize(fyne.NewSize(columnWidth, size.Height))
		colObj.Move(fyne.NewPos(float32(i)*columnWidth, 0))

		cardY := float32(60) // Leave space for column title
		for _, card := range col.Cards {
			card.Resize(fyne.NewSize(columnWidth-20, 80))
			card.Move(fyne.NewPos(10, cardY))
			cardY += 90
		}
	}
}

func (k *KanbanLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(600, 400)
}

func findColumnIndex(columns []*Column, targetCol *Column) int {
	for i, col := range columns {
		if col == targetCol {
			return i
		}
	}
	return -1
}

func onDragEnd(card *DraggableCard) {
	cardCenter := card.Position().Add(fyne.NewPos(card.Size().Width/2, 0))
	for i := range board.Columns {
		colObj := content.Objects[i].(*fyne.Container)
		colPos := colObj.Position()
		if cardCenter.X >= colPos.X && cardCenter.X < colPos.X+colObj.Size().Width {
			// Remove card from its current column
			if card.parent != nil {
				for j, c := range card.parent.Cards {
					if c == card {
						card.parent.Cards = append(card.parent.Cards[:j], card.parent.Cards[j+1:]...)
						break
					}
				}
				// Remove from UI
				oldColumnIndex := findColumnIndex(board.Columns, card.parent)
				if oldColumnIndex != -1 {
					colContent := content.Objects[oldColumnIndex].(*fyne.Container).Objects[0].(*fyne.Container)
					colContent.Remove(card)
				}
			}
			// Add card to the new column
			board.Columns[i].Cards = append(board.Columns[i].Cards, card)
			card.parent = board.Columns[i]
			// Add to UI
			colContent := colObj.Objects[0].(*fyne.Container)
			colContent.Add(card)
			break
		}
	}
	content.Refresh()
}

func showDeleteCardModal(card *DraggableCard) {
	window := fyne.CurrentApp().Driver().AllWindows()[0]
	deleteDialog := dialog.NewConfirm("Delete Card", "Are you sure you want to delete this card?", func(confirmed bool) {
		if !confirmed {
			return
		}
		if card.parent != nil {
			for i, c := range card.parent.Cards {
				if c == card {
					card.parent.Cards = append(card.parent.Cards[:i], card.parent.Cards[i+1:]...)
					break
				}
			}
		}
		content.Refresh()
	}, window)
	deleteDialog.Show()
}

func NewDraggableCard(title, subtitle, content string, onDragEnd func(*DraggableCard), parent *Column) *DraggableCard {
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
	for i, col := range board.Columns {
		if col == card.parent {
			statusEntry.SetSelectedIndex(i)
			break
		}
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

		newColumnIndex := statusEntry.SelectedIndex()
		if card.parent != board.Columns[newColumnIndex] {
			for i, c := range card.parent.Cards {
				if c == card {
					card.parent.Cards = append(card.parent.Cards[:i], card.parent.Cards[i+1:]...)
					break
				}
			}
			board.Columns[newColumnIndex].Cards = append(board.Columns[newColumnIndex].Cards, card)
			card.parent = board.Columns[newColumnIndex]
		}

		card.Refresh()
		content.Refresh()
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
	content.Refresh() // Refresh the entire content container
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

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Movable Sticky Notes")

	board = &KanbanBoard{
		Columns: []*Column{
			{Title: NotStarted.String(), Cards: []*DraggableCard{}},
			{Title: InProgress.String(), Cards: []*DraggableCard{}},
			{Title: Done.String(), Cards: []*DraggableCard{}},
		},
	}

	content = container.New(&KanbanLayout{board: board})

	// Create column containers and add them to the content
	for _, col := range board.Columns {
		colContent := container.NewVBox()
		borderedCol := createColumnWithBorderAndHeader(colContent, col.Title)
		content.Add(borderedCol)
	}

	// Add sample cards
	for i := 1; i <= 5; i++ {
		card := NewDraggableCard(fmt.Sprintf("Sticky note %d", i), "This is a sticky note", "Content", onDragEnd, board.Columns[i%3])
		board.Columns[i%3].Cards = append(board.Columns[i%3].Cards, card)
		colContent := content.Objects[i%3].(*fyne.Container).Objects[0].(*fyne.Container)
		colContent.Add(card)
	}

	// Create the plus button
	plusButton := widget.NewButton("+", func() {
		showNewCardModal(myWindow)
	})

	// Create a container with the content and the plus button
	mainContainer := container.NewBorder(nil, container.NewCenter(plusButton), nil, nil, content)

	myWindow.SetContent(mainContainer)
	myWindow.Resize(fyne.NewSize(2000, 1200))
	myWindow.ShowAndRun()
}

func showNewCardModal(window fyne.Window) {
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
		newCard := NewDraggableCard(titleEntry.Text, subtitleEntry.Text, contentEntry.Text, onDragEnd, board.Columns[statusEntry.SelectedIndex()])
		board.Columns[statusEntry.SelectedIndex()].Cards = append(board.Columns[statusEntry.SelectedIndex()].Cards, newCard)
		content.Refresh()
	}

	form := dialog.NewForm("New Card", "Create", "Cancel", items, submit, window)
	form.Resize(fyne.NewSize(400, 400))
	form.Show()
}
