# Product Manager

A desktop application for managing product information with file attachments, built with Go and WebView.

## Features

- **Product Management**: Add, edit, view, and delete products
- **File Attachments**: Support for multiple file types including photos, drawings, CAD files, CNC code, and invoices
- **Search & Filter**: Search products by part number, name, description, or material
- **Sorting**: Sort products by various fields in ascending or descending order
- **Export Data**: Export product data to Excel format
- **File Organization**: Automatic folder organization by part number
- **Cross-Platform**: Runs as a desktop application using WebView

## Product Information

Each product can store:
- Part Number (unique identifier)
- Part Name
- Description
- Cost
- Quantity
- Material specifications (type, size, cost)
- Finishing details (type, cost)
- File attachments in categorized folders:
    - Photos
    - 2D Drawings
    - 3D CAD Files
    - CNC Code
    - Invoices


## Installation

### Prerequisites

- Go 1.16 or higher
- GCC compiler (for SQLite3)

### Build Instructions

1. **Clone or download the source code**
2. **Install dependencies**:

```bash
go mod init product-manager
go get github.com/mattn/go-sqlite3
go get github.com/webview/webview_go
go get github.com/xuri/excelize/v2
```

3. **Create required directories**:

```bash
mkdir -p templates static
```

4. **Build the application**:

```bash
go build -o product-manager
```

5. **Run the application**:

```bash
./product-manager
```

## Project Structure

``` text
product-manager/
├── main.go                 # Main application file
├── templates/              # HTML templates
│   ├── index.html         # Product list view
│   ├── add.html           # Add product form
│   ├── modify.html        # Edit product form
│   └── detail.html        # Product detail view
├── static/                # Static assets (CSS, JS, images)
├── uploads/               # File upload directory
└── products.db           # SQLite database (auto-created)
```

## File Organization

Uploaded files are organized in the following structure:

text

```text
uploads/
└── [part_number]/
    ├── photos/
    ├── drawings/
    ├── cad/
    ├── cnc/
    └── invoice/
```

## API Endpoints

- `GET /` - Main product list
- `GET /api/products` - JSON API for products (supports pagination, search, sorting)
- `GET /add` - Add product form
- `POST /save` - Save new product
- `GET /modify/{id}` - Edit product form
- `POST /update` - Update product
- `GET /detail/{partNo}` - Product detail view
- `GET /delete/{id}` - Delete product
- `POST /remove-file` - Remove attached file
- `GET /export` - Export to Excel
- `POST /open-folder` - Open product folder

## Usage

### Adding a Product

1. Click "Add New Product" from the main screen
2. Fill in product details
3. Upload relevant files in their respective categories
4. Save the product

### Searching and Sorting

- Use the search box to find products by part number, name, description, or material
- Click column headers to sort by that field
- Toggle between ascending and descending order

### Managing Files

- Each product can have multiple files in different categories
- Files are automatically organized in part number folders
- Remove individual files without deleting the entire product
- Open the product's folder directly from the application

### Exporting Data

- Click "Export to Excel" to download all product data
- Export includes all product fields in a formatted Excel spreadsheet

## Database Schema

The application uses SQLite with the following table structure:

```sqllist
CREATE TABLE products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    partNo TEXT UNIQUE,
    partName TEXT,
    description TEXT,
    cost TEXT,
    qty INTEGER DEFAULT 0,
    material TEXT,
    material_size TEXT,
    material_cost TEXT,
    finishing_type TEXT,
    finishing_cost TEXT,
    photos TEXT,           -- JSON array of file info
    drawing_2d TEXT,       -- JSON array of file info
    cad_3d TEXT,          -- JSON array of file info
    cnc_code TEXT,        -- JSON array of file info
    invoice TEXT,         -- JSON array of file info
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## Technical Details

- **Backend**: Go with SQLite database
- **Frontend**: HTML templates with server-side rendering
- **UI Framework**: WebView for desktop application
- **File Handling**: Multi-part form uploads with automatic organization
- **Export**: Excel file generation using excelize library

## Troubleshooting

### Common Issues

1. **Database errors**: Ensure write permissions in the application directory
2. **File upload failures**: Check available disk space and directory permissions
3. **WebView initialization**: Make sure all dependencies are properly installed
    

### Logs

The application logs to stdout and can help diagnose issues with:

- Database operations
- File uploads
- HTTP requests
- Export operations

## License
This project is provided as-is for educational and personal use.