package validation

import (
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/recfile"
	"os"
	"path"
	"slices"
	"strings"
)

// We want a dialogue script checker
// It should go through all dialogue files
// 1. Looking for missing nodes

type DialogueReport struct {
	DialogueFile       string
	NodeNamesPresent   map[string]bool
	NodeNamesMentioned map[string]bool
}

func (dr DialogueReport) MissingNodes() []string {
	var missing []string
	for name := range dr.NodeNamesMentioned {
		if _, ok := dr.NodeNamesPresent[name]; !ok {
			missing = append(missing, name)
		}
	}

	slices.SortStableFunc(missing, strings.Compare)

	return missing
}

func (dr DialogueReport) UnreferencedNodes() []string {
	var unreferenced []string
	for name := range dr.NodeNamesPresent {
		if _, ok := dr.NodeNamesMentioned[name]; !ok {
			unreferenced = append(unreferenced, name)
		}
	}

	slices.SortStableFunc(unreferenced, strings.Compare)

	return unreferenced
}

type DialogueChecker struct {
	rootDir string
	reports []DialogueReport
}

func ValidateDialogue(rootDir string) {
	checker := NewDialogueChecker(rootDir)
	reports := checker.CreateReports()
	for _, report := range reports {
		missing := report.MissingNodes()
		unreferenced := report.UnreferencedNodes()
		if len(missing) > 0 || len(unreferenced) > 0 {
			println("\nDialogue file: ", report.DialogueFile)
			if len(missing) > 0 {
				println("Missing nodes: ")
				for _, name := range missing {
					println("  ", name)
				}
			}
			if len(unreferenced) > 0 {
				println("Unreferenced nodes: ")
				for _, name := range unreferenced {
					println("  ", name)
				}
			}
		}
	}
}

func NewDialogueChecker(rootDir string) *DialogueChecker {
	return &DialogueChecker{rootDir: rootDir}
}

func (dc *DialogueChecker) CreateReports() []DialogueReport {
	dir, err := os.ReadDir(dc.rootDir)
	if err != nil {
		return nil
	}
	var reports []DialogueReport
	for _, entry := range dir {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), "_") {
			continue
		}
		filePath := path.Join(dc.rootDir, entry.Name())
		records := recfile.ReadMulti(fxtools.MustOpen(filePath))
		openingBranchRecords := records["OpeningBranch"]
		nodeRecords := records["Nodes"]
		nodeNamesPresent := make(map[string]bool)
		nodeNamesMentioned := make(map[string]bool)

		nodeNamesMentioned = appendFromOpeningBranch(nodeNamesMentioned, openingBranchRecords)
		nodeNamesPresent, nodeNamesMentioned = appendFromNodes(nodeNamesPresent, nodeNamesMentioned, nodeRecords)

		reports = append(reports, DialogueReport{
			DialogueFile:       entry.Name(),
			NodeNamesPresent:   nodeNamesPresent,
			NodeNamesMentioned: nodeNamesMentioned,
		})
	}
	return reports
}

func appendFromOpeningBranch(mentioned map[string]bool, records []recfile.Record) map[string]bool {
	isFieldWithNode := func(fieldName string) bool {
		return strings.ToLower(fieldName) == "goto"
	}
	for _, record := range records {
		for _, field := range record {
			if isFieldWithNode(field.Name) {
				mentioned[field.Value] = true
			}
		}
	}
	return mentioned
}

func appendFromNodes(present map[string]bool, mentioned map[string]bool, records []recfile.Record) (map[string]bool, map[string]bool) {
	for _, record := range records {
		for _, field := range record {
			switch strings.ToLower(field.Name) {
			case "name":
				present[field.Value] = true
			case "o_fail":
				fallthrough
			case "o_succ":
				fallthrough
			case "o_goto":
				mentioned[field.Value] = true
			case "effect": // GotoNode('NodeName')
				if fxtools.LooksLikeAFunction(field.Value) {
					name, args := fxtools.GetNameAndArgs(field.Value)
					if strings.ToLower(name) == "gotonode" {
						mentioned[args.Get(0)] = true
					}
				}
			}
		}
	}
	return present, mentioned
}
