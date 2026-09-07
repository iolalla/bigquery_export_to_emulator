package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/bigquery"
	"go.yaml.in/yaml/v3"
	"google.golang.org/api/iterator"
)

func main() {
	project := flag.String("project", "YOURPROJECT", "What's the project name")
	dataset := flag.String("dataset", "", "Optional dataset name; exports all datasets when omitted")
	outFile := flag.String("outfile", "out.yaml", "File to store the data")
	limit := flag.Uint64("limit", 250, "The limit  limit ")
	flag.Parse()
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, *project)
	if err != nil {
		panic(fmt.Sprintf("Failed to open connection to Bigquery: %v", err))
	}
	defer client.Close()
	result := fmt.Sprintf("projects:\n")
	result += fmt.Sprintf("  - id: %s\n", *project)
	result += fmt.Sprintf("    datasets:\n")
	itz := client.Datasets(ctx)
	datasets, err := selectDatasetIDs(*dataset, itz.Next)
	if err != nil {
		panic(fmt.Sprintf("Failed to open connection to Bigquery: %v", err))
	}
	for _, datasetID := range datasets {
		result += fmt.Sprintf("      - id: %s\n", datasetID)
		result += fmt.Sprintf("        tables:\n")
		ds := client.DatasetInProject(*project, datasetID)
		it := ds.Tables(ctx)
		for {
			t, err := it.Next()
			if errors.Is(err, iterator.Done) {
				break
			}
			result += fmt.Sprintf("          - id: %s\n", t.TableID)
			// If you want to get the first rows of the table you can remove those three lines and remove Offset
			// TODO: Add a control based on limit == 0, so in this case we get all the rows
			meta, err := client.Dataset(datasetID).Table(t.TableID).Metadata(ctx)
			if err != nil {
				panic(fmt.Sprintf("Failed to open connection to Bigquery: %v", err))
			}
			// With this we get a
			var offset uint64 = 0
			if meta.NumRows > *limit {
				offset = meta.NumRows - *limit
			}
			query := fmt.Sprint("SELECT * FROM `", *project, ".", datasetID, ".", t.TableID, "` ", "LIMIT ", *limit, " OFFSET ", offset)
			text, err1 := GenerateTableData(query, client, ctx)
			if err1 != nil {
				panic(fmt.Sprintf("Failed to export query %s: %v", query, err1))
			}
			result += text
		}
	}
	f, fe := os.Create(*outFile)
	if fe != nil {
		panic(fe)
	}
	_, err3 := f.WriteString(result)
	if err3 != nil {
		panic(err3)
	}
	err1 := f.Sync()
	if err1 != nil {
		panic(err1)
	}
	// print(result)
}

func selectDatasetIDs(requested string, next func() (*bigquery.Dataset, error)) ([]string, error) {
	if requested != "" {
		return []string{requested}, nil
	}
	var datasets []string
	for {
		dataset, err := next()
		if errors.Is(err, iterator.Done) {
			return datasets, nil
		}
		if err != nil {
			return nil, err
		}
		datasets = append(datasets, dataset.DatasetID)
	}
}

func trimUTCDateTimeSuffix(value interface{}) interface{} {
	switch value := value.(type) {
	case string:
		return strings.TrimSuffix(value, " +0000 UTC")
	case []interface{}:
		for i := range value {
			value[i] = trimUTCDateTimeSuffix(value[i])
		}
	case map[string]interface{}:
		for key := range value {
			value[key] = trimUTCDateTimeSuffix(value[key])
		}
	}
	return value
}

func renderJSONField(name string, value bigquery.Value, isFirst bool) (string, error) {
	parsed := value
	if value != nil {
		if text, ok := value.(string); ok {
			if err := json.Unmarshal([]byte(text), &parsed); err != nil {
				return "", fmt.Errorf("failed to parse JSON value for column %s: %w", name, err)
			}
		}
	}
	parsed = trimUTCDateTimeSuffix(parsed)
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(map[string]bigquery.Value{name: parsed}); err != nil {
		return "", fmt.Errorf("failed to render JSON value for column %s: %w", name, err)
	}
	if err := encoder.Close(); err != nil {
		return "", fmt.Errorf("failed to render JSON value for column %s: %w", name, err)
	}
	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	var result strings.Builder
	for i, line := range lines {
		if i == 0 && isFirst {
			result.WriteString("              - " + line + "\n")
		} else {
			result.WriteString("                " + line + "\n")
		}
	}
	return result.String(), nil
}

func GenerateTableData(query string, client *bigquery.Client, ctx context.Context) (string, error) {
	result := fmt.Sprintf("            columns:\n")
	q := client.Query(query)
	// Execute the query.
	it1, err1 := q.Read(ctx)
	if err1 != nil {
		// TODO: Handle error.
		return "", err1
	}
	contador := 0
	var namez []string
	var typez []string
	var rowsRead int
	for {
		var rows []bigquery.Value
		err := it1.Next(&rows)
		if errors.Is(err, iterator.Done) {
			fmt.Printf("ITERATION COMPLETE. Rows read %v \n", rowsRead)
			break
		}
		if err != nil {
			fmt.Printf("error!: %v\n", err)
			return result, err
		}
		if contador == 0 {
			for _, fs := range it1.Schema {
				result += fmt.Sprintf("              - name: %s\n", fs.Name)
				result += fmt.Sprintf("                type: %s\n", string(fs.Type))
				namez = append(namez, fs.Name)
				typez = append(typez, string(fs.Type))
			}
			contador = 1
			result += fmt.Sprintf("            data:\n")
		}
		for i, row := range rows {
			if typez[i] == string(bigquery.JSONFieldType) {
				yamlText, err := renderJSONField(namez[i], row, i == 0)
				if err != nil {
					return result, err
				}
				result += yamlText
			} else if i == 0 {
				result += fmt.Sprintf("              - %s: %v\n", namez[i], row)
			} else {
				result += fmt.Sprintf("                %s: \"%v\"\n", namez[i], row)
			}
		}
		rowsRead++
	}
	return result, nil
}
