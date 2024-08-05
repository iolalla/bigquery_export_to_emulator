package main

import (
	"cloud.google.com/go/bigquery"
	"context"
	"errors"
	"flag"
	"fmt"
	"google.golang.org/api/iterator"
	"os"
)

func main() {
	project := flag.String("project", "YOURPROJECT", "What's the project name")
	//dataset := flag.String("dataset", "YOURDATASET", "What's the dataset name")
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
	for {
		datazet, err := itz.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			panic(fmt.Sprintf("Failed to open connection to Bigquery: %v", err))
		}
		dataset := datazet.DatasetID
		result += fmt.Sprintf("      - id: %s\n", dataset)
		result += fmt.Sprintf("        tables:\n")
		ds := client.DatasetInProject(*project, dataset)
		it := ds.Tables(ctx)
		for {
			t, err := it.Next()
			if errors.Is(err, iterator.Done) {
				break
			}
			result += fmt.Sprintf("          - id: %s\n", t.TableID)
			// If you want to get the first rows of the table you can remove those three lines and remove Offset
			// TODO: Add a control based on limit == 0, so in this case we get all the rows
			meta, err := client.Dataset(dataset).Table(t.TableID).Metadata(ctx)
			if err != nil {
				panic(fmt.Sprintf("Failed to open connection to Bigquery: %v", err))
			}
			// With this we get a
			var offset uint64 = 0
			if meta.NumRows > *limit {
				offset = meta.NumRows - *limit
			}
			query := fmt.Sprint("SELECT * FROM `", *project, ".", dataset, ".", t.TableID, "` ", "LIMIT ", *limit, " OFFSET ", offset)
			text, err1 := GenerateTableData(query, client, ctx)
			if err1 != nil {
				fmt.Printf("error with query : %s\n", query)
				fmt.Printf("error: %v\n", err1)
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
			}
			contador = 1
			result += fmt.Sprintf("            data:\n")
		}
		i := 0
		for _, row := range rows {
			if i == 0 {
				result += fmt.Sprintf("              - %s: %v\n", namez[i], row)
			} else {
				result += fmt.Sprintf("                %s: \"%v\"\n", namez[i], row)
			}
			i++
		}
		rowsRead++
	}
	return result, nil
}
