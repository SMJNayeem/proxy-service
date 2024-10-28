#!/bin/bash

# Output file
OUTPUT_FILE="project_dump.txt"

# Clear the output file if it already exists
> "$OUTPUT_FILE"

# Function to recursively write the content of all files
write_files_to_output() {
  local DIR="$1"

  # Loop through all files and directories
  for FILE in "$DIR"/*; do
    if [ -d "$FILE" ]; then
      # If it's a directory, recurse into it
      write_files_to_output "$FILE"
    elif [ -f "$FILE" ]; then
      # Write the file header and content to the output file
      echo "=========================================" >> "$OUTPUT_FILE"
      echo "File: $FILE" >> "$OUTPUT_FILE"
      echo "=========================================" >> "$OUTPUT_FILE"
      cat "$FILE" >> "$OUTPUT_FILE"
      echo -e "\n\n" >> "$OUTPUT_FILE"
    fi
  done
}

# Start writing from the current directory
write_files_to_output "."

echo "Project dump created successfully in $OUTPUT_FILE"
