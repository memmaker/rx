# Apply the updated parser to the provided REC file
def parse_rec_file_with_record_types(file_path):
    """
    Parse a .rec file with support for `%rec` types and return its contents as a dictionary.

    Args:
        file_path (str): Path to the .rec file.

    Returns:
        dict: A dictionary where keys are `%rec` types, and values are lists of records of that type.
    """
    records_by_type = {}
    current_record = {}
    current_type = None

    def unescape_value(value):
        """Unescape multiline values."""
        import re
        return re.sub(r'\\n\+\s', '\n', value)

    with open(file_path, 'r') as file:
        for line in file:
            line = line.rstrip()  # Retain trailing spaces for multiline detection

            # Skip comments
            if line.strip().startswith("#"):
                continue

            if line.startswith("%rec"):
                # New record type detected
                if current_record and current_type:
                    # Save the current record to its type
                    if current_type not in records_by_type:
                        records_by_type[current_type] = []
                    records_by_type[current_type].append(current_record)
                    current_record = {}

                # Set the new type
                current_type = line.split(":", 1)[1].strip()
            elif ": " in line:
                # End current record if it's not empty
                if current_record and not line.startswith(" "):
                    if current_type not in records_by_type:
                        records_by_type[current_type] = []
                    records_by_type[current_type].append(current_record)
                    current_record = {}

                # New key-value pair
                key, value = line.split(": ", 1)
                current_record[key] = unescape_value(value)
            elif line.startswith("\\n+"):
                # Continuation of the previous value
                if current_record:
                    last_key = list(current_record.keys())[-1]
                    current_record[last_key] += "\n" + unescape_value(line.strip())
            elif not line.strip() and current_record:
                # Blank line indicates the end of a record
                if current_type not in records_by_type:
                    records_by_type[current_type] = []
                records_by_type[current_type].append(current_record)
                current_record = {}

        # Append the last record if present
        if current_record and current_type:
            if current_type not in records_by_type:
                records_by_type[current_type] = []
            records_by_type[current_type].append(current_record)

    return records_by_type

# Parse the file and display the results
parsed_records_with_types = parse_rec_file_with_record_types(rec_file_path)

# Display the parsed data for review
parsed_records_with_types
