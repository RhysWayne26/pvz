#!/bin/bash

MIGRATIONS_DIR="$(pwd)/migrations"
OUTPUT_FILE="$(pwd)/merged_migrations/all_migrations.sql"

if [ ! -d "$MIGRATIONS_DIR" ]; then
  echo "Directory $MIGRATIONS_DIR does not exist!"
  exit 1
fi

# Create directory if it doesn't exist
mkdir -p "$(dirname "$OUTPUT_FILE")"

# Clear the output file
# shellcheck disable=SC2188
> "$OUTPUT_FILE"

echo "Merging migrations from $MIGRATIONS_DIR into $OUTPUT_FILE"

# Simply concatenate all migration files in order
for file in $(find "$MIGRATIONS_DIR" -type f -name "*.sql" | sort); do
  echo "Adding: $file"
  # shellcheck disable=SC2129
  echo "-- Source: $file" >> "$OUTPUT_FILE"
  cat "$file" >> "$OUTPUT_FILE"
  echo "" >> "$OUTPUT_FILE"
  echo "" >> "$OUTPUT_FILE"
done

echo "All migrations have been merged into $OUTPUT_FILE"