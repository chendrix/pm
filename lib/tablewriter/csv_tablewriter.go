package tablewriter

import (
	"encoding/csv"
	"io"
)

type CSVTableWriter struct {
	io.Writer

	header []string
	footer []string
	rows   [][]string
}

func NewCSVTableWriter(w io.Writer) *CSVTableWriter {
	return &CSVTableWriter{
		Writer: w,
	}
}

func (c *CSVTableWriter) SetHeader(keys []string) {
	c.header = keys
}

func (c *CSVTableWriter) SetFooter(keys []string) {
	c.footer = keys
}

func (c *CSVTableWriter) Append(row []string) {
	c.rows = append(c.rows, row)
}

func (c *CSVTableWriter) Render() error {
	w := csv.NewWriter(c.Writer)

	var err error
	if c.header != nil {
		err = w.Write(c.header)
		if err != nil {
			return err
		}
	}

	for _, r := range c.rows {
		err = w.Write(r)
		if err != nil {
			return err
		}
	}

	if c.footer != nil {
		err = w.Write(c.footer)
		if err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}
