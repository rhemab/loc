/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
	"golang.org/x/text/message"
)

type Ext struct {
	Files    int64
	Lines    int64
	Comments int64
	Blanks   int64
	Code     int64
}
type FileTypes[K comparable, V any] struct {
	m sync.Map
}

func (m *FileTypes[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

func (m *FileTypes[K, V]) Load(key K) (V, bool) {
	val, ok := m.m.Load(key)
	if !ok {
		var zero V
		return zero, false
	}
	return val.(V), true
}

var totalDirs, totalFiles, totalLines, totalCode, totalComments, totalBlanks int64
var exclude = []string{"node-modules", ".node", ".git", ".gitignore", "Makefile", ".DS_Store", "dist", ".toml", ".yml"}
var filter []string
var fileTypes FileTypes[string, *Ext]
var wg sync.WaitGroup

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "loc",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()

		filePath, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		input := getStdInput()
		if len(input) > 0 {
			filePath = strings.TrimSuffix(string(input), string('\n'))
		}
		if len(args) > 0 {
			filePath = args[0]
		}

		readDir(filePath)
		wg.Wait()

		p := message.NewPrinter(message.MatchLanguage("en"))
		printTable(p)
		fmt.Println("Searched", p.Sprint(totalDirs), "directories")
		fmt.Println("Took:", time.Since(start))
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.loc.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringSliceVarP(&filter, "filter", "f", []string{}, "Filter by file type: .go,.py,.lua,etc...")
}

func readFile(filePath string) {
	// filter by file extension
	ext := path.Ext(filePath)
	if ext == "" {
		return
	}
	if len(filter) == 0 && slices.Contains(exclude, ext) {
		return
	}
	if len(filter) > 0 && !slices.Contains(filter, ext) {
		return
	}

	atomic.AddInt64(&totalFiles, 1)

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	value, found := fileTypes.Load(ext)
	if !found {
		value = &Ext{Files: 1}
		fileTypes.Store(ext, value)
	} else {
		atomic.AddInt64(&value.Files, 1)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		atomic.AddInt64(&totalLines, 1)
		atomic.AddInt64(&value.Lines, 1)

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			atomic.AddInt64(&totalBlanks, 1)
			atomic.AddInt64(&value.Blanks, 1)
		} else if isComment(line) {
			atomic.AddInt64(&totalComments, 1)
			atomic.AddInt64(&value.Comments, 1)
		} else {
			atomic.AddInt64(&totalCode, 1)
			atomic.AddInt64(&value.Code, 1)
		}
	}
}

func readDir(filePath string) {
	files, err := os.ReadDir(filePath)
	if err != nil {
		readFile(filePath)
		return
	}

	atomic.AddInt64(&totalDirs, 1)

	for _, file := range files {
		// if no filter, exclude files
		if len(filter) == 0 && slices.Contains(exclude, file.Name()) {
			continue
		}

		newPath := path.Join(filePath, file.Name())

		if file.IsDir() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				readDir(newPath)
			}()
			continue
		}

		readFile(newPath)
	}
}

var commentPrefixes = []string{"//", "#", ";", "--"} // Go, Shell/Python, INI, SQL

func isComment(line string) bool {
	if line == "" {
		return false
	}
	for _, prefix := range commentPrefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

func getStdInput() []byte {
	if !isInputPiped() {
		return nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func isInputPiped() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func printTable(p *message.Printer) {
	rows := [][]string{}

	fileTypes.m.Range(func(key, value any) bool {
		row := []string{
			key.(string),
			p.Sprint(value.(*Ext).Files),
			p.Sprint(value.(*Ext).Lines),
			p.Sprint(value.(*Ext).Code),
			p.Sprint(value.(*Ext).Comments),
			p.Sprint(value.(*Ext).Blanks),
		}
		rows = append(rows, row)
		return true
	})

	var (
		purple    = lipgloss.Color("5")
		gray      = lipgloss.Color("9")
		lightGray = lipgloss.Color("12")

		headerStyle  = lipgloss.NewStyle().Foreground(purple).Bold(true).Align(lipgloss.Center)
		cellStyle    = lipgloss.NewStyle().Padding(0, 1).Width(14)
		oddRowStyle  = cellStyle.Foreground(gray)
		evenRowStyle = cellStyle.Foreground(lightGray)
	)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(purple)).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == table.HeaderRow:
				return headerStyle
			case row%2 == 0:
				return evenRowStyle
			default:
				return oddRowStyle
			}
		}).
		Headers("File Type", "Files", "Lines", "Code", "Comments", "Blanks").
		Rows(rows...)

	t.Row()
	t.Row()
	t.Row("Total", p.Sprint(totalFiles), p.Sprint(totalLines), p.Sprint(totalCode), p.Sprint(totalComments), p.Sprint(totalBlanks))

	fmt.Println(t)
}
