# LOC: Lines of Code Counter

LOC is a command-line tool for counting lines of code, blank lines, and comments in a given directory. It recursively scans directories and provides a breakdown of different file types.

## Features
- **Recursive Scanning**: Automatically traverses directories and subdirectories to analyze files.
- **File Type Filtering**: Allows filtering by file extensions (e.g., `.go`, `.py`, `.js`).
- **Exclusion Rules**: Ignores common directories and files such as `.git`, `node_modules`, and `Makefile`.
- **Detailed Statistics**: Reports total files, lines of code, comments, and blank lines per file type.
- **Multi-threaded Performance**: Uses goroutines for efficient scanning.
- **Pretty Table Output**: Displays results in a well-formatted table using `lipgloss`.

## Installation

To install LOC, use the following command:

```sh
go install github.com/yourusername/loc@latest
```

## Usage

Run the command in the desired directory:

```sh
loc
```

Or specify a target directory:

```sh
loc /path/to/directory
```

### Filtering by File Type

To count only specific file types:

```sh
loc --filter .go,.py
```

### Example Output
```
+-----------+-------+--------+--------+----------+--------+
| File Type | Files | Lines  | Code   | Comments | Blanks |
+-----------+-------+--------+--------+----------+--------+
| .go       |   10  |   500  |   400  |     50   |    50  |
| .py       |    5  |   200  |   150  |     30   |    20  |
+-----------+-------+--------+--------+----------+--------+
| Total     |   15  |   700  |   550  |     80   |    70  |
+-----------+-------+--------+--------+----------+--------+
Searched 5 directories
Took: 150ms
```

## Excluded Files and Directories
By default, LOC excludes the following:
- `node_modules`
- `.git`
- `.gitignore`
- `Makefile`
- `.DS_Store`
- `dist`
- `.toml`, `.yml`

## Contributing
Contributions are welcome! Feel free to submit issues and pull requests.
