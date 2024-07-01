package main

import (
	"encoding/json"
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
	"log"
	"os"
)

type TaskStatus int

var unsavedChanges bool

type Todo struct {
	Title    string     `json:"title"`
	Subtitle string     `json:"subtitle"`
	Content  string     `json:"content"`
	Status   TaskStatus `json:"status"`
}

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
	onDragEnd    func(*DraggableCard, fyne.Position)
	parent       fyne.CanvasObject
	contentLabel *widget.Label
	editButton   *widget.Button
	deleteButton *widget.Button
}

func loadTodos() error {
	data, err := os.ReadFile("saves/todos.json")
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, which is fine for first run
		}
		return err
	}

	var todos []Todo
	err = json.Unmarshal(data, &todos)
	if err != nil {
		return err
	}

	for _, todo := range todos {
		card := NewDraggableCard(todo.Title, todo.Subtitle, todo.Content, onDragEnd, content)
		columns[todo.Status-1].Add(card)
	}

	return nil
}

func saveTodos() error {
	var todos []Todo

	for i, column := range columns {
		for _, obj := range column.Objects {
			if card, ok := obj.(*DraggableCard); ok {
				todo := Todo{
					Title:    card.Title,
					Subtitle: card.Subtitle,
					Content:  card.contentLabel.Text,
					Status:   TaskStatus(i + 1),
				}
				todos = append(todos, todo)
			}
		}
	}

	jsonData, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile("saves/todos.json", jsonData, 0644)
	if err != nil {
		return err
	}

	unsavedChanges = false
	return nil
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

func NewDraggableCard(title, subtitle, content string, onDragEnd func(*DraggableCard, fyne.Position), parent fyne.CanvasObject) *DraggableCard {
	card := &DraggableCard{onDragEnd: onDragEnd, parent: parent}
	card.ExtendBaseWidget(card)
	card.SetTitle(title)
	card.SetSubTitle(subtitle)

	card.contentLabel = widget.NewLabel(content)
	card.contentLabel.Wrapping = fyne.TextWrapWord

	card.editButton = widget.NewButtonWithIcon("Edit", resourceEditPenSvgrepoComSvg, func() {
		showEditCardModal(card)
		unsavedChanges = true
	})
	card.deleteButton = widget.NewButtonWithIcon("Delete", resourceTrashIconSvg, func() {
		showDeleteCardModal(card)
		unsavedChanges = true
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

	form := dialog.NewForm("Edit Card", "Edit", "Cancel", items, submit, window)
	form.Resize(fyne.NewSize(400, 400))
	form.Show()
}

func (d *DraggableCard) Dragged(ev *fyne.DragEvent) {
	if d.isDragging {
		d.Move(d.Position().Add(ev.Dragged))
		d.dragEndPos = ev.AbsolutePosition
	}
}

func (d *DraggableCard) DragEnd() {
	if d.isDragging {
		d.isDragging = false
		if d.onDragEnd != nil {
			d.onDragEnd(d, d.dragEndPos)
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

func (d *DraggableCard) MouseUp(ev *desktop.MouseEvent) {
	d.dragEndPos = ev.AbsolutePosition
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
func onDragEnd(card *DraggableCard, p fyne.Position) {
	cardCenter := card.Position().Add(fyne.NewPos(card.Size().Width/2, 0))
	log.Print(cardCenter)
	for i, col := range borderedColumns {
		colPos := col.Position()
		xCardCenter := p.X
		if xCardCenter < 0 {
			xCardCenter = 0
		} else if xCardCenter > borderedColumns[2].Position().X+borderedColumns[2].Size().Width {
			xCardCenter = borderedColumns[2].Position().X + borderedColumns[2].Size().Width - 1
		}
		if xCardCenter >= colPos.X && xCardCenter < colPos.X+col.Size().Width {
			// Remove card from its current column
			for _, c := range columns {
				c.Remove(card)
			}
			// Add card to the new column
			columns[i].Add(card)
			card.Move(fyne.NewPos(0, 0)) // Reset position within new column
			card.Refresh()
			columns[i].Refresh()
			break
		}
	}
	unsavedChanges = true
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

	// Load todos at startup
	err := loadTodos()
	if err != nil {
		log.Printf("Error loading todos: %v", err)
		dialog.ShowError(err, myWindow)
	}

	// Create the plus button
	plusButton := widget.NewButton("+", func() {
		showNewCardModal(myWindow, columns)
	})

	// Create the save button
	saveButton := widget.NewButton("Save", func() {
		err := saveTodos()
		if err != nil {
			dialog.ShowError(err, myWindow)
		} else {
			dialog.ShowInformation("Success", "Todos saved successfully", myWindow)
			unsavedChanges = false
		}
	})

	// Create a container with the content, plus button, and save button
	buttonContainer := container.NewHBox(plusButton, saveButton)
	mainContainer := container.NewBorder(nil, container.NewCenter(buttonContainer), nil, nil, content)

	myWindow.SetContent(mainContainer)
	myWindow.Resize(fyne.NewSize(2000, 1200))
	myWindow.SetCloseIntercept(func() {
		if unsavedChanges {
			showUnsavedChangesDialog(myWindow)
		} else {
			myWindow.Close()
		}
	})

	myWindow.ShowAndRun()
}

func showUnsavedChangesDialog(window fyne.Window) {
	dialog.ShowCustomConfirm("Unsaved Changes", "Save", "Don't Save", widget.NewLabel("You have unsaved changes. Do you want to save before closing?"),
		func(save bool) {
			if save {
				err := saveTodos()
				if err != nil {
					dialog.ShowError(err, window)
					return
				}
				dialog.ShowInformation("Success", "Todos saved successfully", window)
			}
			window.Close()
		}, window)
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
		unsavedChanges = true
		window.Canvas().Refresh(column)
	}

	form := dialog.NewForm("New Card", "Create", "Cancel", items, submit, window)
	form.Resize(fyne.NewSize(400, 400))
	form.Show()
}
