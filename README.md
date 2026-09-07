# Bigquery Export to Emulator in YAML format

This project aims to help to have an export of your bigquery dataset compatible with the project [bigquery-emulator](https://github.com/goccy/bigquery-emulator).

The main objetive is to be able to have all the tables of a dataset exported to a yaml file, so you can test locally.

This is implemented in Go, as the emulator.

# Env Vars
```
export GOOGLE_CLOUD_PROJECT=YOUR_PROYECT
export GOOGLE_APPLICATION_CREDENTIALS=/home/YOURUSER/YOURCREDENTIALS.json
```
# Compile
```
make all
```
# Execute
Export one dataset:
```
be_exp --project=YOURPROJECT --dataset=YOURDATASET --outfile=YOUROUTFILE.yaml
```

Omit `--dataset` to export every dataset in the project:
```
be_exp --project=YOURPROJECT --outfile=YOUROUTFILE.yaml
```

BigQuery JSON columns are decoded and written as structured YAML so the output can be loaded with `bigquery-emulator --data-from-yaml=YOUROUTFILE.yaml`.

# License

MIT
