package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

type DraggableCard struct {
	widget.Card
	isDragging bool
	dragStart  fyne.Position
	dragOffset fyne.Position
	onDragEnd  func(*DraggableCard)
}

func NewDraggableCard(title, subtitle string, content fyne.CanvasObject, onDragEnd func(*DraggableCard)) *DraggableCard {
	card := &DraggableCard{onDragEnd: onDragEnd}
	card.ExtendBaseWidget(card)
	card.SetTitle(title)
	card.SetSubTitle(subtitle)
	card.SetContent(content)
	return card
}

func (d *DraggableCard) Dragged(ev *fyne.DragEvent) {
	if d.isDragging {
		d.Move(ev.Position.Subtract(d.dragOffset))
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
	d.dragStart = ev.Position
	d.dragOffset = ev.Position.Subtract(d.Position())
}

func (d *DraggableCard) MouseUp(*desktop.MouseEvent) {
	d.DragEnd()
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Movable Sticky Notes")

	columns := make([]*fyne.Container, 3)
	for i := 0; i < 3; i++ {
		columns[i] = container.NewVBox()
	}

	content := container.NewHBox(columns[0], columns[1], columns[2])

	onDragEnd := func(card *DraggableCard) {
		cardCenter := card.Position().Add(fyne.NewPos(card.Size().Width/2, card.Size().Height/2))
		for _, col := range columns {
			colPos := col.Position()
			if cardCenter.X >= colPos.X && cardCenter.X < colPos.X+col.Size().Width {
				// Remove card from its current column
				for _, c := range columns {
					c.Remove(card)
				}
				// Add card to the new column
				col.Add(card)
				break
			}
		}
		content.Refresh()
	}

	for i := 1; i <= 5; i++ {
		card := NewDraggableCard(fmt.Sprintf("Sticky note %d", i), "This is a sticky note", widget.NewLabel("Content"), onDragEnd)
		columns[i%3].Add(card)
	}

	// Add a canvas under the content to catch mouse events
	canvasObj := canvas.NewRectangle(color.RGBA{R: 173, G: 219, B: 156, A: 200})
	canvasContainer := container.NewMax(canvasObj, content)

	myWindow.SetContent(canvasContainer)
	myWindow.Resize(fyne.NewSize(600, 400))
	myWindow.ShowAndRun()
}
