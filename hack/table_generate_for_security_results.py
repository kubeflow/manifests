import json
from prettytable import PrettyTable

# Path to your JSON file
json_file = '../docs/image_lists/summary_of_severity_counts_for_WG/severity_summary_in_json_format.json'

# Load JSON content from the file
with open(json_file, 'r') as file:
    data = json.load(file)

# Define a mapping for working group names
groupnames = {
    "Automl": "AutoML",
    "Pipelines": "Pipelines",
    "Workbenches":"Workbenches(Notebooks)",
    "Serving": "Kserve",
    "Manifests":"Manifests",
    "Training": "Training",
    "Model-registry":"Model Registry",
    "total": "All Images",
}

# Create PrettyTable
table = PrettyTable()
table.field_names = ["Working Group", "Images", "Critical CVE", "High CVE", "Medium CVE", "Low CVE"]

# Populate the table with data
for group_name in groupnames:
    if group_name in data:  # Check if group_name exists in data
        value = data[group_name]
        table.add_row([groupnames[group_name], value["images"], value["CRITICAL"], value["HIGH"], value["MEDIUM"], value["LOW"]])


# Print the table
print(table)

output_folder='../docs/image_lists/summary_of_severity_counts_for_WG/'

# Write the table output to a file in the specified folder
output_file = output_folder + 'summary_of_severity_counts_for_WGs_in_table.txt'
with open(output_file, 'w') as f:
    f.write(str(table))

print("Output saved to:", output_file)