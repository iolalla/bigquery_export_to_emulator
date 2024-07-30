# Bigquery Export to Emulator in YAML format

This project aims to help to have an export of your bigquery dataset compatible with the project [bigquery-emulator](https://github.com/goccy/bigquery-emulator).

The main objetive is to be able to have all the tables of a dataset exported to a yaml file, so you can test locally.

This is implemented in Go, as the emulator.

# Env Vars
```
export GOOGLE_CLOUD_PROJECT=game-bolsa
export GOOGLE_APPLICATION_CREDENTIALS=/home/YOURUSER/YOURCREDENTIALS.json
```
# Compile
```
make all
```
# Execute
```
be_exp --project=YOURPROJECT --dataset=YOURDATASET --outfile=YOUROUTFILE.yaml
```
# License

MIT