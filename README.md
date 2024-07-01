# Movable Sticky Notes

## Created by DonSlon1

Movable Sticky Notes is a Kanban-style task management application built with Go and the Fyne toolkit. It allows users to create, edit, and organize tasks across different status columns.

## Features

- Create, edit, and delete task cards
- Drag and drop cards between columns
- Three status columns: Not Started, In Progress, and Done
- Automatic saving and loading of tasks
- Unsaved changes warning before closing the application

## Prerequisites

Before you begin, ensure you have met the following requirements:

- Go 1.16 or later installed on your system
- Fyne toolkit and its dependencies

## Installation

1. Clone this repository:
   ```
   git clone https://github.com/DonSlon1/Go-To-Do-App.git
   ```

2. Navigate to the project directory:
   ```
   cd Go-To-Do-App
   ```

3. Install the required dependencies:
   ```
   go mod tidy
   ```

## Running the Application

To run the Movable Sticky Notes application, follow these steps:

1. Open a terminal and navigate to the project directory.

2. Run the following command:
   ```
   go run main.go
   ```

3. The application window should appear, and you can start creating and managing your tasks.

## Usage

- **Creating a New Card**: Click the "+" button at the bottom of the window to create a new task card.
- **Editing a Card**: Click the "Edit" button on a card to modify its details.
- **Deleting a Card**: Click the "Delete" button on a card to remove it.
- **Moving Cards**: Drag and drop cards between columns to change their status.
- **Saving Changes**: Click the "Save" button at the bottom of the window to manually save your changes.

## File Storage

The application automatically saves your tasks to a JSON file located at `saves/todos.json`. This file is created when you first save your tasks and is updated each time you make changes and save.

## Closing the Application

When you attempt to close the application with unsaved changes, a dialog will appear asking if you want to save your changes before exiting.

## Contributing

Contributions to the Movable Sticky Notes project are welcome. Please feel free to submit a Pull Request.

## License

This project is open source and available under the [MIT License](LICENSE).

## Contact

If you want to contact me, you can reach me at <lukindihel@gmail.com>.

## Acknowledgements

- [Fyne Toolkit](https://fyne.io/) for providing the GUI framework
- All contributors who have helped with the project
