package recfile

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"image/color"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// https://en.wikipedia.org/wiki/Recfiles
// https://www.gnu.org/software/recutils/manual/The-Rec-Format.html#The-Rec-Format
type Field struct {
	Name  string
	Value string
}
type Record []Field

func (f Field) String() string {
	return fmt.Sprintf("%s: %s", f.Name, f.Value)
}

func (f Field) IsEmpty() bool {
	return f.Name == "" && f.Value == ""
}

// UnEscapedValue returns the value with the sequence "\n+ " replaced with newlines.
func (f Field) UnEscapedValue() string {
	return regexp.MustCompile(`\n\+\s`).ReplaceAllString(f.Value, "\n")
}

// EscapedValue returns the value with newlines escaped as "\n+ ".
// This is useful for writing .rec files from a constructed Record.
func (f Field) EscapedValue() string {
	return strings.ReplaceAll(f.Value, "\n", "\n+ ")
}

func (f Field) AsInt() int {
	if value, err := strconv.Atoi(f.Value); err == nil {
		return value
	}
	return 0
}

func (f Field) AsRune() rune {
	return []rune(f.Value)[0]
}
func (f Field) AsInt32() int32 {
	value, _ := strconv.ParseInt(f.Value, 10, 32)
	return int32(value)
}

func (f Field) AsBool() bool {
	return f.Value == "true"
}

func (f Field) AsFloat() float64 {
	if value, err := strconv.ParseFloat(f.Value, 64); err == nil {
		return value
	}
	return 0.0
}

func (f Field) AsList(sep string) []Field {
	return fieldMap(stringMap(strings.Split(f.Value, sep), strings.TrimSpace))
}

func (f Field) AsRGB(sep string) color.RGBA {
	parts := colorMap(stringMap(strings.Split(f.Value, sep), strings.TrimSpace))
	return color.RGBA{
		R: parts[0],
		G: parts[1],
		B: parts[2],
		A: 255,
	}
}

func fieldMap(i []string) []Field {
	result := make([]Field, len(i))
	for j, value := range i {
		result[j] = Field{Value: value}
	}
	return result
}

func colorMap(inputValues []string) [3]uint8 {
	result := [3]uint8{}
	for j, value := range inputValues {
		val, _ := strconv.Atoi(value)
		result[j] = uint8(val)
	}
	return result
}

func stringMap(fields []string, mapFunc func(string) string) []string {
	result := make([]string, len(fields))
	for i, field := range fields {
		result[i] = mapFunc(field)
	}
	return result
}

type DataMap map[string]string

func (d DataMap) GetInt(key string) (int, error) {
	if value, ok := d[key]; ok {
		return strconv.Atoi(value)
	}
	return -1, fmt.Errorf("no value")
}

func (d DataMap) GetBoolOrFalse(key string) bool {
	if value, ok := d[key]; ok {
		return value == "true"
	}
	return false
}
func (d DataMap) GetBoolOrTrue(key string) bool {
	if value, ok := d[key]; ok {
		return value == "true"
	}
	return true
}
func (d DataMap) GetStringOrDefault(key string, defaultValue string) string {
	if value, ok := d[key]; ok {
		return value
	}
	return defaultValue
}

func (d DataMap) GetFloatOrDefault(key string, defaultValue float64) float64 {
	if value, ok := d[key]; ok {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func (d DataMap) GetIntOrDefault(key string, defaultValue int) int {
	if value, ok := d[key]; ok {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func (r Record) ToMap(listSeperator string) DataMap {
	m := make(map[string]string, len(r))
	for _, field := range r {
		if _, exists := m[field.Name]; exists {
			// append
			m[field.Name] += listSeperator + field.Value
		} else {
			m[field.Name] = field.Value
		}
	}
	return m
}

func (r Record) ToValueList() []string {
	result := make([]string, len(r))
	for i, field := range r {
		result[i] = field.Value
	}
	return result
}

func (r Record) ToFixedSizeValueList(fieldNamesInOrder []string) []string {
	var result []string
	asMap := r.ToMap("|")

	for _, fieldName := range fieldNamesInOrder {
		if value, ok := asMap[fieldName]; ok {
			result = append(result, value)
		} else {
			result = append(result, "")
		}
	}

	return result
}

type RecReader struct {
	records           map[string][]Record
	currentRecord     []Field
	currentField      Field
	linePart          string
	currentRecordType string
}

func NewReader() *RecReader {
	return &RecReader{
		records:           make(map[string][]Record),
		currentRecord:     make([]Field, 0),
		currentField:      Field{},
		linePart:          "",
		currentRecordType: "default",
	}
}
func (r *RecReader) ReadLine(line string) {
	//scanner := bufio.NewScanner(file)
	fieldNamePattern := regexp.MustCompile(`^([a-zA-Z%][a-zA-Z0-9_]*):[\t ]?`)
	plusPrefixPattern := regexp.MustCompile(`^\+\s?`)
	// eg. %rec: Article
	recordTypeRegex := regexp.MustCompile(`^%rec:\s*([a-zA-Z][a-zA-Z0-9_]*)`)
	line = r.linePart + line
	r.linePart = ""

	if strings.HasPrefix(line, "#") {
		return
	}
	if plusPrefixPattern.MatchString(line) {
		line = plusPrefixPattern.ReplaceAllString(line, "\n")
	}
	if matches := recordTypeRegex.FindStringSubmatch(line); matches != nil {
		r.tryCommitCurrentField()
		r.tryCommitCurrentRecord()
		r.currentRecord = make([]Field, 0)
		r.currentField = Field{}
		r.currentRecordType = matches[1]
		r.records[r.currentRecordType] = make([]Record, 0)
		return
	}

	if strings.HasSuffix(line, "\\") {
		r.linePart = line[:len(line)-1]
		return
	}

	if fieldNamePattern.MatchString(line) {
		r.tryCommitCurrentField()
		matches := fieldNamePattern.FindStringSubmatch(line)
		r.currentField = Field{
			Name:  matches[1],
			Value: strings.Trim(line[len(matches[0]):], " \t"),
		}
	} else if line == "" {
		r.tryCommitCurrentField()
		r.currentField = Field{}
		r.tryCommitCurrentRecord()
		r.currentRecord = make([]Field, 0)
	} else {
		r.currentField.Value += strings.Trim(line, " \t")
	}
}

func (r *RecReader) tryCommitCurrentRecord() {
	if len(r.currentRecord) > 0 {
		r.records[r.currentRecordType] = append(r.records[r.currentRecordType], r.currentRecord)
	}
}

func (r *RecReader) tryCommitCurrentField() {
	if !r.currentField.IsEmpty() {
		r.currentRecord = append(r.currentRecord, r.currentField)
	}
}

func (r *RecReader) End() map[string][]Record {
	r.tryCommitCurrentField()
	r.currentField = Field{}
	r.tryCommitCurrentRecord()
	return r.records
}

func (r *RecReader) ReadLines(data []string) map[string][]Record {
	for _, line := range data {
		r.ReadLine(line)
	}
	return r.End()
}
func defaultOnly(records map[string][]Record) []Record {
	return records["default"]
}
func Read(file io.Reader) []Record {
	return defaultOnly(ReadMulti(file))
}
func ReadMulti(input io.Reader) map[string][]Record {
	scanner := bufio.NewScanner(input)
	reader := NewReader()
	for scanner.Scan() {
		reader.ReadLine(scanner.Text())
	}
	return reader.End()
}
func Write(file io.StringWriter, records []Record) error {
	return WriteMulti(file, map[string][]Record{"default": records})
}
func WriteCSV(output io.Writer, fieldNames []string, records []Record) {
	csvWriter := csv.NewWriter(output)
	csvWriter.Write(fieldNames)
	for _, record := range records {
		asValues := record.ToFixedSizeValueList(fieldNames)
		csvWriter.Write(asValues)
	}
	csvWriter.Flush()
}
func WriteMulti(file io.StringWriter, recordsInCategories map[string][]Record) error {
	sanitizeFieldname := func(s string) string {
		saneFieldname := strings.ReplaceAll(s, " ", "_")
		if saneFieldname != s {
			println(fmt.Sprintf("WARNING - Sanitizing fieldname: '%s' -> '%s'", s, saneFieldname))
		}
		return saneFieldname
	}
	for recordCategory, records := range recordsInCategories {
		_, catErr := file.WriteString(fmt.Sprintf("%%rec: %s\n\n", recordCategory))
		if catErr != nil {
			return catErr
		}
		for _, record := range records {
			for _, field := range record {
				_, err := file.WriteString(sanitizeFieldname(field.Name) + ": " + field.EscapedValue() + "\n")
				if err != nil {
					return err
				}
			}
			_, err := file.WriteString("\n")
			if err != nil {
				return err
			}
		}
	}
	return nil
}
