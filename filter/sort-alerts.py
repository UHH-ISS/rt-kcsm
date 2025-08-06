import pandas as pd
import argparse

parser = argparse.ArgumentParser("log sorter by time")
parser.add_argument("input", help="input file of alerts", type=str)
parser.add_argument("field", help="field to be sorted (timestamp)", type=str)
parser.add_argument("output", help="output file", type=str)
args = parser.parse_args()

df = pd.read_json(args.input, lines=True)

field = args.field

print("Done reading")
df.sort_values(by=field, inplace=True) # timestamp
print("Done sorting")

def convert(row):
    row[field] = str(row[field]).replace(" ", "T")

    splits = row[field].split(":")

    row[field] = ":".join(splits[:-1]) + splits[-1]

    return row

df = df.apply(convert, axis=1)

df.to_json(args.output, lines=True, orient="records")