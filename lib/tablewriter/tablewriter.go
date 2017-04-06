package tablewriter

type TableWriter interface {
	SetHeader(keys []string)
	SetFooter(keys []string)
	Append(row []string)
	Render() error
}
